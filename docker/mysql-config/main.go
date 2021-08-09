package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/strategy"
	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/pkg/mysql"
	"github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
	"gorm.io/gorm"
)

var (
	whitelist            = kingpin.Flag("whitelist", "mysql variable: whitelist").Default(envOrDefault("WHITELIST", "10.0.0.0/8,192.168.1.0/24")).String()
	applierThreshold     = kingpin.Flag("applier-threshold", "mysql variable: loose-group_replication_flow_control_applier_threshold").Default(envOrDefault("APPLIER_THRESHOLD", "25000")).String()
	mgrLocalAddr         = kingpin.Flag("mgr-local-addr", "mysql variable: loose-group_replication_local_address").Default(envOrDefault("HOSTNAME", "127.0.0.1") + ":33061").String()
	mgrRetries           = kingpin.Flag("mgr-reties", "mysql variable loose-group_replication_recovery_retry_count").Default(envOrDefault("MGR_RETRIES", "100")).Int()
	mysqlExtraConfigDir  = kingpin.Flag("mysql-extra-config-file-dir", "my.cnf !include args").Default(envOrDefault("MYSQL_EXTRA_CONFIG_DIR", "/etc/my.cnf.d")).String()
	mgrSeeds             = kingpin.Flag("mgr-seeds", "mysql variable: loose-group_replication_local_address").Default(envOrDefault("MGR_SEEDS", "")).String()
	serverID             = kingpin.Flag("server-id", "mysql variable: server_id").Default(envOrDefault("SERVER_ID", getServerID())).String()
	clusterMode          = kingpin.Flag("cluster-mode", "mysql cluster mode").Default(envOrDefault("CLUSTER_MDOE", "MGRSP")).String()
	bootstrap            = kingpin.Flag("bootstrap", "run as bootstrap mode for boot mysql cluster").Default(envOrDefault("BOOTSTRAP_CLUSTER", "false")).Bool()
	mysqlRootPassword    = kingpin.Flag("mysql-root-password", "mysql root password").Default(envOrDefault("MYSQL_ROOT_PASSWORD", "")).String()
	mysqlBootUsers       = kingpin.Flag("mysql-boot-users", "json format user list").Default(envOrDefault("MYSQL_BOOT_USERS", "[]")).String()
	mysqlReplicationUser = kingpin.Flag("mysql-replication-users", "json format user list").Default(envOrDefault("MYSQL_REPLICATION_USER", "root")).String()
)

func init() {
	logrus.Infof("args -> %s", kingpin.Parse())
	logrus.Infof("clusterModel=%s ,serverID=%s, mysqlExtraConfigDir=%s, mgrLocalAddr=%s, mgrSeeds=%s", *clusterMode, *serverID, *mysqlExtraConfigDir, *mgrLocalAddr, *mgrSeeds)
}

func main() {
	if *bootstrap {
		bootstrapMgrCluster()
	} else {
		gernateConfig()
	}
}

func bootstrapMgrCluster() {
	var mysqlDSNs []string
	var replicationUser, replicationPassword string
	var db *gorm.DB

	seeds := strings.Split(*mgrSeeds, ",")
	for _, seed := range seeds {
		mysqlHostInfoArr := strings.Split(seed, ":")

		if mysqlHostID, err := strconv.Atoi(mysqlHostInfoArr[len(mysqlHostInfoArr)-1]); mysqlHostID > 0 && err != nil {
			mysqlDSN := fmt.Sprintf("root:%s@tcp(%s:3306)/mysql", *mysqlRootPassword, mysqlHostInfoArr[0])
			mysqlDSNs = append(mysqlDSNs, mysqlDSN)
		}
	}

	if len(seeds) < 1 {
		logrus.Fatal("mysql slave member list less than 1")
	}

	var users []rdsv1alpha1.MysqlUser

	if err := json.Unmarshal([]byte(*mysqlBootUsers), &users); err != nil {
		logrus.WithField("err", err.Error()).Fatal("parse mysql boot users list failed")
	}

	err := retry.Retry(func(attempt uint) (err error) {
		db, err = mysql.NewDB(fmt.Sprintf("root:%s@tcp(127.0.0.1:3306)/mysql", *mysqlRootPassword))
		return err
	},
		strategy.Limit(uint(*mgrRetries)),
		strategy.Backoff(backoff.Linear(time.Second*1)),
	)

	if err != nil {
		logrus.WithField("err", err.Error()).Fatal("connect local mysqld failed")
	}

	booter := mysql.NewMGRSinglePrimaryBoot(mysql.NewMGRSinglePrimaryBootOpts{DB: db})

	for _, v := range users {
		if err := booter.CheckUserUpdate(v.Username, v.Password, v.Domain, v.Privileges, v.DatabaseTarget); err != nil {
			logrus.WithField("err", err.Error()).Fatal("CheckUserUpdate failed")
		}

		if v.Username == *mysqlReplicationUser {
			replicationUser = v.Username
			replicationPassword = v.Password
		}
	}

	if replicationUser == "" {
		replicationUser = "root"
		replicationPassword = *mysqlRootPassword
	}

	if *serverID == "1" {
		if booter.CheckClusterHasAliveNode(mysqlDSNs) {
			logrus.Infof("serverID=%s join cluster", *serverID)
			err = booter.JoinCluster(replicationUser, replicationPassword)
		} else {
			logrus.Infof("serverID=%s boostrap cluster", *serverID)
			err = booter.BootCluster()
		}
	} else {
		logrus.Infof("serverID=%s join cluster", *serverID)
		err = booter.JoinCluster(replicationUser, replicationPassword)
	}

	if err != nil {
		logrus.WithField("err", err.Error()).Fatal("start cluster failed")
	}

	sqlDB, err := db.DB()
	if err != nil || sqlDB == nil {
		logrus.WithField("err", err.Error()).Fatal("try to close db conn failed")
	}

	sqlDB.Close()

	logrus.Info("all bootstrap action is successfully")

	for {
		time.Sleep(time.Second * 1)
	}
}

func gernateConfig() {
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
