package mysql

import "testing"

func TestConfWriter(t *testing.T) {
	var cw ProxySQLConfWriter
	cw.Datadir = "/var/lib/proxysql"
	cw.AdminVariables = map[string]string{
		"aa": "bb",
		"dd": "cc",
	}

	aaa := map[string]string{
		"user":     "root",
		"password": "abc",
	}
	bbb := map[string]string{
		"user":     "root-2",
		"password": "abc-2",
	}
	cw.MysqlUsers = []map[string]string{
		aaa,
		bbb,
	}

	println(cw.String())
}
