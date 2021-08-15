package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/hakur/rds-operator/pkg/mysql"
	"github.com/sirupsen/logrus"
)

// gernateProxySQLConfig generate proxysql
func gernateProxySQLConfig() {
	var cw mysql.ProxySQLConfWriter
	cw.Datadir = "/var/lib/proxysql"
	cw.AdminVariables = map[string]string{
		"admin_credentials": *proxysqlAdminUser + ":" + *proxysqlAdminPasword,
		"mysql_ifaces":      "0.0.0.0:6032",
	}

	cw.MysqlVariables = map[string]string{
		"threads":                    "4",
		"max_connections":            strconv.Itoa(*mysqlMaxConn),
		"default_query_delay":        "0",
		"default_query_timeout":      "10000",
		"have_compress":              "true",
		"poll_timeout":               "2000",
		"interfaces":                 "0.0.0.0:3306",
		"default_schema":             "information_schema",
		"stacksize":                  "1048576",
		"server_version":             *mysqlServerVersion,
		"connect_timeout_server":     "2000",
		"monitor_history":            "600000",
		"monitor_connect_interval":   "10000",
		"monitor_ping_interval":      "5000",
		"monitor_read_only_interval": "1500",
		"monitor_read_only_timeout":  "2000",
		"ping_interval_server_msec":  "3000",
		"ping_timeout_server":        "5000",
		"commands_stats":             "true",
		"sessions_sort":              "true",
		"connect_retries_on_failure": "10",
		"monitor_username":           *proxysqlMonitorUser,
		"monitor_password":           *proxysqlMonitorPasword,
	}

	seeds := strings.Split(*mgrSeeds, ",")
	var mysqlHostList []string
	for _, seed := range seeds {
		mysqlHostInfoArr := strings.Split(seed, ":")
		mysqlHostList = append(mysqlHostList, mysqlHostInfoArr[0])

		if GetMysqlServerIDByPodName(mysqlHostInfoArr[0]) == 1 {
			cw.MysqlServers = append(cw.MysqlServers, map[string]string{
				"address":         mysqlHostInfoArr[0],
				"port":            "3306",
				"hostgroup":       "1",
				"max_connections": strconv.Itoa(*mysqlMaxConn),
			})
		} else {
			cw.MysqlServers = append(cw.MysqlServers, map[string]string{
				"address":         mysqlHostInfoArr[0],
				"port":            "3306",
				"hostgroup":       "2",
				"max_connections": strconv.Itoa(*mysqlMaxConn),
			})
		}
	}

	if len(seeds) < 1 {
		logrus.Fatal("mysql slave member list less than 1")
	}

	// cw.MysqlUsers = append(cw.MysqlUsers, map[string]string{
	// 	"username":               "root",
	// 	"password":               *mysqlRootPassword,
	// 	"default_hostgroup":      "1",
	// 	"transaction_persistent": "1",
	// 	"active":                 "1",
	// })

	cw.MysqlQueryRules = generateProxySQLQueryRules()
	cw.Scheduler = append(cw.Scheduler, map[string]string{
		"id":          "1",
		"interval_ms": "5000",
		"filename":    "/mgrsp_switch.sh",
		"arg1":        *proxysqlAdminUser,
		"arg2":        *proxysqlAdminPasword,
		"arg3":        strings.Join(mysqlHostList, " "),
	})
	cw.Scheduler = append(cw.Scheduler, map[string]string{
		"id":          "2",
		"interval_ms": "60000",
		"filename":    "/load_servers.sh",
		"arg1":        *proxysqlAdminUser,
		"arg2":        *proxysqlAdminPasword,
	})

	if err := os.WriteFile("/etc/proxysql.cnf.d/proxysql.cnf", []byte(cw.String()), 0755); err != nil {
		logrus.WithField("err", err.Error()).Fatal("write /etc/proxysql.cnf.d/proxysql.cnf failed")
	}
}

func generateProxySQLQueryRules() (data []map[string]string) {
	data = append(data, map[string]string{
		"rule_id":               "1",
		"active":                "1",
		"match_pattern":         "^SELECT .* FOR UPDATE$",
		"destination_hostgroup": "1",
		"apply":                 "1",
	})
	data = append(data, map[string]string{
		"rule_id":               "2",
		"active":                "1",
		"match_pattern":         "^CREATE .* FOR UPDATE$",
		"destination_hostgroup": "2",
		"apply":                 "1",
	})
	data = append(data, map[string]string{
		"rule_id":               "3",
		"active":                "1",
		"match_pattern":         "^(UPDATE)|(INSERT)|(CREATE)|(DELETE)|(DROP)|(ALERT)|(MODIFY)",
		"destination_hostgroup": "1",
		"apply":                 "1",
	})
	data = append(data, map[string]string{
		"rule_id":               "4",
		"active":                "1",
		"match_pattern":         "^SELECT",
		"destination_hostgroup": "1",
		"apply":                 "1",
	})
	return data
}

func GetMysqlServerIDByPodName(podName string) (serverID int) {
	arr := strings.Split(podName, "-")
	if len(arr) > 1 {
		serverID, _ = strconv.Atoi(arr[len(arr)-1])
		serverID++
	}
	return serverID
}
