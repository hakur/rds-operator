package mysql

import (
	"strconv"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/pkg/reconciler"
	"github.com/jinzhu/copier"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// buildMyCnfCM generate mysql my.cnf configmap
func buildMyCnfCM(cr *rdsv1alpha1.Mysql) (cm *corev1.ConfigMap) {
	cnfDir := "/etc/my.cnf.d/"
	cm = new(corev1.ConfigMap)
	cm.APIVersion = "v1"
	cm.Kind = "ConfigMap"
	cm.Name = cr.Name + "-mycnf"
	cm.Namespace = cr.Namespace
	cm.Labels = buildMysqlLabels(cr)

	if cr.Spec.Mysql.ExtraConfigDir != nil {
		cnfDir = *cr.Spec.Mysql.ExtraConfigDir
	}

	cnfContent := `
[mysqld]
!includedir ` + cnfDir
	cm.Data = map[string]string{
		"my.cnf": cnfContent,
	}
	return
}

// buildMysqlVolumeMounts generate pod volumeMouns
func buildMysqlVolumeMounts() (data []corev1.VolumeMount) {
	data = append(data, corev1.VolumeMount{Name: "mysql-sock", MountPath: "/var/run/mysqld"})
	data = append(data, corev1.VolumeMount{Name: "my-cnf", MountPath: "/etc/my.cnf", SubPath: "my.cnf"})
	data = append(data, corev1.VolumeMount{MountPath: "/etc/my.cnf.d", Name: "my-cnfd"})
	data = append(data, corev1.VolumeMount{MountPath: "/var/lib/mysql", Name: "data"})
	return
}

// buildMysqlVolumes generate pod volumes
func buildMysqlVolumes(cr *rdsv1alpha1.Mysql) (data []corev1.Volume) {
	var mysqlConfigVolumeMode int32 = 0755

	data = append(data, corev1.Volume{
		Name: "mysql-sock",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				SizeLimit: &resource.Quantity{
					Format: resource.Format("32KiB"),
				},
			},
		},
	})

	data = append(data, corev1.Volume{Name: "my-cnf", VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{
		Items:                []corev1.KeyToPath{{Key: "my.cnf", Path: "my.cnf"}},
		LocalObjectReference: corev1.LocalObjectReference{Name: cr.Name + "-mycnf"},
		DefaultMode:          &mysqlConfigVolumeMode,
	}}})

	data = append(data, corev1.Volume{
		Name: "my-cnfd",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{
				SizeLimit: &resource.Quantity{
					Format: resource.Format("32KiB"),
				},
			},
		},
	})

	return
}

// buildMysqlEnvs generate pod environments variables
func buildMysqlEnvs(cr *rdsv1alpha1.Mysql) (data []corev1.EnvVar) {
	data = []corev1.EnvVar{
		{Name: "POD_IP", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIP"}}},
		{Name: "POD_NAME", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "metadata.name"}}},
	}

	return data
}

// buildMysqlContainer generate mysql container spec
func buildMysqlContainer(cr *rdsv1alpha1.Mysql) (container corev1.Container) {
	container.Image = cr.Spec.Mysql.Image
	container.ImagePullPolicy = cr.Spec.ImagePullPolicy
	container.Name = "mysql"
	container.Env = buildMysqlEnvs(cr)
	container.VolumeMounts = buildMysqlVolumeMounts()
	container.EnvFrom = []corev1.EnvFromSource{{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: cr.Name + "-secret"}}}}
	container.Resources = cr.Spec.Mysql.Resources
	return container
}

// buildMysqlBootContainer generate mysql container spec
func buildMysqlBootContainer(cr *rdsv1alpha1.Mysql) (container corev1.Container) {
	container.Image = cr.Spec.ConfigImage
	container.ImagePullPolicy = cr.Spec.ImagePullPolicy
	container.Name = "boot"
	container.Env = buildMysqlEnvs(cr)
	container.EnvFrom = []corev1.EnvFromSource{{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: cr.Name + "-secret"}}}}
	container.VolumeMounts = buildMysqlVolumeMounts()
	container.Resources = cr.Spec.Mysql.Resources
	container.Env = append(container.Env, corev1.EnvVar{Name: "BOOTSTRAP_CLUSTER", Value: "true"})
	return container
}

// buildMysqlConfigContainer generate mysql config render caontainer spec
func buildMysqlConfigContainer(cr *rdsv1alpha1.Mysql) (container corev1.Container) {
	container.Image = cr.Spec.ConfigImage
	container.ImagePullPolicy = cr.Spec.ImagePullPolicy
	container.Name = "config"
	container.Env = buildMysqlEnvs(cr)
	container.EnvFrom = []corev1.EnvFromSource{{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: cr.Name + "-secret"}}}}
	container.VolumeMounts = buildMysqlVolumeMounts()
	return container
}

// buildMysqlSts generate mysql statefulset
func buildMysqlSts(cr *rdsv1alpha1.Mysql) (sts *appsv1.StatefulSet, err error) {
	var spec appsv1.StatefulSetSpec
	var podTemplateSpec corev1.PodTemplateSpec
	var mysqlDataVolumeClaim corev1.PersistentVolumeClaim
	var shareProcessNamespace = true

	sts = new(appsv1.StatefulSet)

	sts.ObjectMeta = metav1.ObjectMeta{
		Name:      cr.Name,
		Namespace: cr.Namespace,
		Labels:    buildMysqlLabels(cr),
	}

	sts.TypeMeta = metav1.TypeMeta{
		Kind:       "StatefulSet",
		APIVersion: "apps/v1",
	}

	spec.Replicas = cr.Spec.Mysql.Replicas
	spec.ServiceName = cr.Name
	spec.Selector = &metav1.LabelSelector{MatchLabels: buildMysqlLabels(cr)}

	podTemplateSpec.ObjectMeta = metav1.ObjectMeta{Labels: buildMysqlLabels(cr)}
	podTemplateSpec.Spec.Volumes = buildMysqlVolumes(cr)
	podTemplateSpec.Spec.ShareProcessNamespace = &shareProcessNamespace
	podTemplateSpec.Spec.InitContainers = []corev1.Container{buildMysqlConfigContainer(cr)}
	podTemplateSpec.Spec.Containers = []corev1.Container{buildMysqlContainer(cr), buildMysqlBootContainer(cr)}
	podTemplateSpec.Spec.PriorityClassName = cr.Spec.PriorityClassName

	quantity, err := resource.ParseQuantity(cr.Spec.Mysql.StorageSize)
	if err != nil {
		return nil, err
	}

	mysqlDataVolumeClaim.ObjectMeta = metav1.ObjectMeta{Name: "data", Labels: reconciler.BuildCRPVCLabels(cr.Name, cr.GroupVersionKind().String())} // use labels for gc , gc date is annotation types.PVCDeleteDateAnnotationName
	mysqlDataVolumeClaim.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"}
	mysqlDataVolumeClaim.Spec.StorageClassName = &cr.Spec.StorageClassName
	mysqlDataVolumeClaim.Spec.Resources = corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceStorage: quantity}}

	spec.Template = podTemplateSpec
	spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{mysqlDataVolumeClaim}
	sts.Spec = spec
	return sts, nil
}

// buildMysqlService generate mysql services
func buildMysqlService(cr *rdsv1alpha1.Mysql) (svc *corev1.Service) {
	var spec corev1.ServiceSpec
	svc = new(corev1.Service)

	svc.ObjectMeta = metav1.ObjectMeta{
		Name:      cr.Name,
		Namespace: cr.Namespace,
		Labels:    buildMysqlLabels(cr),
	}

	svc.TypeMeta = metav1.TypeMeta{
		Kind:       "StatefulSet",
		APIVersion: "apps/v1",
	}

	spec.Selector = buildMysqlLabels(cr)

	spec.Ports = []corev1.ServicePort{
		{Name: "mysql", Port: 3306},
		{Name: "mysql-mgr", Port: 33061},
		{Name: "galera-replication", Port: 4444},
		{Name: "galera-peers", Port: 4567},
	}

	svc.Spec = spec
	return svc
}

// buildMysqlContainerServices generate mysql services for each mysql container
func buildMysqlContainerServices(cr *rdsv1alpha1.Mysql) (services []*corev1.Service) {
	for i := 0; i < int(*cr.Spec.Mysql.Replicas); i++ {
		var spec corev1.ServiceSpec
		svc := new(corev1.Service)

		svc.ObjectMeta = metav1.ObjectMeta{
			Name:      cr.Name + "-" + strconv.Itoa(i),
			Namespace: cr.Namespace,
			Labels:    buildMysqlLabels(cr),
		}

		svc.TypeMeta = metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		}

		spec.Selector = buildMysqlLabels(cr)
		spec.Selector["statefulset.kubernetes.io/pod-name"] = cr.Name + "-" + strconv.Itoa(i)

		spec.Ports = []corev1.ServicePort{
			{Name: "mysql", Port: 3306},
			{Name: "mysql-mgr", Port: 33061},
			{Name: "galera-replication", Port: 4444},
			{Name: "galera-peers", Port: 4567},
		}

		svc.Spec = spec
		services = append(services, svc)
	}
	return services
}

// buildMysqlLabels generate labels from cr resource, used for pod list filter
func buildMysqlLabels(cr *rdsv1alpha1.Mysql) (labels map[string]string) {
	labels = map[string]string{
		"app":       "mysql",
		"cr-name":   cr.Name,
		"api-group": rdsv1alpha1.GroupVersion.Group,
	}
	copier.Copy(labels, cr.Labels)
	return labels
}
