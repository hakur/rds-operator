package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hakur/rds-operator/pkg/types"
	"github.com/sirupsen/logrus"
)

type SemiSync struct {
	// DataSrouces mysql instance data sources
	DataSrouces    []*DSN
	DoubleMasterHA bool
}

func (t *SemiSync) StartCluster(ctx context.Context) (err error) {
	select {
	case <-ctx.Done():
		return types.ErrCtxTimeout
	default:
		var masters []*DSN
		var master *DSN
		var maxMasterServerID = 1

		if t.DoubleMasterHA {
			maxMasterServerID = 2
			for k, dsn := range t.DataSrouces { // generate masters
				if k+1 <= maxMasterServerID {
					masters = append(masters, dsn)
				}
			}
		}

		for k, dsn := range t.DataSrouces { // ordinary start mysql semi sync nodes
			if k+1 <= maxMasterServerID {
				err = t.bootCluster(ctx, dsn, masters)
				if err == nil {
					master = dsn
				}
			} else {
				err = t.joinMaster(ctx, dsn, master, 1)
			}

			if err != nil && !errors.Is(err, types.ErrMysqlSemiSyncIsAlreadyRunning) {
				return err
			}
		}
	}

	return nil
}

// bootCluster set mysql instance as cluster bootstrap node
func (t *SemiSync) bootCluster(ctx context.Context, dsn *DSN, masters []*DSN) (err error) {
	dbConn, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/mysql",
		dsn.Username,
		dsn.Password,
		dsn.Host,
		dsn.Port,
	))
	if err != nil {
		return types.ErrMyqlConnectFaild
	}

	defer dbConn.Close()

	if on, err := t.checkMasterON(ctx, dbConn); !on {
		logrus.WithFields(map[string]interface{}{"err": err.Error(), "host": dsn.Host}).Debugf("mysql check semi sync master is not running")
		_, err = dbConn.ExecContext(ctx, "SET GLOBAL rpl_semi_sync_master_enabled=ON")
		if err != nil {
			return fmt.Errorf("%w, enable [host=%s] master module err -> %s", types.ErrMysqlStartSemiSyncMasterFailed, dsn.Host, err.Error())
		}

		_, err = dbConn.ExecContext(ctx, "SET GLOBAL super_read_only=0")
		if err != nil {
			return fmt.Errorf("%w, enable [host=%s] super read only = %d err -> %s", types.ErrMysqlStartSemiSyncMasterFailed, dsn.Host, 0, err.Error())
		}
	}

	if t.DoubleMasterHA {
		var anotherMaster *DSN
		for _, v := range masters {
			if v.Host != dsn.Host {
				anotherMaster = v
				break
			}
		}

		if anotherMaster == nil {
			return fmt.Errorf("mysql [host=%s] semi sync DoubleMasterHA enabled, but another master not found", dsn.Host)
		} else {
			err = t.joinMaster(ctx, dsn, anotherMaster, 0)
		}
	}

	return
}

// joinMaster make mysql instance join master node as slave
func (t *SemiSync) joinMaster(ctx context.Context, dsn *DSN, master *DSN, superReadOnly int) (err error) {
	dbConn, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/mysql",
		dsn.Username,
		dsn.Password,
		dsn.Host,
		dsn.Port,
	))
	if err != nil {
		return types.ErrMyqlConnectFaild
	}

	defer dbConn.Close()

	if on, err := t.checkSlaveON(ctx, dbConn); !on {
		logrus.WithFields(map[string]interface{}{"err": err.Error(), "host": dsn.Host}).Debugf("mysql check semi sync slave process is running failed")

		if _, err = dbConn.ExecContext(ctx, "SET GLOBAL rpl_semi_sync_slave_enabled=ON"); err != nil {
			return fmt.Errorf("%w, enable [host=%s] slave module err -> %s", types.ErrMysqlStartSemiSyncSlaveFailed, dsn.Host, err.Error())
		}

		if _, err = dbConn.ExecContext(ctx, "SET GLOBAL super_read_only="+strconv.Itoa(superReadOnly)); err != nil {
			return fmt.Errorf("%w, set [host=%s] super read only = %d err -> %s", types.ErrMysqlStartSemiSyncSlaveFailed, dsn.Host, superReadOnly, err.Error())
		}
	}

	myMaster, err := t.getMyMaster(ctx, dbConn)
	if err != nil {
		logrus.WithFields(map[string]interface{}{"err": err.Error(), "host": dsn.Host}).Debugf(types.ErrMysqlFindMasterFromSalveFailed.Error())
	}

	if myMaster != master.Host || myMaster == "" {
		if _, err = dbConn.ExecContext(ctx, "STOP SLAVE"); err != nil {
			return fmt.Errorf("%w,stop slave [host=%s] err -> %s", types.ErrMysqlStartSemiSyncSlaveFailed, dsn.Host, err.Error())
		}

		if _, err = dbConn.ExecContext(ctx, "CHANGE MASTER TO MASTER_HOST='"+master.Host+"',MASTER_USER='"+dsn.Username+"' ,MASTER_PASSWORD='"+dsn.Password+"',MASTER_AUTO_POSITION=1"); err != nil {
			return fmt.Errorf("%w, change master [host=%s] err -> %s", types.ErrMysqlStartSemiSyncSlaveFailed, dsn.Host, err.Error())
		}

		if _, err = dbConn.ExecContext(ctx, "START SLAVE"); err != nil {
			return fmt.Errorf("%w, start slave [host=%s] err -> %s", types.ErrMysqlStartSemiSyncSlaveFailed, dsn.Host, err.Error())
		}
	}

	return
}

func (t *SemiSync) FindMaster(ctx context.Context) (masters []*DSN, err error) {
	select {
	case <-ctx.Done():
		return nil, types.ErrCtxTimeout
	default:
		for _, dsn := range t.DataSrouces {
			dbConn, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/mysql",
				dsn.Username,
				dsn.Password,
				dsn.Host,
				dsn.Port,
			))
			if err != nil {
				logrus.WithFields(map[string]interface{}{"err": err.Error(), "host": dsn.Host}).Debugf(types.ErrMyqlConnectFaild.Error())
				continue
			}
			defer dbConn.Close()

			if masterON, err := t.checkMasterON(ctx, dbConn); err == nil && masterON {
				masters = append(masters, dsn)
			}
		}
	}

	if len(masters) < 1 {
		return masters, types.ErrMasterNoutFound
	}

	return masters, nil
}

func (t *SemiSync) getMyMaster(ctx context.Context, dbConn *sql.DB) (masterHost string, err error) {
	result, err := dbConn.QueryContext(ctx, "SHOW SLAVE STATUS")
	if err != nil {
		return masterHost, fmt.Errorf("mysql query result scan global status master_host failed, err -> %s", err.Error())
	}

	// code source http://noops.me/?p=1128
	cols, _ := result.Columns()
	buff := make([]interface{}, len(cols))
	data := make([]string, len(cols))
	for i, _ := range buff {
		buff[i] = &data[i]
	}

	for result.Next() {
		result.Scan(buff...)
	}

	for k, v := range cols {
		if v == "Master_Host" {
			masterHost = data[k]
		}
	}

	return masterHost, nil
}

func (t *SemiSync) checkMasterON(ctx context.Context, dbConn *sql.DB) (on bool, err error) {
	result, err := dbConn.QueryContext(ctx, "show variables like 'rpl_semi_sync_master_enabled'")
	if err != nil {
		return
	}

	var masterON string
	if result.Next() {
		result.Scan(&masterON, &masterON)
	} else {
		err = errors.New("mysql query result scan rpl_semi_sync_master_enabled status failed")
	}

	if masterON == "ON" {
		on = true
	} else {
		err = errors.New("mysql rpl_semi_sync_master_enabled is OFF")
	}

	return
}

func (t *SemiSync) checkSlaveON(ctx context.Context, dbConn *sql.DB) (on bool, err error) {
	result, err := dbConn.QueryContext(ctx, "show variables like 'rpl_semi_sync_slave_enabled'")
	if err != nil {
		return
	}

	var slaveON string
	if result.Next() {
		result.Scan(&slaveON, &slaveON)
	} else {
		err = errors.New("mysql query result scan rpl_semi_sync_slave_enabled status failed")
	}

	if slaveON == "ON" {
		on = true
	} else {
		err = errors.New("mysql rpl_semi_sync_slave_enabled is OFF")
	}

	return
}

func (t *SemiSync) HealthyMembers(ctx context.Context) (members []*DSN) {
	select {
	case <-ctx.Done():
		return members
	default:
		var wg sync.WaitGroup

		for _, dsn := range t.DataSrouces {
			wg.Add(1)
			go func(dsn *DSN) {
				defer wg.Done()
				dbConn, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/mysql",
					dsn.Username,
					dsn.Password,
					dsn.Host,
					dsn.Port,
				))
				if err != nil {
					logrus.WithFields(map[string]interface{}{"err": err.Error(), "host": dsn.Host}).Debugf(types.ErrMyqlConnectFaild.Error())
					return
				}
				defer dbConn.Close()

				if on, err := t.checkSlaveON(ctx, dbConn); on {
					members = append(members, dsn)
				} else {
					logrus.WithFields(map[string]interface{}{"err": err.Error(), "host": dsn.Host}).Debugf("mysql check semi sync is running failed")
				}
			}(dsn)
		}

		wg.Wait()
	}

	return
}
