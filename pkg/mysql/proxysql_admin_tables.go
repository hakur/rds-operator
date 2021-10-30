package mysql

import "database/sql"

// TableProxySQLServers proxysql table proxysql_servers
// CREATE TABLE runtime_proxysql_servers (
//     hostname VARCHAR NOT NULL,
//     port INT NOT NULL DEFAULT 6032,
//     weight INT CHECK (weight >= 0) NOT NULL DEFAULT 0,
//     comment VARCHAR NOT NULL DEFAULT '',
//     PRIMARY KEY (hostname, port) )
type TableProxySQLServers struct {
	Hostname string // hostname
	Port     int    `default:"6032"` // port
	Weight   int    // weight
	Comment  string // comment
}

// TableMysqlServers proxysql table mysql_servers define
// CREATE TABLE runtime_mysql_servers (
//     hostgroup_id INT CHECK (hostgroup_id>=0) NOT NULL DEFAULT 0,
//     hostname VARCHAR NOT NULL,
//     port INT CHECK (port >= 0 AND port <= 65535) NOT NULL DEFAULT 3306,
//     gtid_port INT CHECK ((gtid_port <> port OR gtid_port=0) AND gtid_port >= 0 AND gtid_port <= 65535) NOT NULL DEFAULT 0,
//     status VARCHAR CHECK (UPPER(status) IN ('ONLINE','SHUNNED','OFFLINE_SOFT', 'OFFLINE_HARD')) NOT NULL DEFAULT 'ONLINE',
//     weight INT CHECK (weight >= 0 AND weight <=10000000) NOT NULL DEFAULT 1,
//     compression INT CHECK (compression IN(0,1)) NOT NULL DEFAULT 0,
//     max_connections INT CHECK (max_connections >=0) NOT NULL DEFAULT 1000,
//     max_replication_lag INT CHECK (max_replication_lag >= 0 AND max_replication_lag <= 126144000) NOT NULL DEFAULT 0,
//     use_ssl INT CHECK (use_ssl IN(0,1)) NOT NULL DEFAULT 0,
//     max_latency_ms INT UNSIGNED CHECK (max_latency_ms>=0) NOT NULL DEFAULT 0,
//     comment VARCHAR NOT NULL DEFAULT '',
//     PRIMARY KEY (hostgroup_id, hostname, port) )
type TableMysqlServers struct {
	HostGroupID       int    //hostgroup_id
	Hostname          string // hostname
	Port              int    `default:"3306"` //port
	GTIDPort          int    // gtid_port
	Status            string `default:"ONLINE"` //status
	Weight            int    `default:"1"`      // weight
	Compression       bool   //compression
	MaxConnections    int    `default:"200"` //max_connections
	MaxReplicationLag int    //max_replication_lag
	UseSSL            bool   //use_ssl
	MaxLatencyMS      int    // max_latency_ms
	Comment           string //comment
}

// TableMysqlUsers proxysql table mysql_users
// CREATE TABLE mysql_users (
//     username VARCHAR NOT NULL,
//     password VARCHAR,
//     active INT CHECK (active IN (0,1)) NOT NULL DEFAULT 1,
//     use_ssl INT CHECK (use_ssl IN (0,1)) NOT NULL DEFAULT 0,
//     default_hostgroup INT NOT NULL DEFAULT 0,
//     default_schema VARCHAR,
//     schema_locked INT CHECK (schema_locked IN (0,1)) NOT NULL DEFAULT 0,
//     transaction_persistent INT CHECK (transaction_persistent IN (0,1)) NOT NULL DEFAULT 1,
//     fast_forward INT CHECK (fast_forward IN (0,1)) NOT NULL DEFAULT 0,
//     backend INT CHECK (backend IN (0,1)) NOT NULL DEFAULT 1,
//     frontend INT CHECK (frontend IN (0,1)) NOT NULL DEFAULT 1,
//     max_connections INT CHECK (max_connections >=0) NOT NULL DEFAULT 10000,
//     attributes VARCHAR CHECK (JSON_VALID(attributes) OR attributes = '') NOT NULL DEFAULT '',
//     comment VARCHAR NOT NULL DEFAULT '',
//     PRIMARY KEY (username, backend),
//     UNIQUE (username, frontend))
type TableMysqlUsers struct {
	Username              string         //username
	Password              string         //password
	Active                bool           `default:"true"` // active
	UseSSL                bool           //use_ssl
	DefaultHostgroup      int            // default_hostgroup
	DefaultSchema         sql.NullString //default_schema
	SchemaLocked          bool           //schema_locked
	TransactionPersistent bool           `default:"true"` //transaction_persistent
	FastForward           bool           //fast_forward
	Backend               bool           `default:"true"` //backend
	Frontend              bool           `default:"true"` //frontend
	MaxConnections        int            `default:"200"`  //max_connections
	Attributes            string         //attributes
	Comment               string         //comment
}
