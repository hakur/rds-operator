package main

import (
	"os"
	"strconv"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/pkg/mysql"
	hutil "github.com/hakur/util"
	"gopkg.in/alecthomas/kingpin.v2"
)

type ProxySQLConfigCommand struct {
	// Dump output generated proxysql config on stdout
	Dump                 bool
	GlobalVar            *ProxySQLCommandFlagValues
	AdminCredentials     string
	ClusterUsername      string
	ClusterPassword      string
	MaxWriteNodes        int
	MysqlMaxConns        int
	QueryTimeout         int
	MysqlVersion         string
	MysqlMonitorUsername string
	MysqlMonitorPassword string
}

func (t *ProxySQLConfigCommand) Register(cmd *kingpin.CmdClause) {
	cmd.Action(t.Action)
	cmd.Flag("admin-credentials", "proxysql AdminCredentials config").Default(hutil.Base64Decode(hutil.EnvOrDefault("ADMIN_CREDENTIALS", "admin:admin"))).StringVar(&t.AdminCredentials)
	cmd.Flag("max-write-nodes", "proxysql max_writers, max write nodes").Default(hutil.EnvOrDefault("PROXYSQL_MAX_WRITE_NODES", "1")).IntVar(&t.MaxWriteNodes)
	cmd.Flag("cluster-username", "proxysql cluster monitor username").Default(hutil.EnvOrDefault("PROXYSQL_CLUSTER_USERNAME", "radmin")).StringVar(&t.ClusterUsername)
	cmd.Flag("cluster-password", "proxysql cluster monitor password").Default(hutil.Base64Decode(hutil.EnvOrDefault("PROXYSQL_CLUSTER_PASSWORD", "radmin"))).StringVar(&t.ClusterPassword)
	cmd.Flag("mysql-max-conns", "max conn per mysql instance").Default(hutil.EnvOrDefault("MYSQL_MAX_CONNS", "250")).IntVar(&t.MysqlMaxConns)
	cmd.Flag("query-timeout", "sql query exec on mysql instance timeout milliseconds").Default(hutil.EnvOrDefault("QUERY_TIMEOUT", "60000")).IntVar(&t.MysqlMaxConns)
	cmd.Flag("mysql-version", "mysql server version").Default(hutil.EnvOrDefault("MYSQL_VERSION", "5.7.34")).StringVar(&t.MysqlVersion)
	cmd.Flag("mysql-monitor-username", "username of user on mysql instance and have sys database select privilege").Default(hutil.EnvOrDefault("MYSQL_MONITOR_USERNAME", "monitor")).StringVar(&t.MysqlMonitorUsername)
	cmd.Flag("mysql-monitor-password", "password of user on mysql instance and have sys database select privilege").Default(hutil.Base64Decode(hutil.EnvOrDefault("MYSQL_MONITOR_PASSWORD", "monitor"))).StringVar(&t.MysqlMonitorPassword)
	cmd.Flag("dump", "output generated mysqld config on stdout, to enable in format --dump without any argument").Default(hutil.EnvOrDefault("PROXYSQL_CFG_DUMP", "true")).BoolVar(&t.Dump)
}

func (t *ProxySQLConfigCommand) Action(ctx *kingpin.ParseContext) (err error) {
	var cfg = mysql.ProxySQLConfWriter{}
	// https://www.cnblogs.com/cqdba/p/14809248.html
	// mysql -uadmin -preplication_password -P6032 -h127.0.0.1
	// LOAD MYSQL SERVERS TO RUNTIME;
	cfg.AdminVariables = map[string]string{
		"admin_credentials":                    t.AdminCredentials,
		"mysql_ifaces":                         "0.0.0.0:6032",
		"admin-cluster_username":               t.ClusterUsername,
		"admin-cluster_password":               t.ClusterPassword,
		"admin-cluster_check_interval_ms":      "200",
		"admin-cluster_check_status_frequency": "100",
		// "cluster_mysql_query_rules_save_to_disk":      "true",
		// "cluster_mysql_servers_save_to_disk":          "true",
		// "cluster_mysql_users_save_to_disk":            "true",
		// "cluster_proxysql_servers_save_to_disk":       "true",
		// "cluster_mysql_query_rules_diffs_before_sync": "3",
		// "cluster_mysql_servers_diffs_before_sync":     "3",
		// "cluster_mysql_users_diffs_before_sync":       "3",
		// "cluster_admin_variables_diffs_before_sync":   "3",
		// "cluster_proxysql_servers_diffs_before_sync":  "3",
	}

	cfg.MysqlVariables = map[string]string{
		"threads":                    "8",
		"max_connections":            strconv.Itoa(t.MysqlMaxConns),
		"default_query_delay":        "0",
		"default_query_timeout":      "0",
		"have_compress":              "true",
		"poll_timeout":               "4000",
		"interfaces":                 "0.0.0.0:3306",
		"default_schema":             "mysql",
		"stacksize":                  "1048576",
		"server_version":             t.MysqlVersion,
		"connect_timeout_server":     "2000",
		"monitor_username":           t.MysqlMonitorUsername,
		"monitor_password":           t.MysqlMonitorPassword,
		"monitor_history":            "600000",
		"monitor_connect_interval":   "10000",
		"monitor_connect_timeout":    "1000",
		"monitor_ping_interval":      "10000",
		"monitor_read_only_interval": "1500",
		"monitor_read_only_timeout":  "500",
		"ping_interval_server_msec":  "120000",
		"ping_timeout_server":        "500",
		"commands_stats":             "true",
		"sessions_sort":              "true",
		"connect_retries_on_failure": "10",
	}

	// cfg.MysqlUsers = append(cfg.MysqlUsers, map[string]string{
	// 	"username":          "root",
	// 	"password":          "123456",
	// 	"default_hostgroup": "10",
	// 	"max_connections":   "200",
	// 	"default_schema":    "mysql",
	// 	"active":            "1",
	// })

	// cfg.MysqlServers = append(cfg.MysqlServers, map[string]string{
	// 	"addresses":       "yuxing-mysql-0",
	// 	"port":            "3306",
	// 	"hostgroup":       "10",
	// 	"max_connections": "200",
	// })
	// cfg.MysqlServers = append(cfg.MysqlServers, map[string]string{
	// 	"addresses":       "yuxing-mysql-1",
	// 	"port":            "3306",
	// 	"hostgroup":       "10",
	// 	"max_connections": "200",
	// })
	// cfg.MysqlServers = append(cfg.MysqlServers, map[string]string{
	// 	"addresses":       "yuxing-mysql-2",
	// 	"port":            "3306",
	// 	"hostgroup":       "10",
	// 	"max_connections": "200",
	// })

	// cfg.ProxySQLServers = append(cfg.ProxySQLServers, map[string]string{
	// 	"hostname": "yuxing-proxysql-0",
	// 	"port":     "6032",
	// 	"weight":   "1",
	// })

	cfg.Datadir = "/data/proxysql"

	switch t.GlobalVar.Mode {
	case string(rdsv1alpha1.ModeMGRSP):
		t.mgrspConfig(&cfg)
	case string(rdsv1alpha1.ModeMGRMP):
		t.mgrmpConfig(&cfg)
	case string(rdsv1alpha1.ModeSemiSync):
		t.semiSyncConfig(&cfg)
	}

	content := cfg.String()
	if t.Dump {
		if t.Dump {
			os.Stdout.WriteString("generated config :\n" + content)
		}
	}
	return os.WriteFile("/etc/proxysql.cnf.d/proxysql.cnf", []byte(content), 0755)
}

func (t *ProxySQLConfigCommand) mgrspConfig(cfg *mysql.ProxySQLConfWriter) {
	cfg.MysqlGroupReplicationHostgroups = append(cfg.MysqlGroupReplicationHostgroups, map[string]string{
		"writer_hostgroup":        "10",
		"reader_hostgroup":        "20",
		"backup_writer_hostgroup": "11",
		"offline_hostgroup":       "0",
		"active":                  "1",
		"max_writers":             "1",
		"writer_is_also_reader":   "0",
		"max_transactions_behind": "0",
	})
}

func (t *ProxySQLConfigCommand) mgrmpConfig(cfg *mysql.ProxySQLConfWriter) {
	cfg.MysqlGroupReplicationHostgroups = append(cfg.MysqlGroupReplicationHostgroups, map[string]string{
		"writer_hostgroup":        "10",
		"reader_hostgroup":        "20",
		"backup_writer_hostgroup": "11",
		"offline_hostgroup":       "0",
		"active":                  "1",
		"max_writers":             strconv.Itoa(t.MaxWriteNodes),
		"writer_is_also_reader":   "1",
		"max_transactions_behind": "0",
	})
}

func (t *ProxySQLConfigCommand) semiSyncConfig(cfg *mysql.ProxySQLConfWriter) {
	cfg.MysqlReplicationHostgroups = append(cfg.MysqlReplicationHostgroups, map[string]string{
		"writer_hostgroup": "30",
		"reader_hostgroup": "40",
	})
}
