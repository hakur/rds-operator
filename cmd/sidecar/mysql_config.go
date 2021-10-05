package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/pkg/mysql"
	"github.com/hakur/rds-operator/util"
	"gopkg.in/alecthomas/kingpin.v2"
)

var BasicConf = `
[client]
socket=/tmp/mysql.sock
[mysqld]
skip-name-resolve
socket=/tmp/mysql.sock
secure-file-priv=/var/lib/mysql-files
user=mysql
symbolic-links=0
pid-file=/var/run/mysqld/mysqld.pid
default-storage-engine=INNODB
character-set-server=utf8
collation-server=utf8_general_ci
transaction_isolation=READ-COMMITTED

gtid_mode=ON
enforce-gtid-consistency=true

sync_binlog=1
log_bin=bin.log
binlog_format=row
binlog_gtid_simple_recovery=1

relay_log=relay.log
relay_log_recovery=1
relay_log_info_repository=TABLE
master_info_repository=TABLE
binlog_checksum=NONE

slave_skip_errors=ddl_exist_errors
`

type MysqlConfigCommand struct {
	// Dump output generated mysqld config on stdout
	Dump      bool
	GlobalVar *MysqlGlobalFlagValues
	// ConfigFile mysql config file path
	ConfigFile string
	// Whitelist mysql server whitelist
	Whitelist []string
}

func (t *MysqlConfigCommand) Register(cmd *kingpin.CmdClause) {
	cmd.Action(t.Action)
	cmd.Flag("config-file", "mysqld config file path").Default(util.EnvOrDefault("MYSQL_CFG_EXTRA_DIR", "/etc/my.cnf.d") + "/my.cnf").StringVar(&t.ConfigFile)
	cmd.Flag("white-list", "mysql server white list").Default(util.EnvOrDefault("MYSQL_CFG_WHITE_LIST", "10.0.0.0/8,192.0.0.0/8")).StringsVar(&t.Whitelist)
	cmd.Flag("dump", "output generated mysqld config on stdout, to enable in format --dump without any argument").Default(util.EnvOrDefault("MYSQL_CFG_DUMP", "false")).BoolVar(&t.Dump)
}

func (t *MysqlConfigCommand) Action(ctx *kingpin.ParseContext) (err error) {
	var mysqlConfigContent string

	switch t.GlobalVar.Mode {
	case string(rdsv1alpha1.ModeMGRSP):
		mysqlConfigContent, err = t.mgrspConfig()
	case string(rdsv1alpha1.ModeMGRMP):
		mysqlConfigContent, err = t.mgrmpConfig()
	case string(rdsv1alpha1.ModeSemiSync):
		mysqlConfigContent, err = t.semiSyncConfig()
	}

	if t.Dump {
		os.Stdout.WriteString("generated config :\n" + mysqlConfigContent)
	}

	if err != nil {
		return err
	}

	err = os.WriteFile(t.ConfigFile, []byte(mysqlConfigContent), 0644)

	return err
}

func (t *MysqlConfigCommand) mgrspConfig() (fileContent string, err error) {
	var seeds string
	for _, v := range AddressesToDSN(t.GlobalVar.Addresses) {
		seeds += v.Host + ":33061,"
	}
	seeds = strings.Trim(seeds, ",")

	writer := mysql.NewConfigParser()
	if err = writer.Parse(strings.NewReader(BasicConf)); err != nil {
		return fileContent, fmt.Errorf("parse base mysql conf failed, err -> %s", err.Error())
	}
	mysqld := mysql.NewConfigSection("mysqld")

	mysqld.Set("plugin_load_add", "group_replication.so")
	mysqld.Set("transaction_write_set_extraction", "XXHASH64")
	mysqld.Set("loose-group_replication_group_name", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	mysqld.Set("loose-group_replication_start_on_boot", "off")
	mysqld.Set("loose-group_replication_bootstrap_group", "off")

	mysqld.Set("loose-group_replication_single_primary_mode", "on")
	mysqld.Set("loose-group_replication_enforce_update_everywhere_checks", "off")

	mysqld.Set("loose-group_replication_recovery_retry_count", "100")
	mysqld.Set("loose-group_replication_group_seeds", seeds)
	mysqld.Set("loose_group_replication_ip_whitelist", strings.Join(t.Whitelist, ","))

	mysqld.Set("loose_group_replication_local_address", os.Getenv("HOSTNAME")+":33061")

	mysqld.Set("server-id", strconv.Itoa(getMysqlServerID()))
	mysqld.Set("log_slave_updates", "ON")

	writer.MergeSection(mysqld)
	// merge extra config
	if err := writer.ParseFile(util.EnvOrDefault("MYSQL_CFG_EXTRA_DIR", "/etc/my.cnf.d") + "/extra_config"); err != nil && !os.IsNotExist(err) {
		return fileContent, err
	}
	fileContent = writer.String()
	return fileContent, nil
}

func (t *MysqlConfigCommand) mgrmpConfig() (fileContent string, err error) {
	var seeds string
	for _, v := range AddressesToDSN(t.GlobalVar.Addresses) {
		seeds += v.Host + ":33061,"
	}
	seeds = strings.Trim(seeds, ",")

	writer := mysql.NewConfigParser()
	if err = writer.Parse(strings.NewReader(BasicConf)); err != nil {
		return fileContent, fmt.Errorf("parse base mysql conf failed, err -> %s", err.Error())
	}
	mysqld := mysql.NewConfigSection("mysqld")

	mysqld.Set("plugin_load_add", "group_replication.so")
	mysqld.Set("transaction_write_set_extraction", "XXHASH64")
	mysqld.Set("loose-group_replication_group_name", "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")

	mysqld.Set("loose-group_replication_start_on_boot", "off")
	mysqld.Set("loose-group_replication_bootstrap_group", "off")

	mysqld.Set("loose-group_replication_single_primary_mode", "off")
	mysqld.Set("loose-group_replication_enforce_update_everywhere_checks", "on")

	mysqld.Set("loose-group_replication_recovery_retry_count", "100")
	mysqld.Set("loose-group_replication_group_seeds", seeds)
	mysqld.Set("loose_group_replication_ip_whitelist", strings.Join(t.Whitelist, ","))

	mysqld.Set("loose_group_replication_local_address", os.Getenv("HOSTNAME")+":33061")

	mysqld.Set("server-id", strconv.Itoa(getMysqlServerID()))
	mysqld.Set("log_slave_updates", "ON")

	writer.MergeSection(mysqld)
	// merge extra config
	if err := writer.ParseFile("/etc/my.cnf.d/extra_config"); err != nil && !os.IsNotExist(err) {
		return fileContent, err
	}
	fileContent = writer.String()
	return fileContent, nil
}

func (t *MysqlConfigCommand) semiSyncConfig() (fileContent string, err error) {
	writer := mysql.NewConfigParser()
	if err = writer.Parse(strings.NewReader(BasicConf)); err != nil {
		return fileContent, fmt.Errorf("parse base mysql conf failed, err -> %s", err.Error())
	}
	mysqld := mysql.NewConfigSection("mysqld")
	mysqld.Set("plugin_load_add", "semisync_master.so;semisync_slave.so")
	mysqld.Set("master_info_repository", "TABLE")
	mysqld.Set("rpl_semi_sync_master_wait_point", "AFTER_SYNC")
	mysqld.Set("log-slave-updates", "ON")
	mysqld.Set("slave-parallel-type", "LOGICAL_CLOCK")
	mysqld.Set("slave_parallel_workers", "16")
	mysqld.Set("server-id", strconv.Itoa(getMysqlServerID()))
	if t.GlobalVar.SemiSyncDoubleMasterHA { // avoid auto increment id conflict
		mysqld.Set("auto_increment_offset ", strconv.Itoa(getMysqlServerID()))
		mysqld.Set("auto_increment_increment ", "2")
	}

	writer.MergeSection(mysqld)
	// merge extra config
	if err := writer.ParseFile("/etc/my.cnf.d/extra_config"); err != nil && !os.IsNotExist(err) {
		return fileContent, err
	}
	fileContent = writer.String()
	return fileContent, nil
}

func getMysqlServerID() int {
	hostname := os.Getenv("HOSTNAME")
	arr := strings.Split(hostname, "-")
	idStr := arr[len(arr)-1]
	id, _ := strconv.Atoi(idStr)
	id += 1
	return id
}
