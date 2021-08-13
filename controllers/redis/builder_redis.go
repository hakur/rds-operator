package redis

import (
	"strconv"
	"strings"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/pkg/reconciler"
	"github.com/jinzhu/copier"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// buildRedisSts generate statefulset of redis servers
func buildRedisSts(cr *rdsv1alpha1.Redis) (sts *appsv1.StatefulSet, err error) {
	var spec appsv1.StatefulSetSpec
	var podTemplateSpec corev1.PodTemplateSpec
	var dataVolumeClaim corev1.PersistentVolumeClaim
	var replicas = int32(caculateReplicas(cr))

	sts = new(appsv1.StatefulSet)

	sts.ObjectMeta = metav1.ObjectMeta{
		Name:      cr.Name,
		Namespace: cr.Namespace,
		Labels:    buildRedisLabels(cr),
	}

	sts.TypeMeta = metav1.TypeMeta{
		Kind:       "StatefulSet",
		APIVersion: "apps/v1",
	}

	spec.Replicas = &replicas
	spec.ServiceName = cr.Name
	spec.Selector = &metav1.LabelSelector{MatchLabels: buildRedisLabels(cr)}
	spec.PodManagementPolicy = appsv1.ParallelPodManagement

	podTemplateSpec.ObjectMeta = metav1.ObjectMeta{Labels: buildRedisLabels(cr)}
	podTemplateSpec.Spec.Containers = []corev1.Container{buildRedisContainer(cr)}
	podTemplateSpec.Spec.ServiceAccountName = cr.Spec.ServiceAccountName
	podTemplateSpec.Spec.Affinity = cr.Spec.Affinity
	podTemplateSpec.Spec.Tolerations = cr.Spec.Tolerations
	podTemplateSpec.Spec.PriorityClassName = cr.Spec.PriorityClassName

	quantity, err := resource.ParseQuantity(cr.Spec.Redis.StorageSize)
	if err != nil {
		return nil, err
	}

	dataVolumeClaim.ObjectMeta = metav1.ObjectMeta{Name: "data", Labels: reconciler.BuildCRPVCLabels(cr, cr)}
	dataVolumeClaim.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"}
	dataVolumeClaim.Spec.StorageClassName = &cr.Spec.StorageClassName
	dataVolumeClaim.Spec.Resources = corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceStorage: quantity}}

	spec.Template = podTemplateSpec
	spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{dataVolumeClaim}
	sts.Spec = spec

	return
}

// buildRedisSvc gernate servers
func buildRedisSvc(cr *rdsv1alpha1.Redis) (svc *corev1.Service) {
	var spec corev1.ServiceSpec
	svc = new(corev1.Service)

	svc.ObjectMeta = metav1.ObjectMeta{
		Name:      cr.Name,
		Namespace: cr.Namespace,
		Labels:    buildRedisLabels(cr),
	}

	svc.TypeMeta = metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: "v1",
	}

	spec.Selector = buildRedisLabels(cr)

	spec.Ports = []corev1.ServicePort{
		{Name: "redis", Port: 6379},
	}

	svc.Spec = spec
	return svc
}

// buildRedisLabels generate labels from cr resource, used for pod list filter
func buildRedisLabels(cr *rdsv1alpha1.Redis) (labels map[string]string) {
	labels = map[string]string{
		"app":       "redis",
		"cr-name":   cr.Name,
		"api-group": rdsv1alpha1.GroupVersion.Group,
	}
	copier.Copy(labels, cr.Labels)
	return labels
}

// caculateReplicas cacaulate redis statefulset replicas by masterReplicas and dataReplicas
func caculateReplicas(cr *rdsv1alpha1.Redis) (replicas int) {
	replicas = cr.Spec.MasterReplicas + cr.Spec.MasterReplicas*cr.Spec.DataReplicas
	return replicas
}

// buildRedisContainer generate redis container spec
func buildRedisContainer(cr *rdsv1alpha1.Redis) (container corev1.Container) {
	var nodes []string
	var redisPassword string
	var allowEmptyPassword = "false"
	svc := buildRedisSvc(cr)

	if cr.Spec.Password != nil {
		redisPassword = *cr.Spec.Password
	}

	if cr.Spec.Password == nil {
		allowEmptyPassword = "true"
	}

	for i := 0; i < caculateReplicas(cr); i++ {
		nodes = append(nodes, cr.Name+"-"+strconv.Itoa(i)+"."+svc.Name)
	}

	container.Image = cr.Spec.Redis.Image
	container.ImagePullPolicy = cr.Spec.ImagePullPolicy
	container.Name = "redis"
	container.Env = []corev1.EnvVar{
		{Name: "REDIS_PASSWORD", Value: redisPassword},
		{Name: "REDISCLI_AUTH", Value: redisPassword},
		{Name: "REDIS_NODES", Value: strings.Join(nodes, " ")},
		{Name: "ALLOW_EMPTY_PASSWORD", Value: allowEmptyPassword},
		{Name: "REDIS_CLUSTER_REPLICAS", Value: strconv.Itoa(cr.Spec.DataReplicas)},
		{Name: "TZ", Value: cr.Spec.TimeZone},
	}

	if cr.Spec.Redis.BackupMethod == "AOF" {
		container.Env = append(container.Env, corev1.EnvVar{Name: "REDIS_AOF_ENABLED", Value: "yes"})
	} else {
		container.Env = append(container.Env, corev1.EnvVar{Name: "REDIS_AOF_ENABLED", Value: "no"})
	}

	container.VolumeMounts = []corev1.VolumeMount{
		{Name: "data", MountPath: "/bitnami"},
	}

	container.Command = cr.Spec.Redis.Command
	container.Args = cr.Spec.Redis.Args
	container.Resources = cr.Spec.Resources

	return container
}
