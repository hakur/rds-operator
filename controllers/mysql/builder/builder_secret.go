package builder

import (
	"fmt"
	"strconv"
	"strings"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/util"
	corev1 "k8s.io/api/core/v1"
)

var mgrMonitorView = `
USE sys;

DROP VIEW IF EXISTS gr_member_routing_candidate_status;

DROP FUNCTION IF EXISTS IFZERO;
DROP FUNCTION IF EXISTS LOCATE2;
DROP FUNCTION IF EXISTS GTID_NORMALIZE;
DROP FUNCTION IF EXISTS GTID_COUNT;
DROP FUNCTION IF EXISTS gr_applier_queue_length;
DROP FUNCTION IF EXISTS gr_member_in_primary_partition;
DROP FUNCTION IF EXISTS gr_transactions_to_cert;
DROP FUNCTION IF EXISTS my_server_uuid;

DELIMITER $$

CREATE FUNCTION IFZERO(a INT, b INT)
RETURNS INT
DETERMINISTIC
RETURN IF(a = 0, b, a)$$

CREATE FUNCTION LOCATE2(needle TEXT(10000), haystack TEXT(10000), offset INT)
RETURNS INT
DETERMINISTIC
RETURN IFZERO(LOCATE(needle, haystack, offset), LENGTH(haystack) + 1)$$

CREATE FUNCTION GTID_NORMALIZE(g TEXT(10000))
RETURNS TEXT(10000)
DETERMINISTIC
RETURN GTID_SUBTRACT(g, '')$$

CREATE FUNCTION GTID_COUNT(gtid_set TEXT(10000))
RETURNS INT
DETERMINISTIC
BEGIN
  DECLARE result BIGINT DEFAULT 0;
  DECLARE colon_pos INT;
  DECLARE next_dash_pos INT;
  DECLARE next_colon_pos INT;
  DECLARE next_comma_pos INT;
  SET gtid_set = GTID_NORMALIZE(gtid_set);
  SET colon_pos = LOCATE2(':', gtid_set, 1);
  WHILE colon_pos != LENGTH(gtid_set) + 1 DO
     SET next_dash_pos = LOCATE2('-', gtid_set, colon_pos + 1);
     SET next_colon_pos = LOCATE2(':', gtid_set, colon_pos + 1);
     SET next_comma_pos = LOCATE2(',', gtid_set, colon_pos + 1);
     IF next_dash_pos < next_colon_pos AND next_dash_pos < next_comma_pos THEN
       SET result = result +
         SUBSTR(gtid_set, next_dash_pos + 1,
                LEAST(next_colon_pos, next_comma_pos) - (next_dash_pos + 1)) -
         SUBSTR(gtid_set, colon_pos + 1, next_dash_pos - (colon_pos + 1)) + 1;
     ELSE
       SET result = result + 1;
     END IF;
     SET colon_pos = next_colon_pos;
  END WHILE;
  RETURN result;
END$$

CREATE FUNCTION gr_applier_queue_length()
RETURNS INT
DETERMINISTIC
BEGIN
  RETURN (SELECT sys.gtid_count( GTID_SUBTRACT( (SELECT
Received_transaction_set FROM performance_schema.replication_connection_status
WHERE Channel_name = 'group_replication_applier' ), (SELECT
@@global.GTID_EXECUTED) )));
END$$


CREATE FUNCTION gr_transactions_to_cert() RETURNS int(11)
    DETERMINISTIC
BEGIN
  RETURN (select  performance_schema.replication_group_member_stats.COUNT_TRANSACTIONS_IN_QUEUE AS transactions_to_cert
    FROM
        performance_schema.replication_group_member_stats where MEMBER_ID=@@SERVER_UUID );
END$$

CREATE FUNCTION my_server_uuid() RETURNS TEXT(36) DETERMINISTIC NO SQL RETURN (SELECT @@global.server_uuid as my_id);$$

CREATE VIEW gr_member_routing_candidate_status AS
    SELECT 
        IFNULL((SELECT 
                        IF(MEMBER_STATE = 'ONLINE'
                                    AND ((SELECT 
                                        COUNT(*)
                                    FROM
                                        performance_schema.replication_group_members
                                    WHERE
                                        MEMBER_STATE != 'ONLINE') >= ((SELECT 
                                        COUNT(*)
                                    FROM
                                        performance_schema.replication_group_members) / 2) = 0),
                                'YES',
                                'NO')
                    FROM
                        performance_schema.replication_group_members
                            JOIN
                        performance_schema.replication_group_member_stats rgms USING (member_id)
                    WHERE
                        rgms.MEMBER_ID = my_server_uuid()),
                'NO') AS viable_candidate,
        IF((SELECT 
                    ((SELECT 
                                GROUP_CONCAT(performance_schema.global_variables.VARIABLE_VALUE
                                        SEPARATOR ',')
                            FROM
                                performance_schema.global_variables
                            WHERE
                                (performance_schema.global_variables.VARIABLE_NAME IN ('read_only' , 'super_read_only'))) <> 'OFF,OFF')
                ),
            'YES',
            'NO') AS read_only,
        IFNULL(sys.gr_applier_queue_length(), 0) AS transactions_behind,
        IFNULL(sys.gr_transactions_to_cert(), 0) AS transactions_to_cert;$$

DELIMITER ;
`

// BuildSecret generate secret environment variables for mysql pods
func BuildSecret(cr *rdsv1alpha1.Mysql) (secret *corev1.Secret) {
	var seeds string
	var semiSyncMasters string
	var semiSyncDoubleMaster bool
	var rootPassword []byte
	var initSQL string
	mysqlMaxConn := "300"
	if cr.Spec.MaxConn != nil {
		mysqlMaxConn = strconv.Itoa(*cr.Spec.MaxConn)
	}

	if cr.Spec.RootPassword != nil {
		rootPassword = []byte(util.Base64Decode(*cr.Spec.RootPassword))
	}

	for i := 0; i < int(*cr.Spec.Replicas); i++ {
		mysqlhost := cr.Name + "-mysql-" + strconv.Itoa(i) + ","
		seeds += mysqlhost

		if i == 0 {
			semiSyncMasters = mysqlhost
		}

		if cr.Spec.SemiSync != nil && cr.Spec.SemiSync.DoubleMasterHA && i == 1 {
			semiSyncDoubleMaster = true
			semiSyncMasters += mysqlhost
		}
	}
	seeds = strings.Trim(seeds, ",")
	semiSyncMasters = strings.Trim(semiSyncMasters, ",")

	secret = new(corev1.Secret)
	secret.APIVersion = "v1"
	secret.Kind = "Secret"
	secret.Name = cr.Name + "-mysql-secret"
	secret.Namespace = cr.Namespace
	secret.Labels = BuildMysqlLabels(cr)
	secret.Annotations = BuildMysqlAnnotaions(cr)

	secret.Data = map[string][]byte{
		"TZ":                   []byte(cr.Spec.TimeZone),
		"MYSQL_ROOT_PASSWORD":  rootPassword,
		"MYSQL_DATA_DIR":       []byte("/var/lib/mysql"),
		"MYSQL_CLUSTER_MODE":   []byte(string(cr.Spec.ClusterMode)),
		"MYSQL_CFG_MAX_CONN":   []byte(mysqlMaxConn),
		"MYSQL_CFG_WHITE_LIST": []byte(strings.Join(cr.Spec.Whitelist, ",")),
		"MYSQL_ROOT_HOST":      []byte("%"),
		"MYSQL_ADDRESSES":      []byte(seeds),
	}

	if cr.Spec.ClusterMode == rdsv1alpha1.ModeSemiSync {
		secret.Data["SEMI_SYNC_DOUBLE_MASTER_HA"] = []byte(strconv.FormatBool(semiSyncDoubleMaster))
		secret.Data["SEMI_SYNC_FIXED_MASTERS"] = []byte(semiSyncMasters)
	} else if cr.Spec.ClusterMode == rdsv1alpha1.ModeMGRSP {
		initSQL += mgrMonitorView
	} else if cr.Spec.ClusterMode == rdsv1alpha1.ModeMGRMP {
		initSQL += mgrMonitorView
	}

	if cr.Spec.ClusterUser != nil {
		mysqlPassword := []byte(util.Base64Decode(cr.Spec.ClusterUser.Password))
		initSQL += fmt.Sprintf(`
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
	}

	secret.Data["init.sql"] = []byte(initSQL)

	if cr.Spec.ExtraConfigDir != nil {
		secret.Data["MYSQL_CFG_EXTRA_DIR"] = []byte(*cr.Spec.ExtraConfigDir)
	}
	return secret
}
