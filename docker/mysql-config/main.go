package main

import (
	"os"
	"strconv"
	"strings"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/pkg/mysql"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	whitelist           = kingpin.Flag("whitelist", "mysql variable: whitelist").Default(envOrDefault("WHITELIST", "10.0.0.0/8,192.168.1.0/24")).String()
	applierThreshold    = kingpin.Flag("applier-threshold", "mysql variable: loose-group_replication_flow_control_applier_threshold").Default(envOrDefault("APPLIER_THRESHOLD", "25000")).String()
	mgrLocalAddr        = kingpin.Flag("mgr-local-addr", "mysql variable: loose-group_replication_local_address").Default(envOrDefault("HOSTNAME", "127.0.0.1") + ":33061").String()
	mgrRetries          = kingpin.Flag("mgr-reties", "mysql variable loose-group_replication_recovery_retry_count").Default(envOrDefault("MGR_RETRIES", "100")).Int()
	mysqlExtraConfigDir = kingpin.Flag("mysql-extra-config-file-dir", "my.cnf !include args").Default(envOrDefault("MYSQL_EXTRA_CONFIG_DIR", "/etc/my.cnf.d")).String()
	mgrSeeds            = kingpin.Flag("mgr-seeds", "mysql variable: loose-group_replication_local_address").Default(envOrDefault("MGR_SEEDS", "")).String()
	serverID            = kingpin.Flag("server-id", "mysql variable: server_id").Default(envOrDefault("SERVER_ID", getServerID())).String()
	clusterMode         = kingpin.Flag("cluster-mode", "mysql cluster mode").Default(envOrDefault("CLUSTER_MDOE", "MGRSP")).String()
)

func init() {
	logrus.Infof("args -> %s", kingpin.Parse())
	logrus.Infof("clusterModel=%s ,serverID=%s, mysqlExtraConfigDir=%s, mgrLocalAddr=%s, mgrSeeds=%s", *clusterMode, *serverID, *mysqlExtraConfigDir, *mgrLocalAddr, *mgrSeeds)
}

func main() {
	emptyConfig := `
	[client]
	[mysqld]
	`
	parser := mysql.NewConfigParser()
	parser.Parse(strings.NewReader(emptyConfig))

	clientSeciont, _ := parser.GetSection("client")
	for k, v := range mysql.MysqlClientVars {
		clientSeciont.Set(k, v)
	}

	mysqldSection, _ := parser.GetSection("mysqld")
	for k, v := range mysql.MysqldVars {
		mysqldSection.Set(k, v)
	}

	mysqldSection.Set("server-id", *serverID)
	if *clusterMode == string(rdsv1alpha1.ModeMGRSP) {
		addMGRSPVars(mysqldSection)
	}

	parser.SetSection("client", clientSeciont)
	parser.SetSection("mysqld", mysqldSection)

	cnfContent := parser.String() + "\n"

	if err := os.WriteFile(*mysqlExtraConfigDir+"/my.cnf", []byte(cnfContent), 0755); err != nil {
		logrus.WithField("err", err.Error()).Fatalf("write file to %s failed", *mysqlExtraConfigDir+"/my.cnf")
	}
	logrus.Infof("write file to %s successfully", *mysqlExtraConfigDir+"/my.cnf")
}

func addMGRSPVars(section *mysql.ConfigSection) {
	for k, v := range mysql.SiglePrimaryVars {
		section.Set(k, v)
	}

	section.Set("loose-group_replication_group_seeds", *mgrSeeds)
	section.Set("loose_group_replication_ip_whitelist", *whitelist)
	section.Set("loose-group_replication_flow_control_applier_threshold", *applierThreshold)
	section.Set("loose-group_replication_local_address", *mgrLocalAddr)
	section.Set("loose-group_replication_recovery_retry_count", strconv.Itoa(*mgrRetries))
	section.Set("loose-group_replication_start_on_boot", "off")
	section.Set("loose-group_replication_bootstrap_group", "off")
}

func envOrDefault(key string, defaultValue string) (value string) {
	value = strings.TrimSpace(os.Getenv(key))
	if value == "" {
		value = defaultValue
	}
	return value
}

// getServerID get server id from kubernetes statefulset's pod hostname,it an string by in int format
func getServerID() string {
	arr := strings.Split(os.Getenv("HOSTNAME"), "-")
	sid, _ := strconv.Atoi(arr[len(arr)-1])
	sid++
	return strconv.Itoa(sid)
}
