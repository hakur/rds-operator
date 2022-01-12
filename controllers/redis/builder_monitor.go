package redis

import (
	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func buildRedisExporter(cr *rdsv1alpha1.Redis) (container corev1.Container) {
	var image = "oliver006/redis_exporter:latest"
	if cr.Spec.Monitor.Image != "" {
		image = cr.Spec.Monitor.Image
	}
	secret := buildSecret(cr)

	container.Image = image
	container.ImagePullPolicy = cr.Spec.ImagePullPolicy
	container.Name = "metrics"
	container.EnvFrom = []corev1.EnvFromSource{
		{
			SecretRef: &corev1.SecretEnvSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: secret.Name,
				},
			},
		},
	}
	container.Resources = cr.Spec.Monitor.Resources
	container.LivenessProbe = cr.Spec.Monitor.LivenessProbe
	container.ReadinessProbe = cr.Spec.Monitor.ReadinessProbe

	container.Ports = []corev1.ContainerPort{
		{Name: "metrics", ContainerPort: 9121},
	}

	container.Args = cr.Spec.Monitor.Args

	return
}

func BuildPodMonitor(cr *rdsv1alpha1.Redis) (mon *monitorv1.PodMonitor) {
	mon = new(monitorv1.PodMonitor)

	mon.TypeMeta = metav1.TypeMeta{
		Kind:       "PodMonitor",
		APIVersion: "monitoring.coreos.com/v1",
	}

	labels := buildRedisLabels(cr)

	mon.ObjectMeta = metav1.ObjectMeta{
		Name:      cr.Name + "-redis-pod-monitor",
		Namespace: cr.Namespace,
		Labels:    labels,
	}

	mon.Spec = monitorv1.PodMonitorSpec{
		PodMetricsEndpoints: []monitorv1.PodMetricsEndpoint{
			{
				Port:     "metrics",
				Interval: cr.Spec.Monitor.Interval,
			},
		},
		NamespaceSelector: monitorv1.NamespaceSelector{
			MatchNames: []string{cr.Namespace},
		},
		Selector: metav1.LabelSelector{
			MatchLabels: labels,
		},
	}

	return
}
