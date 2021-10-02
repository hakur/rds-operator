package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
)

type MysqlBackupCommand struct {
	// Username username used for backup operation, password use env var MYSQL_PWD
	Username  string
	GlobalVar *MysqlGlobalFlagValues
}

func (t *MysqlBackupCommand) Register(cmd *kingpin.CmdClause) {
	cmd.Action(t.Action)
	cmd.Flag("username", "mysql username").Default("root").StringVar(&t.Username)
}

func (t *MysqlBackupCommand) Action(ctx *kingpin.ParseContext) (err error) {
	return err
}
