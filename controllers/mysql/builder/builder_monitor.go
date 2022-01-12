package builder

import (
	"fmt"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	hutil "github.com/hakur/util"
	monitorv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func buildMysqlExporter(cr *rdsv1alpha1.Mysql) (container corev1.Container) {
	var image = "prom/mysqld-exporter:latest"
	var dsn = "user:password@(hostname:3306)/"
	if cr.Spec.Monitor.Image != "" {
		image = cr.Spec.Monitor.Image
	}

	dsn = fmt.Sprintf("%s:%s@(%s:%d)/",
		cr.Spec.Monitor.User.Username,
		hutil.Base64Decode(cr.Spec.Monitor.User.Password),
		"127.0.0.1",
		3306,
	)
	container.Image = image
	container.ImagePullPolicy = cr.Spec.ImagePullPolicy
	container.Name = "metrics"
	container.Env = []corev1.EnvVar{{Name: "DATA_SOURCE_NAME", Value: dsn}}
	container.Resources = cr.Spec.Monitor.Resources
	container.LivenessProbe = cr.Spec.Monitor.LivenessProbe
	container.ReadinessProbe = cr.Spec.Monitor.ReadinessProbe

	container.Ports = []corev1.ContainerPort{
		{Name: "metrics", ContainerPort: 9104},
	}

	container.Args = cr.Spec.Monitor.Args

	return
}

// func BuildServiceMonitor(cr *rdsv1alpha1.Mysql) (sm *monitorv1.ServiceMonitor) {
// 	sm = new(monitorv1.ServiceMonitor)

// 	sm.TypeMeta = metav1.TypeMeta{
// 		Kind:       "ServiceMonitor",
// 		APIVersion: "monitoring.coreos.com/v1",
// 	}

// 	labels := BuildMysqlLabels(cr)

// 	sm.ObjectMeta = metav1.ObjectMeta{
// 		Name:      cr.Name + "-mysql-sm",
// 		Namespace: cr.Namespace,
// 		Labels:    labels,
// 	}

// 	sm.Spec = monitorv1.ServiceMonitorSpec{
// 		Endpoints: []monitorv1.Endpoint{
// 			{
// 				Port:     "metrics",
// 				Interval: cr.Spec.Monitor.Interval,
// 			},
// 		},
// 		NamespaceSelector: monitorv1.NamespaceSelector{
// 			MatchNames: []string{cr.Namespace},
// 		},
// 		Selector: metav1.LabelSelector{
// 			MatchLabels: labels,
// 		},
// 	}

// 	return
// }

func BuildPodMonitor(cr *rdsv1alpha1.Mysql) (mon *monitorv1.PodMonitor) {
	mon = new(monitorv1.PodMonitor)

	mon.TypeMeta = metav1.TypeMeta{
		Kind:       "PodMonitor",
		APIVersion: "monitoring.coreos.com/v1",
	}

	labels := BuildMysqlLabels(cr)

	mon.ObjectMeta = metav1.ObjectMeta{
		Name:      cr.Name + "-mysql-pod-monitor",
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
