package mysql

import (
	"encoding/json"
	"strconv"
	"strings"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// buildSecret generate secret environment variables for mysql pods and proxysql pods
func buildSecret(cr *rdsv1alpha1.Mysql) (secret *corev1.Secret) {
	var seeds string
	var mysqlUsers string
	mysqlMaxConn := "300"
	if cr.Spec.Mysql.MaxConn != nil {
		mysqlMaxConn = strconv.Itoa(*cr.Spec.Mysql.MaxConn)
	}

	for i := 0; i < int(*cr.Spec.Mysql.Replicas); i++ {
		seeds += cr.Name + "-" + strconv.Itoa(i) + ":33061,"
	}
	seeds = strings.Trim(seeds, ",")

	buf, _ := json.Marshal(cr.Spec.Mysql.Users)
	mysqlUsers = string(buf)

	secret = new(corev1.Secret)
	secret.APIVersion = "v1"
	secret.Kind = "ConfigMap"
	secret.Name = cr.Name + "-secret"
	secret.Namespace = cr.Namespace

	secret.Data = map[string][]byte{
		// pod
		"TZ":                   []byte(cr.Spec.TimeZone),
		"POD_DNS_SERVICE_NAME": []byte(cr.Name),
		"POD_TOTAL_REPLICAS":   []byte(strconv.Itoa(int(*cr.Spec.Mysql.Replicas))),
		// mysql
		"MYSQL_ROOT_PASSWORD":    []byte(*cr.Spec.RootPassword),
		"MYSQL_DATA_DIR":         []byte("/var/lib/mysql"),
		"MYSQL_REPLICATION_USER": []byte(cr.Spec.Mysql.ReplicationUser),
		"MYSQL_BOOT_USERS":       []byte(mysqlUsers),
		"MYSQL_MAX_CONN":         []byte(mysqlMaxConn),
		"MYSQL_SERVER_VERSION":   []byte(cr.Spec.ProxySQL.MysqlVersion),
		"WHITELIST":              []byte(strings.Join(cr.Spec.Mysql.Whitelist, ",")),
		"CLUSTER_MODE":           []byte(string(cr.Spec.ClusterMode)),
		"MYSQL_ROOT_HOST":        []byte("%"),
		// mysql mgr
		"APPLIER_THRESHOLD":         []byte(strconv.Itoa(cr.Spec.Mysql.MGRSP.ApplierThreshold)),
		"MGR_RETRIES":               []byte(strconv.Itoa(cr.Spec.Mysql.MGRSP.MGRRetries)),
		"MGR_SEEDS":                 []byte(seeds),
		"PROXYSQL_MONITOR_USER":     []byte(cr.Spec.ProxySQL.MonitorUser),
		"PROXYSQL_MONITOR_PASSWORD": []byte(cr.Spec.ProxySQL.MonitorPassword),
		"PROXYSQL_ADMIN_USER":       []byte(cr.Spec.ProxySQL.AdminUser),
		"PROXYSQL_ADMIN_PASSWORD":   []byte(cr.Spec.ProxySQL.AdminPassword),
	}

	if cr.Spec.Mysql.ExtraConfigDir != nil {
		secret.Data["MYSQL_EXTRA_CONFIG_DIR"] = []byte(*cr.Spec.Mysql.ExtraConfigDir)
	}
	return secret
}
