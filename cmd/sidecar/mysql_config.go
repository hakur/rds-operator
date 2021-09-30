package main

import (
	"os"
	"strconv"
	"strings"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/pkg/mysql"
	"github.com/hakur/rds-operator/util"
	"gopkg.in/alecthomas/kingpin.v2"
)

type mysqlConfigCommand struct {
	flagVar *MysqlFlagValues
}

func (t *mysqlConfigCommand) action(ctx *kingpin.ParseContext) (err error) {
	var mysqlConfigContent string

	switch t.flagVar.Mode {
	case string(rdsv1alpha1.ModeMGRSP):
		mysqlConfigContent, err = t.mgrspConfig()
	case string(rdsv1alpha1.ModeMGRMP):
		mysqlConfigContent, err = t.mgrmpConfig()
	case string(rdsv1alpha1.ModeSemiSync):
		mysqlConfigContent, err = t.semiSyncConfig()
	}

	if t.flagVar.Dump {
		os.Stdout.WriteString("generated config :\n" + mysqlConfigContent)
	}

	if err != nil {
		return err
	}

	err = os.WriteFile(t.flagVar.ConfigFile, []byte(mysqlConfigContent), 0644)

	return err
}

func (t *mysqlConfigCommand) mgrspConfig() (fileContent string, err error) {
	var seeds string
	nodes := strings.Split(t.flagVar.Nodes, " ")
	for _, v := range nodes {
		seeds += v + ":33061,"
	}
	seeds = strings.Trim(seeds, ",")

	writer := mysql.NewConfigParser()
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
	mysqld.Set("loose_group_replication_ip_whitelist", strings.Join(t.flagVar.Whitelist, ","))

	mysqld.Set("loose_group_replication_local_address", os.Getenv("HOSTNAME")+":33061")

	mysqld.Set("server-id", strconv.Itoa(getMysqlServerID()))
	mysqld.Set("log_slave_updates", "ON")

	writer.SetSection(mysqld)
	// merge extra config
	if err := writer.ParseFile(util.EnvOrDefault("MYSQL_CFG_EXTRA_DIR", "/etc/my.cnf.d") + "/extra_config"); err != nil && !os.IsNotExist(err) {
		return fileContent, err
	}
	fileContent = writer.String()
	return fileContent, nil
}

func (t *mysqlConfigCommand) mgrmpConfig() (fileContent string, err error) {
	var seeds string
	nodes := strings.Split(t.flagVar.Nodes, " ")
	for _, v := range nodes {
		seeds += v + ":33061,"
	}
	seeds = strings.Trim(seeds, ",")

	writer := mysql.NewConfigParser()
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
	mysqld.Set("loose_group_replication_ip_whitelist", strings.Join(t.flagVar.Whitelist, ","))

	mysqld.Set("loose_group_replication_local_address", os.Getenv("HOSTNAME")+":33061")

	mysqld.Set("server-id", strconv.Itoa(getMysqlServerID()))
	mysqld.Set("log_slave_updates", "ON")

	writer.SetSection(mysqld)
	// merge extra config
	if err := writer.ParseFile("/etc/my.cnf.d/extra_config"); err != nil && !os.IsNotExist(err) {
		return fileContent, err
	}
	fileContent = writer.String()
	return fileContent, nil
}

func (t *mysqlConfigCommand) semiSyncConfig() (fileContent string, err error) {
	writer := mysql.NewConfigParser()
	mysqld := mysql.NewConfigSection("mysqld")
	mysqld.Set("plugin_load_add", "semisync_master.so;semisync_slave.so")
	mysqld.Set("master_info_repository", "TABLE")
	mysqld.Set("rpl_semi_sync_master_wait_point", "AFTER_SYNC")
	mysqld.Set("log-slave-updates", "ON")
	mysqld.Set("slave-parallel-type", "LOGICAL_CLOCK")
	mysqld.Set("slave_parallel_workers", "16")
	mysqld.Set("server-id", strconv.Itoa(getMysqlServerID()))
	if t.flagVar.SemiSyncDoubleMasterHA { // avoid auto increment id conflict
		mysqld.Set("auto_increment_offset ", strconv.Itoa(getMysqlServerID()))
		mysqld.Set("auto_increment_increment ", "2")
	}

	writer.SetSection(mysqld)
	// merge extra config
	if err := writer.ParseFile("/etc/my.cnf.d/extra_config"); err != nil && !os.IsNotExist(err) {
		return fileContent, err
	}
	fileContent = writer.String()
	return fileContent, nil
}
