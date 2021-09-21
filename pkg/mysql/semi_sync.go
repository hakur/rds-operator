package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hakur/rds-operator/pkg/types"
	"github.com/sirupsen/logrus"
)

type SemiSync struct {
	// DataSrouces mysql instance data sources
	DataSrouces []*DSN
}

func (t *SemiSync) StartCluster(ctx context.Context) (err error) {

	select {
	case <-ctx.Done():
		return types.ErrCtxTimeout
	default:

		var master *DSN
		master, _ = t.FindMaster(ctx)

		for k, dsn := range t.DataSrouces { // ordinary start mysql semi sync nodes
			if k == 0 && master == nil {
				err = t.bootCluster(ctx, dsn)
				master = dsn
			} else {
				err = t.joinMaster(ctx, dsn)
			}

			if err != nil && !errors.Is(err, types.ErrMysqlSemiSyncIsAlreadyRunning) {
				return err
			}
		}
	}

	return nil
}

// bootCluster set mysql instance as cluster bootstrap node
func (t *SemiSync) bootCluster(ctx context.Context, dsn *DSN) (err error) {
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

	if on, err := t.checkSlaveON(ctx, dbConn); on {
		return types.ErrMysqlSemiSyncIsAlreadyRunning
	} else {
		logrus.WithFields(map[string]interface{}{"err": err.Error(), "host": dsn.Host}).Debugf("mysql check semi sync is running failed")
	}

	_, err = dbConn.ExecContext(ctx, "SET GLOBAL group_replication_bootstrap_group=ON")
	if err != nil {
		return fmt.Errorf("%w, err -> %s", types.ErrMysqlStartSemiSyncClusterFailed, err.Error())
	}

	_, err = dbConn.ExecContext(ctx, "START group_replication")
	if err != nil {
		return fmt.Errorf("%w, err -> %s", types.ErrMysqlStartSemiSyncClusterFailed, err.Error())
	}

	_, err = dbConn.ExecContext(ctx, "SET GLOBAL group_replication_bootstrap_group=OFF")
	if err != nil {
		return fmt.Errorf("%w, err -> %s", types.ErrMysqlStartSemiSyncClusterFailed, err.Error())
	}

	return
}

// joinMaster make mysql instance join master node as slave
func (t *SemiSync) joinMaster(ctx context.Context, dsn *DSN) (err error) {
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

	if on, err := t.checkSlaveON(ctx, dbConn); on {
		return types.ErrMysqlSemiSyncIsAlreadyRunning
	} else {
		logrus.WithFields(map[string]interface{}{"err": err.Error(), "host": dsn.Host}).Debugf("mysql check semi sync is running failed")
	}

	_, err = dbConn.ExecContext(ctx, "CHANGE MASTER TO MASTER_USER='"+dsn.Username+"' ,MASTER_PASSWORD='"+dsn.Password+"'  FOR CHANNEL 'group_replication_recovery'")
	if err != nil {
		return fmt.Errorf("%w, err -> %s", types.ErrMysqlStartSemiSyncClusterFailed, err.Error())
	}

	_, err = dbConn.ExecContext(ctx, "START group_replication")
	if err != nil {
		return fmt.Errorf("%w, err -> %s", types.ErrMysqlStartSemiSyncClusterFailed, err.Error())
	}

	return
}

func (t *SemiSync) FindMaster(ctx context.Context) (masterDSN *DSN, err error) {
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

			row, err := dbConn.QueryContext(ctx, "show  variables like 'rpl_semi_sync_master_enabled'")
			if err != nil {
				logrus.WithFields(map[string]interface{}{"err": err.Error(), "host": dsn.Host}).Debugf("mysql query local server uuid failed")
				continue
			}

			var masterON string
			if row.Next() { // two colume but second colume value is needed
				row.Scan(&masterON, &masterON)
			}

			if masterON == "ON" {
				return dsn, nil
			}
		}

	}

	return nil, types.ErrMasterNoutFound
}

func (t *SemiSync) checkSlaveON(ctx context.Context, dbConn *sql.DB) (on bool, err error) {
	result, err := dbConn.QueryContext(ctx, "show variables like 'rpl_semi_sync_master_enabled'")
	if err != nil {
		return
	}

	var slaveON string
	if result.Next() {
		result.Scan(&slaveON)
	} else {
		err = errors.New("mysql query result scan rpl_semi_sync_master_enabled status failed")
	}

	if slaveON == "ON" {
		on = true
	} else {
		err = errors.New("mysql rpl_semi_sync_master_enabled is OFF")
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
