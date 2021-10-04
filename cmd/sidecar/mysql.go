package main

import (
	"strconv"
	"strings"

	"github.com/hakur/rds-operator/pkg/mysql"
	"github.com/hakur/rds-operator/util"
	"gopkg.in/alecthomas/kingpin.v2"
)

type MysqlGlobalFlagValues struct {
	// Addresses (host|ip:port) string, splited by comma
	Addresses []string
	// Mode mysql cluster mode
	Mode string
	// SemiSyncDoubleMasterHA is cluster under semi sync replication mode, and there need two master node join each other
	SemiSyncDoubleMasterHA bool
	// MysqlVersion mysql server version must be X.X.X number
	MysqlVersion string
}

type MysqlCommand struct {
	GlobalVar *MysqlGlobalFlagValues
}

func (t *MysqlCommand) Register() {
	t.GlobalVar = new(MysqlGlobalFlagValues)

	mysqlCmd := kingpin.Command("mysql", "mysql tools")

	mysqlCmd.Flag("cluster-mode", "mysql cluster mode").Default(util.EnvOrDefault("MYSQL_CLUSTER_MODE", "MGRSP")).EnumVar(&t.GlobalVar.Mode, "MGRSP", "MGRMP", "SemiSync")
	mysqlCmd.Flag("address", "mysql address ,(host|ip):port string, use '--address=127.0.0.1:3306 --address=127.0.0.1:3307 --address=127.0.0.1:3308' for multiple addresses").StringsVar(&t.GlobalVar.Addresses)
	mysqlCmd.Flag("semi-sync-double-master-ha", "is cluster under semi sync replication mode, and there need two master node join each other").Default(util.EnvOrDefault("SEMI_SYNC_DOUBLE_MASTER_HA", "false")).BoolVar(&t.GlobalVar.SemiSyncDoubleMasterHA)
	mysqlCmd.Flag("version", "mysql server version").Default(util.EnvOrDefault("MYSQL_VERSION", "5.7.34")).StringVar(&t.GlobalVar.MysqlVersion)

	(&MysqlBackupCommand{GlobalVar: t.GlobalVar}).Register(mysqlCmd.Command("backup", "mysql backup"))
	(&MysqlConfigCommand{GlobalVar: t.GlobalVar}).Register(mysqlCmd.Command("cfg", "generate mysql config"))
}

// AddressesToDSN convert host/ip:port to dsn list
func AddressesToDSN(addresses []string) (data []*mysql.DSN) {
	var host string
	var port int
	for _, address := range addresses {
		arr := strings.Split(address, ":")
		if len(arr) > 1 {
			host = arr[0]
			if port, _ = strconv.Atoi(arr[1]); port < 1 {
				continue
			}
		} else if len(arr) == 1 {
			host = arr[0]
			port = 3306
		} else {
			continue
		}

		data = append(data, &mysql.DSN{
			Host: host,
			Port: port,
		})
	}
	return
}
