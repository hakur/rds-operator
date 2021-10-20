package main

import (
	"github.com/hakur/rds-operator/util"
	"gopkg.in/alecthomas/kingpin.v2"
)

type ProxySQLCommandFlagValues struct {
	// Mode mysql cluster mode
	Mode string
}

type ProxySQLCommand struct {
	GlobalVar *ProxySQLCommandFlagValues
}

func (t *ProxySQLCommand) Register() {
	t.GlobalVar = new(ProxySQLCommandFlagValues)

	cmd := kingpin.Command("proxysql", "proxysql tools")
	cmd.Flag("cluster-mode", "mysql cluster mode").Default(util.EnvOrDefault("MYSQL_CLUSTER_MODE", "MGRSP")).EnumVar(&t.GlobalVar.Mode, "MGRSP", "MGRMP", "SemiSync")

	(&ProxySQLConfigCommand{GlobalVar: t.GlobalVar}).Register(cmd.Command("cfg", "proxysql config generator"))
}
