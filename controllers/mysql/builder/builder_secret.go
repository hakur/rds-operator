package builder

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// BuildSecret generate secret environment variables for mysql pods
func BuildSecret(cr *rdsv1alpha1.Mysql) (secret *corev1.Secret) {
	var seeds string
	var semiSyncMasters string
	var semiSyncDoubleMaster bool
	mysqlMaxConn := "300"
	if cr.Spec.MaxConn != nil {
		mysqlMaxConn = strconv.Itoa(*cr.Spec.MaxConn)
	}

	for i := 0; i < int(*cr.Spec.Replicas); i++ {
		mysqlhost := cr.Name + "-mysql-" + strconv.Itoa(i) + " "
		seeds += mysqlhost

		if i == 0 {
			semiSyncMasters = mysqlhost
		}

		if cr.Spec.SemiSync != nil && cr.Spec.SemiSync.DoubleMasterHA && i == 1 {
			semiSyncDoubleMaster = true
			semiSyncMasters += mysqlhost
		}
	}
	seeds = strings.Trim(seeds, " ")
	semiSyncMasters = strings.Trim(semiSyncMasters, " ")

	secret = new(corev1.Secret)
	secret.APIVersion = "v1"
	secret.Kind = "Secret"
	secret.Name = cr.Name + "-mysql-secret"
	secret.Namespace = cr.Namespace
	secret.Labels = BuildMysqlLabels(cr)

	secret.Data = map[string][]byte{
		"TZ":                   []byte(cr.Spec.TimeZone),
		"MYSQL_ROOT_PASSWORD":  []byte(*cr.Spec.RootPassword),
		"MYSQL_DATA_DIR":       []byte("/var/lib/mysql"),
		"MYSQL_CLUSTER_MODE":   []byte(string(cr.Spec.ClusterMode)),
		"MYSQL_CFG_MAX_CONN":   []byte(mysqlMaxConn),
		"MYSQL_CFG_WHITE_LIST": []byte(strings.Join(cr.Spec.Whitelist, ",")),
		"MYSQL_ROOT_HOST":      []byte("%"),
		"MYSQL_NODES":          []byte(seeds),
	}

	if cr.Spec.ClusterMode == rdsv1alpha1.ModeSemiSync {
		secret.Data["SEMI_SYNC_DOUBLE_MASTER_HA"] = []byte(strconv.FormatBool(semiSyncDoubleMaster))
		secret.Data["SEMI_SYNC_FIXED_MASTERS"] = []byte(semiSyncMasters)
	}

	if cr.Spec.ClusterUser != nil {
		mysqlPassword, _ := base64.StdEncoding.DecodeString(cr.Spec.ClusterUser.Password)
		sql := fmt.Sprintf(`
			USE mysql;
			CREATE USER IF NOT EXISTS %s@'%s' IDENTIFIED WITH mysql_native_password BY '%s';
			GRANT %s ON %s TO %s@'%s';
			FLUSH PRIVILEGES;
		`,
			cr.Spec.ClusterUser.Username,
			cr.Spec.ClusterUser.Domain,
			mysqlPassword,
			strings.Join(cr.Spec.ClusterUser.Privileges, ","),
			cr.Spec.ClusterUser.DatabaseTarget,
			cr.Spec.ClusterUser.Username,
			cr.Spec.ClusterUser.Domain,
		)
		secret.Data["init.sql"] = []byte(sql)
	}

	if cr.Spec.ExtraConfigDir != nil {
		secret.Data["MYSQL_CFG_EXTRA_DIR"] = []byte(*cr.Spec.ExtraConfigDir)
	}
	return secret
}
