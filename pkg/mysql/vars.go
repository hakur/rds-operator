package mysql

var (
	MysqlClientVars = map[string]string{
		"socket": "/tmp/mysql.sock",
	}
	// MysqldDefaultVars mysqld 默认配置，不包含各种模式的配置
	MysqldVars = map[string]string{
		"log_timestamps":                   "SYSTEM",
		"slow_query_log":                   "on",
		"socket":                           "/tmp/mysql.sock",
		"skip-name-resolve":                "on",
		"user":                             "mysql",
		"symbolic-links":                   "0",
		"pid-file":                         "/var/run/mysqld/mysqld.pid",
		"max_connections":                  "300",
		"default-storage-engine":           "INNODB",
		"character-set-server":             "utf8",
		"collation-server":                 "utf8_general_ci",
		"gtid_mode":                        "on",
		"enforce_gtid_consistency":         "on",
		"master_info_repository":           "TABLE",
		"relay_log_info_repository":        "TABLE",
		"binlog_checksum":                  "NONE",
		"log_slave_updates":                "ON",
		"log_bin":                          "binlog",
		"binlog_format":                    "ROW",
		"transaction_isolation":            "READ-COMMITTED",
		"transaction_write_set_extraction": "XXHASH64",
		"datadir":                          "/var/lib/mysql",
	}

	// MultiParimaryVars mysql 多主模式需要的配置项
	MultiParimaryVars = map[string]string{
		"": "",
	}

	// SiglePrimaryVars mysql 单主模式需要的配置项
	SiglePrimaryVars = map[string]string{
		"plugin_load_add":                                          "group_replication.so",
		"loose-group_replication_group_name":                       "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
		"loose-group_replication_bootstrap_group":                  "off",
		"loose-group_replication_single_primary_mode":              "on",
		"loose-group_replication_enforce_update_everywhere_checks": "off",
		"loose-group_replication_start_on_boot":                    "off",
	}

	GaleraClusterVars = map[string]string{
		// 集群同步方式为同步
	}
)

// BackupPolicy 备份策略类型
type BackupPolicy string

// ClusterMode 集群模式类型
type ClusterMode string

const (
	// BackupDataDir 备份方式为复制目录
	BackupDataDir BackupPolicy = "BackupDataDir"
	// ModeMGRMP 集群模式为 mysql group replication multi primary
	ModeMGRMP ClusterMode = "MGRMP"
	// ModeMGRSP 集群模式为 mysql group replication single primary
	ModeMGRSP ClusterMode = "MGRSP"
	// ModeSemiSync 集群模式为 mysql semi sync
	ModeSemiSync ClusterMode = "SemiSync"
	// ModeAsync 集群模式为 mysql 异步模式
	ModeAsync ClusterMode = "Async"
	// ModeGaleraSync 集群模式为 galera cluster for mysql 集群
	ModeGaleraCluster ClusterMode = "GaleraCluster"
)
