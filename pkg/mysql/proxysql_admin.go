package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hakur/rds-operator/pkg/types"
	hutil "github.com/hakur/util"
)

func NewProxySQLAdmin(dsn DSN) (t *ProxySQLAdmin, err error) {
	t = new(ProxySQLAdmin)
	if db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/main",
		dsn.Username,
		dsn.Password,
		dsn.Host,
		dsn.Port,
	)); err != nil {
		t = nil
	} else {
		t.Conn = db
	}

	return
}

type ProxySQLAdmin struct {
	Conn *sql.DB
}

func (t *ProxySQLAdmin) Close() error {
	return t.Conn.Close()
}

func (t *ProxySQLAdmin) GetProxySQLServers(ctx context.Context) (data []*TableProxySQLServers, err error) {
	result, err := t.Conn.QueryContext(ctx, "SELECT hostname,port,weight,comment FROM proxysql_servers")
	if err != nil {
		return data, err
	}
	for result.Next() {
		ps := new(TableProxySQLServers)
		err = result.Scan(&ps.Hostname, &ps.Port, &ps.Weight, &ps.Comment)
		if err != nil {
			return data, err
		}
		data = append(data, ps)
	}
	return data, err
}

func (t *ProxySQLAdmin) Begin(ctx context.Context) (err error) {
	_, err = t.Conn.ExecContext(ctx, "BEGIN")
	return err
}
func (t *ProxySQLAdmin) Rollback(ctx context.Context) (err error) {
	_, err = t.Conn.ExecContext(ctx, "ROLLBACK")
	return err
}
func (t *ProxySQLAdmin) Commit(ctx context.Context) (err error) {
	_, err = t.Conn.ExecContext(ctx, "COMMIT")
	return err
}

func (t *ProxySQLAdmin) AddProxySQLServers(ctx context.Context, servers []*TableProxySQLServers) (err error) {
	select {
	case <-ctx.Done():
		return types.ErrCtxTimeout
	default:
		for _, server := range servers {
			hutil.DefaultValue(server)
			// check hostname exists
			// INSERT INTO proxysql_servers(hostname,port,weight,comment) values ('yuxing-proxysql-0','6032','1','') ON DUPLICATE KEY UPDATE is not support with proxysql 2.3.2
			hostnameCount := 0
			countResukt, err := t.Conn.QueryContext(ctx, "SELECT COUNT(hostname) FROM proxysql_servers WHERE hostname='"+server.Hostname+"'")
			if err != nil {
				return fmt.Errorf("count hostname=%s from proxysql db error -> %s", server.Hostname, err.Error())
			} else {
				if countResukt.Next() {
					countResukt.Scan(&hostnameCount)
				}
			}

			if hostnameCount > 0 {
				_, err = t.Conn.ExecContext(ctx, fmt.Sprintf("UPDATE proxysql_servers set port=%d,weight='%d',comment='%s' WHERE hostname='%s'",
					server.Port,
					server.Weight,
					server.Comment,
					server.Hostname,
				))

				if err != nil {
					return fmt.Errorf("update hostname=%s to proxysql db error -> %s", server.Hostname, err.Error())
				}
			} else {
				_, err = t.Conn.ExecContext(ctx, fmt.Sprintf("INSERT INTO proxysql_servers(hostname,port,weight,comment) values ('%s','%d','%d','%s')",
					server.Hostname,
					server.Port,
					server.Weight,
					server.Comment,
				))

				if err != nil {
					return fmt.Errorf("insert hostname=%s to proxysql db error -> %s", server.Hostname, err.Error())
				}
			}
		}
	}
	return err
}

func (t *ProxySQLAdmin) RemoveProxySQLServer(ctx context.Context, hostname string) (err error) {
	_, err = t.Conn.ExecContext(ctx, "DELETE FROM proxysql_servers where hostname='"+hostname+"'")
	return err
}

func (t *ProxySQLAdmin) GetMysqlServers(ctx context.Context) (data []*TableMysqlServers, err error) {
	result, err := t.Conn.QueryContext(ctx, fmt.Sprintf("SELECT %s FROM mysql_servers", strings.Join([]string{
		"hostgroup_id", "hostname", "port", "gtid_port", "status", "weight", "compression",
		"max_connections", "max_replication_lag", "use_ssl", "max_latency_ms", "comment",
	}, ",")))
	if err != nil {
		return data, err
	}

	for result.Next() {
		ms := new(TableMysqlServers)
		err = result.Scan(
			&ms.HostGroupID,
			&ms.Hostname,
			&ms.Port,
			&ms.GTIDPort,
			&ms.Status,
			&ms.Weight,
			&ms.Compression,
			&ms.MaxConnections,
			&ms.MaxReplicationLag,
			&ms.UseSSL,
			&ms.MaxLatencyMS,
			&ms.Comment,
		)
		if err != nil {
			return data, err
		}
		data = append(data, ms)
	}
	return data, err
}

func (t *ProxySQLAdmin) AddMysqlServers(ctx context.Context, servers []*TableMysqlServers) (err error) {
	select {
	case <-ctx.Done():
		return types.ErrCtxTimeout
	default:
		for _, server := range servers {
			hutil.DefaultValue(server)
			// check hostname exists
			hostnameCount := 0
			countResukt, err := t.Conn.QueryContext(ctx, "SELECT COUNT(hostname) FROM mysql_servers WHERE hostname='"+server.Hostname+"'")
			if err != nil {
				return fmt.Errorf("count hostname=%s from proxysql db error -> %s", server.Hostname, err.Error())
			} else {
				if countResukt.Next() {
					countResukt.Scan(&hostnameCount)
				}
			}

			if hostnameCount > 0 {
				_, err = t.Conn.ExecContext(ctx, fmt.Sprintf("UPDATE mysql_servers set hostgroup_id=%d,hostname='%s',port=%d,gtid_port=%d,status='%s',weight=%d,compression=%d,max_connections=%d,max_replication_lag=%d,use_ssl=%d,max_latency_ms=%d,comment='%s' WHERE hostname='%s'",
					server.HostGroupID,
					server.Hostname,
					server.Port,
					server.GTIDPort,
					server.Status,
					server.Weight,
					hutil.BoolToInt(server.Compression),
					server.MaxConnections,
					server.MaxReplicationLag,
					hutil.BoolToInt(server.UseSSL),
					server.MaxLatencyMS,
					server.Comment,
					server.Hostname,
				))

				if err != nil {
					return fmt.Errorf("update hostname=%s to proxysql db error -> %s", server.Hostname, err.Error())
				}
			} else {
				_, err = t.Conn.ExecContext(ctx, fmt.Sprintf("INSERT INTO mysql_servers(hostgroup_id,hostname,port,gtid_port,status,weight,compression,max_connections,max_replication_lag,use_ssl,max_latency_ms,comment) values (%d,'%s',%d,%d,'%s',%d,%d,%d,%d,%d,%d,'%s')",
					server.HostGroupID,
					server.Hostname,
					server.Port,
					server.GTIDPort,
					server.Status,
					server.Weight,
					hutil.BoolToInt(server.Compression),
					server.MaxConnections,
					server.MaxReplicationLag,
					hutil.BoolToInt(server.UseSSL),
					server.MaxLatencyMS,
					server.Comment,
				))

				if err != nil {
					return fmt.Errorf("insert hostname=%s to proxysql db error -> %s", server.Hostname, err.Error())
				}
			}
		}
	}
	return err
}

func (t *ProxySQLAdmin) RemoveMysqlServer(ctx context.Context, hostname string) (err error) {
	_, err = t.Conn.ExecContext(ctx, "DELETE FROM mysql_servers where hostname='"+hostname+"'")
	return err
}

func (t *ProxySQLAdmin) GetMysqlUsers(ctx context.Context) (data []*TableMysqlUsers, err error) {
	result, err := t.Conn.QueryContext(ctx, fmt.Sprintf("SELECT %s FROM mysql_users", strings.Join([]string{
		"username", "password", "active", "use_ssl", "default_hostgroup", "default_schema", "schema_locked",
		"transaction_persistent", "fast_forward", "backend", "frontend", "max_connections", "attributes", "comment",
	}, ",")))
	if err != nil {
		return data, err
	}

	for result.Next() {
		user := new(TableMysqlUsers)
		err = result.Scan(
			&user.Username,
			&user.Password,
			&user.Active,
			&user.UseSSL,
			&user.DefaultHostgroup,
			&user.DefaultSchema,
			&user.SchemaLocked,
			&user.TransactionPersistent,
			&user.FastForward,
			&user.Backend,
			&user.Frontend,
			&user.MaxConnections,
			&user.Attributes,
			&user.Comment,
		)
		if err != nil {
			return data, err
		}
		data = append(data, user)
	}
	return data, err
}

func (t *ProxySQLAdmin) AddMysqlUsers(ctx context.Context, users []*TableMysqlUsers) (err error) {
	select {
	case <-ctx.Done():
		return types.ErrCtxTimeout
	default:
		for _, user := range users {
			hutil.DefaultValue(user)
			// check user exists
			usernameCount := 0
			countResukt, err := t.Conn.QueryContext(ctx, "SELECT COUNT(hostname) FROM mysql_servers WHERE hostname='"+user.Username+"' AND frontend="+hutil.BoolToStrNumber(user.Frontend))
			if err != nil {
				return fmt.Errorf("count username=%s,frontend=%s from proxysql db error -> %s", user.Username, hutil.BoolToStrNumber(user.Frontend), err.Error())
			} else {
				if countResukt.Next() {
					countResukt.Scan(&usernameCount)
				}
			}

			// Username              string         //username
			// Password              string         //password
			// Active                bool           // active
			// UseSSL                bool           //use_ssl
			// DefaultHostgroup      int            // default_hostgroup
			// DefaultSchema         sql.NullString //default_schema
			// SchemaLocked          bool           //schema_locked
			// TransactionPersistent bool           //transaction_persistent
			// FastForward           bool           //fast_forward
			// Backend               bool           //backend
			// Frontend              bool           //frontend
			// MaxConnections        int            //max_connections
			// Attributes            string         //attributes
			// Comment               string         //comment
			if usernameCount > 0 {
				_, err = t.Conn.ExecContext(ctx, fmt.Sprintf("UPDATE mysql_servers set username='%s',password='%s',active=%d,use_ssl=%d,default_hostgroup=%d,default_schema='%s',schema_locked=%d,transaction_persistent=%d,fast_forward=%d,backend=%d,frontend=%d,max_connections=%d,attributes='%s',comment=%s WHERE username='%s AND frontend=%d'",
					user.Username,
					user.Password,
					hutil.BoolToInt(user.Active),
					hutil.BoolToInt(user.UseSSL),
					user.DefaultHostgroup,
					user.DefaultSchema.String,
					hutil.BoolToInt(user.SchemaLocked),
					hutil.BoolToInt(user.TransactionPersistent),
					hutil.BoolToInt(user.FastForward),
					hutil.BoolToInt(user.Backend),
					hutil.BoolToInt(user.Frontend),
					user.MaxConnections,
					user.Attributes,
					user.Comment,
					user.Username,
					hutil.BoolToInt(user.Frontend),
				))

				if err != nil {
					return fmt.Errorf("update username=%s,frontend=%s to proxysql db error -> %s", user.Username, hutil.BoolToStrNumber(user.Frontend), err.Error())
				}
			} else {
				_, err = t.Conn.ExecContext(ctx, fmt.Sprintf("INSERT INTO mysql_servers(username,password,active,use_ssl,default_hostgroup,default_schema,schema_locked,transaction_persistent,fast_forward,backend,frontend,max_connections,attributes,comment) values ('%s','%s',%d,%d,%d,'%s',%d,%d,%d,%d,%d,%d,'%s','%s')",
					user.Username,
					user.Password,
					hutil.BoolToInt(user.Active),
					hutil.BoolToInt(user.UseSSL),
					user.DefaultHostgroup,
					user.DefaultSchema.String,
					hutil.BoolToInt(user.SchemaLocked),
					hutil.BoolToInt(user.TransactionPersistent),
					hutil.BoolToInt(user.FastForward),
					hutil.BoolToInt(user.Backend),
					hutil.BoolToInt(user.Frontend),
					user.MaxConnections,
					user.Attributes,
					user.Comment,
				))

				if err != nil {
					return fmt.Errorf("insert username=%s,frontend=%s to proxysql db error -> %s", user.Username, hutil.BoolToStrNumber(user.Frontend), err.Error())
				}
			}
		}
	}
	return err
}

func (t *ProxySQLAdmin) RemoveMysqlUser(ctx context.Context, username string, frontend bool) (err error) {
	_, err = t.Conn.ExecContext(ctx, "DELETE FROM mysql_users where hostname='"+username+"' and frontend="+hutil.BoolToStrNumber(frontend))
	return err
}
