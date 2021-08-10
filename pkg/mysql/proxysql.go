package mysql

import "strings"

type ProxySQLConfWriter struct {
	Datadir         string
	AdminVariables  map[string]string
	MysqlVariables  map[string]string
	MysqlServers    []map[string]string
	MysqlUsers      []map[string]string
	MysqlQueryRules []map[string]string
	Scheduler       []map[string]string
}

func (t *ProxySQLConfWriter) String() (conf string) {
	conf += "datadir=" + t.Datadir + "\n"
	conf += "admin_variables=\n(\n"
	for k, v := range t.AdminVariables {
		conf += "\t" + k + "=" + v + "\n"
	}
	conf += ")\n\n"

	conf += "mysql_variables=\n(\n"
	for k, v := range t.MysqlVariables {
		conf += "\t" + k + "=" + v + "\n"
	}
	conf += ")\n\n"

	conf += "mysql_users=\n(\n"
	for _, v := range t.MysqlServers {
		conf += "\t{ "
		for kk, vv := range v {
			conf += kk + "=" + vv + ", "
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
			conf += kk + "=" + vv + ", "
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
			conf += kk + "=" + vv + ", "
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
			conf += kk + "=" + vv + ", "
		}
		conf = strings.Trim(conf, ", ")
		conf += " },\n"
	}
	conf = strings.Trim(conf, ",\n")
	conf += "\n)\n\n"

	return conf
}
