package main

import (
	"os"
	"strconv"
	"strings"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/pkg/mysql"
	"github.com/sirupsen/logrus"
)

func gernateMysqlConfig() {
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

// getServerID get server id from kubernetes statefulset's pod hostname,it an string by in int format
func getServerID() string {
	arr := strings.Split(os.Getenv("HOSTNAME"), "-")
	sid, _ := strconv.Atoi(arr[len(arr)-1])
	sid++
	return strconv.Itoa(sid)
}
