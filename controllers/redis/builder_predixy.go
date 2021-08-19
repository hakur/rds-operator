package redis

import (
	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/jinzhu/copier"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// buildPredixySvc gernate redis cluster Predixy kind:Service
func buildPredixySvc(cr *rdsv1alpha1.Redis) (svc *corev1.Service) {
	var spec corev1.ServiceSpec
	var redisPort = corev1.ServicePort{Name: "redis", Port: 6379}
	svc = new(corev1.Service)

	svc.ObjectMeta = metav1.ObjectMeta{
		Name:      cr.Name + "-predixy",
		Namespace: cr.Namespace,
		Labels:    buildPredixyLabels(cr),
	}

	svc.TypeMeta = metav1.TypeMeta{
		Kind:       "Service",
		APIVersion: "v1",
	}

	spec.Selector = buildPredixyLabels(cr)

	if cr.Spec.Predixy.NodePort != nil {
		spec.Type = corev1.ServiceTypeNodePort
		if *cr.Spec.Predixy.NodePort > 0 {
			redisPort.NodePort = *cr.Spec.Predixy.NodePort
		}
	}

	spec.Ports = []corev1.ServicePort{redisPort}

	svc.Spec = spec
	return svc
}

// buildPredixyLabels generate labels from cr resource, used for pod list filter
func buildPredixyLabels(cr *rdsv1alpha1.Redis) (labels map[string]string) {
	labels = map[string]string{
		"app":       "redis-cluster-predixy",
		"cr-name":   cr.Name,
		"api-group": rdsv1alpha1.GroupVersion.Group,
	}
	copier.Copy(labels, cr.Labels)
	return labels
}

// buildPredixyContainer generate Predixy container
func buildPredixyContainer(cr *rdsv1alpha1.Redis) (container corev1.Container) {
	secret := buildSecret(cr)
	container.Image = cr.Spec.Predixy.Image
	container.ImagePullPolicy = cr.Spec.ImagePullPolicy
	container.Name = "predixy"
	container.EnvFrom = []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secret.Name,
				},
			},
		},
	}

	container.Command = cr.Spec.Predixy.Command
	container.Args = cr.Spec.Predixy.Args
	container.Resources = cr.Spec.Resources
	container.LivenessProbe = cr.Spec.Predixy.LivenessProbe
	container.ReadinessProbe = cr.Spec.Predixy.ReadinessProbe
	container.VolumeMounts = []corev1.VolumeMount{
		{Name: "localtime", MountPath: "/etc/localtime"},
	}

	return container
}

// buildPredixyDeploy generate deployment of Predixy
func buildPredixyDeploy(cr *rdsv1alpha1.Redis) (deploy *appsv1.Deployment, err error) {
	var spec appsv1.DeploymentSpec
	var podTemplateSpec corev1.PodTemplateSpec
	deploy = new(appsv1.Deployment)

	deploy.ObjectMeta = metav1.ObjectMeta{
		Name:      cr.Name + "-predixy",
		Namespace: cr.Namespace,
		Labels:    buildPredixyLabels(cr),
	}

	deploy.TypeMeta = metav1.TypeMeta{
		Kind:       "Deployment",
		APIVersion: "apps/v1",
	}

	spec.Replicas = cr.Spec.Predixy.Replicas
	spec.Selector = &metav1.LabelSelector{MatchLabels: buildPredixyLabels(cr)}

	podTemplateSpec.ObjectMeta = metav1.ObjectMeta{Labels: buildPredixyLabels(cr)}
	podTemplateSpec.Spec.Containers = []corev1.Container{buildPredixyContainer(cr)}
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
