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
	configType = kingpin.Flag("config-type", "type of config gernate func,values are [ proxysql mysql ]").Default(envOrDefault("CONFIG_TYPE", "mysql")).String()
	// mysql options
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
	mysqlMaxConn         = kingpin.Flag("mysql-max-conn", "max conn on each mysql server").Default(envOrDefault("MYSQL_MAX_CONN", "300")).Int()
	mysqlServerVersion   = kingpin.Flag("mysql-server-version", "mysql server version").Default(envOrDefault("MYSQL_SERVER_VERSION", "5.7.34")).String()

	// proxysql options
	proxysqlMonitorUser    = kingpin.Flag("proxysql-monitor-user", "username on mysql servers but used for proxysql monitor").Default(envOrDefault("PROXYSQL_MONITOR_USER", "monitor")).String()
	proxysqlMonitorPasword = kingpin.Flag("proxysql-monitor-password", "passworf of proxysql monitor user").Default(envOrDefault("PROXYSQL_MONITOR_PASSWORD", "monitor_me")).String()
	proxysqlAdminUser      = kingpin.Flag("proxysql-admin-user", "proxysql admin user").Default(envOrDefault("PROXYSQL_ADMIN_USER", "admin")).String()
	proxysqlAdminPasword   = kingpin.Flag("proxysql-admin-password", "passworf of proxysql admin user").Default(envOrDefault("PROXYSQL_ADMIN_PASSWORD", "admin_pwd")).String()
)

func init() {
	kingpin.Parse()
	logrus.Infof("clusterModel=%s ,serverID=%s, mysqlExtraConfigDir=%s, mgrLocalAddr=%s, mgrSeeds=%s", *clusterMode, *serverID, *mysqlExtraConfigDir, *mgrLocalAddr, *mgrSeeds)
}

func main() {
	if *bootstrap {
		bootstrapMgrCluster()
	} else {
		if *configType == "mysql" {
			gernateMysqlConfig()
		} else {
			gernateProxySQLConfig()
		}
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

func envOrDefault(key string, defaultValue string) (value string) {
	value = strings.TrimSpace(os.Getenv(key))
	if value == "" {
		value = defaultValue
	}
	return value
}
