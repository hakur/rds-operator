package mysql

import (
	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/jinzhu/copier"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// buildProxySQLVolumeMounts generate pod volumeMouns
func buildProxySQLVolumeMounts() (data []corev1.VolumeMount) {
	data = append(data, corev1.VolumeMount{MountPath: "/var/lib/proxysql", Name: "data"})
	return
}

// buildProxySQLEnvs generate pod environments variables
func buildProxySQLEnvs(cr *rdsv1alpha1.Mysql) (data []corev1.EnvVar) {
	data = []corev1.EnvVar{
		{Name: "POD_IP", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIP"}}},
		{Name: "POD_NAME", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "metadata.name"}}},
	}

	return data
}

// buildProxySQLConfigContainer generate proxysql config render caontainer spec
func buildProxySQLConfigContainer(cr *rdsv1alpha1.Mysql) (container corev1.Container) {
	container.Image = cr.Spec.ConfigImage
	container.ImagePullPolicy = cr.Spec.ImagePullPolicy
	container.Name = "config"
	container.Env = buildProxySQLEnvs(cr)
	container.Env = append(container.Env, corev1.EnvVar{Name: "BOOTSTRAP_CLUSTER", Value: "false"})
	container.Env = append(container.Env, corev1.EnvVar{Name: "CONFIG_TYPE", Value: "proxysql"})
	container.EnvFrom = []corev1.EnvFromSource{{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: cr.Name + "-secret"}}}}
	container.VolumeMounts = buildProxySQLVolumeMounts()
	return container
}

// buildProxySQLContainer generate proxysql container spec
func buildProxySQLContainer(cr *rdsv1alpha1.Mysql) (container corev1.Container) {
	container.Image = cr.Spec.ProxySQL.Image
	container.ImagePullPolicy = cr.Spec.ImagePullPolicy
	container.Name = "proxysql"
	container.Env = buildProxySQLEnvs(cr)
	container.VolumeMounts = buildProxySQLVolumeMounts()
	container.EnvFrom = []corev1.EnvFromSource{{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: cr.Name + "-secret"}}}}
	container.Resources = cr.Spec.ProxySQL.Resources
	return container
}

// buildProxySQLSts generate proxysql statefulset spec
func buildProxySQLSts(cr *rdsv1alpha1.Mysql) (sts *appsv1.StatefulSet, err error) {
	var spec appsv1.StatefulSetSpec
	var podTemplateSpec corev1.PodTemplateSpec
	var dataVolumeClaim corev1.PersistentVolumeClaim
	var shareProcessNamespace = true

	sts = new(appsv1.StatefulSet)

	sts.ObjectMeta = metav1.ObjectMeta{
		Name:      cr.Name + "-proxysql",
		Namespace: cr.Namespace,
		Labels:    buildProxySQLLabels(cr),
	}

	sts.TypeMeta = metav1.TypeMeta{
		Kind:       "StatefulSet",
		APIVersion: "apps/v1",
	}

	spec.Replicas = cr.Spec.ProxySQL.Replicas
	spec.ServiceName = cr.Name
	spec.Selector = &metav1.LabelSelector{MatchLabels: buildProxySQLLabels(cr)}

	podTemplateSpec.ObjectMeta = metav1.ObjectMeta{Labels: buildProxySQLLabels(cr)}
	podTemplateSpec.Spec.ShareProcessNamespace = &shareProcessNamespace
	podTemplateSpec.Spec.InitContainers = []corev1.Container{buildProxySQLConfigContainer(cr)}
	podTemplateSpec.Spec.Containers = []corev1.Container{buildProxySQLContainer(cr)}
	podTemplateSpec.Spec.PriorityClassName = cr.Spec.PriorityClassName

	quantity, err := resource.ParseQuantity(cr.Spec.ProxySQL.StorageSize)
	if err != nil {
		return nil, err
	}

	dataVolumeClaim.ObjectMeta = metav1.ObjectMeta{Name: "data"}
	dataVolumeClaim.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"}
	dataVolumeClaim.Spec.StorageClassName = &cr.Spec.StorageClassName
	dataVolumeClaim.Spec.Resources = corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceStorage: quantity}}

	spec.Template = podTemplateSpec
	spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{dataVolumeClaim}
	sts.Spec = spec
	return sts, nil
}

// buildProxySQLService generate proxysql statefulset service
func buildProxySQLService(cr *rdsv1alpha1.Mysql) (svc *corev1.Service) {
	var spec corev1.ServiceSpec
	svc = new(corev1.Service)

	svc.ObjectMeta = metav1.ObjectMeta{
		Name:      cr.Name,
		Namespace: cr.Namespace,
		Labels:    buildProxySQLLabels(cr),
	}

	svc.TypeMeta = metav1.TypeMeta{
		Kind:       "StatefulSet",
		APIVersion: "apps/v1",
	}

	spec.Selector = buildProxySQLLabels(cr)

	spec.Ports = []corev1.ServicePort{
		{Name: "proxysql", Port: 6032},
	}

	if cr.Spec.ProxySQL.NodePort != nil {
		spec.Ports = append(spec.Ports, corev1.ServicePort{
			Name: "mysql", Port: 3306, NodePort: *cr.Spec.ProxySQL.NodePort,
		})
		spec.Type = corev1.ServiceTypeNodePort
	} else {
		spec.Ports = append(spec.Ports, corev1.ServicePort{
			Name: "mysql", Port: 3306,
		})
	}

	svc.Spec = spec
	return svc
}

// buildProxySQLLabels generate labels from cr resource, used for pod list filter
func buildProxySQLLabels(cr *rdsv1alpha1.Mysql) (labels map[string]string) {
	labels = map[string]string{
		"app":     "proxysql",
		"cr-name": cr.Name,
	}
	copier.Copy(labels, cr.Labels)
	return labels
}
