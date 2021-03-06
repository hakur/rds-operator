package mysql

import (
	"regexp"
	"strings"
)

var numberRegxp = regexp.MustCompile(`^\d+$`)

type ProxySQLConfWriter struct {
	Datadir                         string
	AdminVariables                  map[string]string
	MysqlVariables                  map[string]string
	MysqlServers                    []map[string]string
	MysqlUsers                      []map[string]string
	MysqlQueryRules                 []map[string]string
	Scheduler                       []map[string]string
	ProxySQLServers                 []map[string]string
	MysqlReplicationHostgroups      []map[string]string
	MysqlGroupReplicationHostgroups []map[string]string
}

func (t *ProxySQLConfWriter) String() (conf string) {
	conf = "datadir=" + t.valueWrap(t.Datadir) + "\n"
	conf += "admin_variables=\n{\n"
	for k, v := range t.AdminVariables {
		conf += "\t" + k + "=" + t.valueWrap(v) + "\n"
	}
	conf += "}\n\n"

	conf += "mysql_variables=\n{\n"
	for k, v := range t.MysqlVariables {
		conf += "\t" + k + "=" + t.valueWrap(v) + "\n"
	}
	conf += "}\n\n"

	conf += "mysql_servers=\n(\n"
	for _, v := range t.MysqlServers {
		conf += "\t{ "
		for kk, vv := range v {
			conf += kk + "=" + t.valueWrap(vv) + ", "
		}
		conf = strings.Trim(conf, ", ")
		conf += " },\n"
	}
	conf = strings.Trim(conf, ",\n")
	conf += "\n)\n\n"

	conf += "mysql_users=\n(\n"
	for _, v := range t.MysqlUsers {
		conf += "\t{ "
		for kk, vv := range v {
			if kk == "password" {
				conf += kk + "=\"" + vv + "\", "
			} else {
				conf += kk + "=" + t.valueWrap(vv) + ", "
			}
		}
		conf = strings.Trim(conf, ", ")
		conf += " },\n"
	}
	conf = strings.Trim(conf, ",\n")
	conf += "\n)\n\n"

	conf += "mysql_query_rules=\n(\n"
	for _, v := range t.MysqlQueryRules {
		conf += "\t{ "
		for kk, vv := range v {
			conf += kk + "=" + t.valueWrap(vv) + ", "
		}
		conf = strings.Trim(conf, ", ")
		conf += " },\n"
	}
	conf = strings.Trim(conf, ",\n")
	conf += "\n)\n\n"

	conf += "scheduler=\n(\n"
	for _, v := range t.Scheduler {
		conf += "\t{ "
		for kk, vv := range v {
			conf += kk + "=" + t.valueWrap(vv) + ", "
		}
		conf = strings.Trim(conf, ", ")
		conf += " },\n"
	}
	conf = strings.Trim(conf, ",\n")
	conf += "\n)\n\n"

	conf += "proxysql_servers=\n(\n"
	for _, v := range t.ProxySQLServers {
		conf += "\t{ "
		for kk, vv := range v {
			conf += kk + "=" + t.valueWrap(vv) + ", "
		}
		conf = strings.Trim(conf, ", ")
		conf += " },\n"
	}
	conf = strings.Trim(conf, ",\n")
	conf += "\n)\n\n"

	conf += "mysql_replication_hostgroups=\n(\n"
	for _, v := range t.MysqlReplicationHostgroups {
		conf += "\t{ "
		for kk, vv := range v {
			conf += kk + "=" + t.valueWrap(vv) + ", "
		}
		conf = strings.Trim(conf, ", ")
		conf += " },\n"
	}
	conf = strings.Trim(conf, ",\n")
	conf += "\n)\n\n"

	conf += "mysql_group_replication_hostgroups=\n(\n"
	for _, v := range t.MysqlGroupReplicationHostgroups {
		conf += "\t{ "
		for kk, vv := range v {
			conf += kk + "=" + t.valueWrap(vv) + ", "
		}
		conf = strings.Trim(conf, ", ")
		conf += " },\n"
	}

	conf = strings.Trim(conf, ",\n")
	conf += "\n)\n\n"

	return conf
}

func (t *ProxySQLConfWriter) valueWrap(value string) (s string) {
	if numberRegxp.Match([]byte(value)) {
		s = value
	} else {
		s = `"` + value + `"`
	}
	return s
}
