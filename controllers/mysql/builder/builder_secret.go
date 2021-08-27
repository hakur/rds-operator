package builder

import (
	"encoding/json"
	"strconv"
	"strings"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// BuildSecret generate secret environment variables for mysql pods and proxysql pods
func BuildSecret(cr *rdsv1alpha1.Mysql) (secret *corev1.Secret) {
	var seeds string
	var mysqlUsers string
	mysqlMaxConn := "300"
	if cr.Spec.MaxConn != nil {
		mysqlMaxConn = strconv.Itoa(*cr.Spec.MaxConn)
	}

	for i := 0; i < int(*cr.Spec.Replicas); i++ {
		seeds += cr.Name + "-mysql-" + strconv.Itoa(i) + " "
	}
	seeds = strings.Trim(seeds, " ")

	buf, _ := json.Marshal(cr.Spec.Users)
	mysqlUsers = string(buf)

	secret = new(corev1.Secret)
	secret.APIVersion = "v1"
	secret.Kind = "ConfigMap"
	secret.Name = cr.Name + "-mysql-secret"
	secret.Namespace = cr.Namespace
	secret.Labels = BuildMysqlLabels(cr)

	secret.Data = map[string][]byte{
		"TZ":                   []byte(cr.Spec.TimeZone),
		"MYSQL_ROOT_PASSWORD":  []byte(*cr.Spec.RootPassword),
		"MYSQL_DATA_DIR":       []byte("/var/lib/mysql"),
		"MYSQL_BOOT_USERS":     []byte(mysqlUsers),
		"MYSQL_CLUSTER_MODE":   []byte(string(cr.Spec.ClusterMode)),
		"MYSQL_CFG_MAX_CONN":   []byte(mysqlMaxConn),
		"MYSQL_CFG_WHITE_LIST": []byte(strings.Join(cr.Spec.Whitelist, ",")),
		"MYSQL_ROOT_HOST":      []byte("%"),
		"MYSQL_NODES":          []byte(seeds),
		"MYSQL_REPL_USER":      []byte("root"),
		"MYSQL_REPL_PASSWORD":  []byte(*cr.Spec.RootPassword),
	}

	if cr.Spec.ExtraConfigDir != nil {
		secret.Data["MYSQL_CFG_EXTRA_DIR"] = []byte(*cr.Spec.ExtraConfigDir)
	}
	return secret
}