package redis

import (
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
		Name:        cr.Name + "-proxy",
		Namespace:   cr.Namespace,
		Labels:      buildProxyLabels(cr),
		Annotations: buildProxyAnnotations(cr),
	}

	svc.TypeMeta = metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: "v1",
	}

	spec.Selector = buildProxyLabels(cr)

	if cr.Spec.RedisClusterProxy.NodePort != nil {
		spec.Type = corev1.ServiceTypeNodePort
		if *cr.Spec.RedisClusterProxy.NodePort > 0 {
			redisPort.NodePort = *cr.Spec.RedisClusterProxy.NodePort
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
	copier.CopyWithOption(labels, cr.Labels, copier.Option{DeepCopy: true})
	return labels
}

func buildProxyAnnotations(cr *rdsv1alpha1.Redis) (annotations map[string]string) {
	annotations = map[string]string{}
	copier.CopyWithOption(annotations, cr.Annotations, copier.Option{DeepCopy: true})
	return annotations
}

// buildProxyContainer generate redis cluster proxy container
func buildProxyContainer(cr *rdsv1alpha1.Redis) (container corev1.Container) {
	secret := buildSecret(cr)

	container.Image = cr.Spec.RedisClusterProxy.Image
	container.ImagePullPolicy = cr.Spec.ImagePullPolicy
	container.Name = "proxy"
	container.EnvFrom = []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secret.Name,
				},
			},
		},
	}

	container.Command = cr.Spec.RedisClusterProxy.Command
	container.Args = cr.Spec.RedisClusterProxy.Args
	container.Resources = cr.Spec.Resources
	container.LivenessProbe = cr.Spec.RedisClusterProxy.LivenessProbe
	container.ReadinessProbe = cr.Spec.RedisClusterProxy.ReadinessProbe
	container.VolumeMounts = []corev1.VolumeMount{
		{Name: "localtime", MountPath: "/etc/localtime"},
	}

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

	spec.Replicas = cr.Spec.RedisClusterProxy.Replicas
	spec.Selector = &metav1.LabelSelector{MatchLabels: buildProxyLabels(cr)}

	podTemplateSpec.ObjectMeta = metav1.ObjectMeta{Labels: buildProxyLabels(cr)}
	podTemplateSpec.Spec.Containers = []corev1.Container{buildProxyContainer(cr)}
	podTemplateSpec.Spec.ServiceAccountName = cr.Spec.ServiceAccountName
	podTemplateSpec.Spec.Affinity = cr.Spec.Affinity
	podTemplateSpec.Spec.Tolerations = cr.Spec.Tolerations
	podTemplateSpec.Spec.PriorityClassName = cr.Spec.PriorityClassName
	podTemplateSpec.Spec.Volumes = []corev1.Volume{
		{Name: "localtime", VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: "/etc/localtime"}}},
	}

	spec.Template = podTemplateSpec
	deploy.Spec = spec

	return deploy, nil
}
