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

type MGRSP struct {
	// DataSrouces mysql instance data sources
	DataSrouces []*DSN
}

func (t *MGRSP) StartCluster(ctx context.Context) (err error) {
	select {
	case <-ctx.Done():
		return types.ErrCtxTimeout
	default:
		var masters []*DSN
		masters, _ = t.FindMaster(ctx)

		for k, dsn := range t.DataSrouces { // ordinary start mysql group replication
			if k == 0 && len(masters) < 1 {
				err = t.bootCluster(ctx, dsn)
			} else {
				err = t.joinMaster(ctx, dsn)
			}

			if err != nil && !errors.Is(err, types.ErrMysqlMGRIsAlreadyRunning) {
				return err
			}
		}
	}

	return nil
}

// bootCluster set mysql instance as cluster bootstrap node
func (t *MGRSP) bootCluster(ctx context.Context, dsn *DSN) (err error) {
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

	if on, err := t.checkMGRIsRunning(ctx, dbConn); on {
		return types.ErrMysqlMGRIsAlreadyRunning
	} else {
		logrus.WithFields(map[string]interface{}{"err": err.Error(), "host": dsn.Host}).Debugf("mysql check mgr is running failed")
	}

	_, err = dbConn.ExecContext(ctx, "SET GLOBAL group_replication_bootstrap_group=ON")
	if err != nil {
		return fmt.Errorf("%w, err -> %s", types.ErrMysqlStartMGRSPClusterFailed, err.Error())
	}

	_, err = dbConn.ExecContext(ctx, "START group_replication")
	if err != nil {
		return fmt.Errorf("%w, err -> %s", types.ErrMysqlStartMGRSPClusterFailed, err.Error())
	}

	_, err = dbConn.ExecContext(ctx, "SET GLOBAL group_replication_bootstrap_group=OFF")
	if err != nil {
		return fmt.Errorf("%w, err -> %s", types.ErrMysqlStartMGRSPClusterFailed, err.Error())
	}

	return
}

// joinMaster make mysql instance join master node as slave
func (t *MGRSP) joinMaster(ctx context.Context, dsn *DSN) (err error) {
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

	if on, err := t.checkMGRIsRunning(ctx, dbConn); on {
		return types.ErrMysqlMGRIsAlreadyRunning
	} else {
		logrus.WithFields(map[string]interface{}{"err": err.Error(), "host": dsn.Host}).Debugf("mysql check mgr is running failed")
	}

	_, err = dbConn.ExecContext(ctx, "CHANGE MASTER TO MASTER_USER='"+dsn.Username+"' ,MASTER_PASSWORD='"+dsn.Password+"'  FOR CHANNEL 'group_replication_recovery'")
	if err != nil {
		return fmt.Errorf("%w, err -> %s", types.ErrMysqlStartMGRSPClusterFailed, err.Error())
	}

	_, err = dbConn.ExecContext(ctx, "START group_replication")
	if err != nil {
		return fmt.Errorf("%w, err -> %s", types.ErrMysqlStartMGRSPClusterFailed, err.Error())
	}

	return
}

func (t *MGRSP) FindMaster(ctx context.Context) (masters []*DSN, err error) {
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

			row, err := dbConn.QueryContext(ctx, "SHOW VARIABLES LIKE 'server_uuid'")
			if err != nil {
				logrus.WithFields(map[string]interface{}{"err": err.Error(), "host": dsn.Host}).Debugf("mysql query local server uuid failed")
				continue
			}

			var myServerUUID string
			if row.Next() { // only one row data, for loop is not required
				row.Scan(&myServerUUID, &myServerUUID) // two colume, need twice scan, second scan is our target data value
			}

			if myServerUUID == "" {
				logrus.WithFields(map[string]interface{}{"err": err.Error(), "host": dsn.Host}).Debugf("mysql result scan server uuid value failed")
				continue
			}

			row, err = dbConn.QueryContext(ctx, "SELECT VARIABLE_VALUE FROM performance_schema.global_status WHERE VARIABLE_NAME= 'group_replication_primary_member' AND VARIABLE_VALUE=?", myServerUUID)
			if err != nil {
				logrus.WithFields(map[string]interface{}{"err": err.Error(), "host": dsn.Host}).Debugf("mysql query master server uuid failed")
				continue
			}

			var masterServerUUID string
			if row.Next() { // only one row data, for loop is not required
				row.Scan(&masterServerUUID)
			}

			if masterServerUUID != "" && masterServerUUID == myServerUUID {
				masters = append(masters, dsn)
			}
		}

	}

	if len(masters) < 1 {
		return nil, types.ErrMasterNoutFound
	}
	return masters, nil
}

func (t *MGRSP) checkMGRIsRunning(ctx context.Context, dbConn *sql.DB) (on bool, err error) {
	result, err := dbConn.QueryContext(ctx, "select service_state from performance_schema.replication_applier_status where channel_name='group_replication_applier'")
	if err != nil {
		return
	}

	var mgrOn string
	if result.Next() {
		result.Scan(&mgrOn)
	} else {
		err = errors.New("mysql query result scan group_replication_applier service_state failed")
	}

	if mgrOn == "ON" {
		on = true
	} else {
		err = errors.New("mysql group_replication_applier service is not running")
	}

	return
}

func (t *MGRSP) HealthyMembers(ctx context.Context) (members []*DSN) {
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

				if on, err := t.checkMGRIsRunning(ctx, dbConn); on {
					members = append(members, dsn)
				} else {
					logrus.WithFields(map[string]interface{}{"err": err.Error(), "host": dsn.Host}).Debugf("mysql check mgr is running failed")
				}
			}(dsn)
		}

		wg.Wait()
	}

	return
}
