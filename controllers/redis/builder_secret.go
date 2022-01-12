package redis

import (
	"strconv"
	"strings"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	hutil "github.com/hakur/util"
	corev1 "k8s.io/api/core/v1"
)

func buildSecret(cr *rdsv1alpha1.Redis) (secret *corev1.Secret) {
	var nodes []string
	var redisPassword string
	var allowEmptyPassword = "false"
	secret = new(corev1.Secret)
	svc := buildRedisSvc(cr)

	secret.APIVersion = "v1"
	secret.Kind = "ConfigMap"
	secret.Name = cr.Name + "-redis-secret"
	secret.Namespace = cr.Namespace
	secret.Labels = buildRedisLabels(cr)

	if cr.Spec.Password != nil {
		redisPassword = hutil.Base64Decode(*cr.Spec.Password)
	}

	if cr.Spec.Password == nil {
		allowEmptyPassword = "true"
	}

	for i := 0; i < caculateReplicas(cr); i++ {
		nodes = append(nodes, cr.Name+"-redis-"+strconv.Itoa(i)+"."+svc.Name)
	}

	secret.Data = map[string][]byte{
		"REDIS_PASSWORD":         []byte(redisPassword),
		"REDISCLI_AUTH":          []byte(redisPassword),
		"REDIS_NODES":            []byte(strings.Join(nodes, " ")),
		"ALLOW_EMPTY_PASSWORD":   []byte(allowEmptyPassword),
		"REDIS_CLUSTER_REPLICAS": []byte(strconv.Itoa(cr.Spec.DataReplicas)),
		"TZ":                     []byte(cr.Spec.TimeZone),
	}

	if cr.Spec.Redis.BackupMethod == "AOF" {
		secret.Data["REDIS_AOF_ENABLED"] = []byte("yes")
	} else {
		secret.Data["REDIS_AOF_ENABLED"] = []byte("no")
	}

	return
}
