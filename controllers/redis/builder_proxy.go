package redis

import (
	"strconv"
	"strings"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/jinzhu/copier"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// buildProxySvc gernate redis cluster proxy kind:Service
func buildProxySvc(cr *rdsv1alpha1.Redis) (svc *corev1.Service) {
	var spec corev1.ServiceSpec
	var redisPort = corev1.ServicePort{Name: "redis", Port: 6379}
	svc = new(corev1.Service)

	svc.ObjectMeta = metav1.ObjectMeta{
		Name:      cr.Name + "-proxy",
		Namespace: cr.Namespace,
		Labels:    buildProxyLabels(cr),
	}

	svc.TypeMeta = metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: "v1",
	}

	spec.Selector = buildProxyLabels(cr)

	if cr.Spec.Proxy.NodePort != nil {
		spec.Type = corev1.ServiceTypeNodePort
		if *cr.Spec.Proxy.NodePort > 0 {
			redisPort.NodePort = *cr.Spec.Proxy.NodePort
		}
	}

	spec.Ports = []corev1.ServicePort{redisPort}

	svc.Spec = spec
	return svc
}

// buildProxyLabels generate labels from cr resource, used for pod list filter
func buildProxyLabels(cr *rdsv1alpha1.Redis) (labels map[string]string) {
	labels = map[string]string{
		"app":       "redis-cluster-proxy",
		"cr-name":   cr.Name,
		"api-group": rdsv1alpha1.GroupVersion.Group,
	}
	copier.Copy(labels, cr.Labels)
	return labels
}

// buildPorxyContainer generate redis cluster proxy container
func buildPorxyContainer(cr *rdsv1alpha1.Redis) (container corev1.Container) {
	var nodes []string
	var allowEmptyPassword = "false"
	var redisPassword string
	svc := buildRedisSvc(cr)

	if cr.Spec.Password != nil {
		redisPassword = *cr.Spec.Password
	}

	if cr.Spec.Password == nil {
		allowEmptyPassword = "true"
	}

	for i := 0; i < caculateReplicas(cr); i++ {
		nodes = append(nodes, cr.Name+"-redis-"+strconv.Itoa(i)+"."+svc.Name+":6379")
	}

	container.Image = cr.Spec.Proxy.Image
	container.ImagePullPolicy = cr.Spec.ImagePullPolicy
	container.Name = "proxy"
	container.Env = []corev1.EnvVar{
		{Name: "REDIS_PASSWORD", Value: redisPassword},
		{Name: "REDISCLI_AUTH", Value: redisPassword},
		{Name: "REDIS_NODES", Value: strings.Join(nodes, " ")},
		{Name: "ALLOW_EMPTY_PASSWORD", Value: allowEmptyPassword},
		{Name: "REDIS_CLUSTER_REPLICAS", Value: strconv.Itoa(cr.Spec.DataReplicas)},
		{Name: "TZ", Value: cr.Spec.TimeZone},
	}
	container.Command = cr.Spec.Proxy.Command
	container.Args = cr.Spec.Proxy.Args
	container.Resources = cr.Spec.Resources
	container.LivenessProbe = cr.Spec.Proxy.LivenessProbe
	container.ReadinessProbe = cr.Spec.Proxy.ReadinessProbe

	return container
}

// buildProxyDeploy generate deployment of redis cluster proxy
func buildProxyDeploy(cr *rdsv1alpha1.Redis) (deploy *appsv1.Deployment, err error) {
	var spec appsv1.DeploymentSpec
	var podTemplateSpec corev1.PodTemplateSpec
	deploy = new(appsv1.Deployment)

	deploy.ObjectMeta = metav1.ObjectMeta{
		Name:      cr.Name + "-proxy",
		Namespace: cr.Namespace,
		Labels:    buildProxyLabels(cr),
	}

	deploy.TypeMeta = metav1.TypeMeta{
		Kind:       "Deployment",
		APIVersion: "apps/v1",
	}

	spec.Replicas = cr.Spec.Proxy.Replicas
	spec.Selector = &metav1.LabelSelector{MatchLabels: buildProxyLabels(cr)}

	podTemplateSpec.ObjectMeta = metav1.ObjectMeta{Labels: buildProxyLabels(cr)}
	podTemplateSpec.Spec.Containers = []corev1.Container{buildPorxyContainer(cr)}
	podTemplateSpec.Spec.ServiceAccountName = cr.Spec.ServiceAccountName
	podTemplateSpec.Spec.Affinity = cr.Spec.Affinity
	podTemplateSpec.Spec.Tolerations = cr.Spec.Tolerations
	podTemplateSpec.Spec.PriorityClassName = cr.Spec.PriorityClassName

	spec.Template = podTemplateSpec
	deploy.Spec = spec

	return deploy, nil
}
