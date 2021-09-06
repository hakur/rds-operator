package sidecar

import (
	"os"
	"strconv"
	"strings"

	"github.com/hakur/rds-operator/util"
	"gopkg.in/alecthomas/kingpin.v2"
)

type MysqlFlagValues struct {
	// Dump output generated mysqld config on stdout
	Dump bool
	// Mode mysql cluster mode
	Mode string
	// Nodes cluster peers, example "mysql-0,mysql-1,mysql-2"
	Nodes string
	// ConfigFile mysql config file path
	ConfigFile string
	// Whitelist mysql server whitelist
	Whitelist []string
	// Port mysql server port
	Port int
}

type MysqlCommand struct {
	flagVar *MysqlFlagValues
}

func (t *MysqlCommand) Register() {
	t.flagVar = new(MysqlFlagValues)
	mysqlConfigCommandHandler := &mysqlConfigCommand{flagVar: t.flagVar}

	mysqlCmd := kingpin.Command("mysql", "mysql tools")
	mysqlCmd.Flag("cluster-mode", "mysql cluster mode").Default(util.EnvOrDefault("MYSQL_CLUSTER_MODE", "MGRSP")).EnumVar(&t.flagVar.Mode, "MGRSP", "MGRMP", "SemiSync", "Async")
	mysqlCmd.Flag("nodes", "mysql cluster peers node").Default(util.EnvOrDefault("MYSQL_NODES", "")).StringVar(&t.flagVar.Nodes)

	configCmd := mysqlCmd.Command("cfg", "generate mysql config").Action(mysqlConfigCommandHandler.action)
	configCmd.Flag("config-file", "mysqld config file path").Default(util.EnvOrDefault("MYSQL_CFG_EXTRA_DIR", "/etc/my.cnf.d") + "/my.cnf").StringVar(&t.flagVar.ConfigFile)
	configCmd.Flag("white-list", "mysql server white list").Default(util.EnvOrDefault("MYSQL_CFG_WHITE_LIST", "10.0.0.0/8,192.0.0.0/8")).StringsVar(&t.flagVar.Whitelist)
	configCmd.Flag("port", "mysql server listen port").Default(util.EnvOrDefault("MYSQL_CFG_PORT", "3306")).IntVar(&t.flagVar.Port)
	configCmd.Flag("dump", "output generated mysqld config on stdout, to enable in format --dump without any argument").Default(util.EnvOrDefault("MYSQL_CFG_DUMP", "false")).BoolVar(&t.flagVar.Dump)
}

func getMysqlServerID() int {
	hostname := os.Getenv("HOSTNAME")
	arr := strings.Split(hostname, "-")
	idStr := arr[len(arr)-1]
	id, _ := strconv.Atoi(idStr)
	id += 1
	return id
}
