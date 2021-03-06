package builder

import (
	"fmt"
	"strconv"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/pkg/reconciler"
	"github.com/jinzhu/copier"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MysqlBuilder struct {
	CR *rdsv1alpha1.Mysql
}

// BuildMyCnfCM generate mysql my.cnf configmap
func (t *MysqlBuilder) BuildMyCnfCM(cr *rdsv1alpha1.Mysql) (cm *corev1.ConfigMap) {
	cnfDir := "/etc/my.cnf.d/"
	cm = new(corev1.ConfigMap)
	cm.APIVersion = "v1"
	cm.Kind = "ConfigMap"
	cm.Name = cr.Name + "-mycnf"
	cm.Namespace = cr.Namespace
	cm.Labels = BuildMysqlLabels(t.CR)
	cm.Annotations = BuildMysqlAnnotaions(t.CR)

	if cr.Spec.ExtraConfigDir != nil {
		cnfDir = *cr.Spec.ExtraConfigDir
	}

	cnfContent := `
[client]
[mysqld]

!includedir ` + cnfDir
	cm.Data = map[string]string{
		"my.cnf":       cnfContent,
		"extra_config": t.CR.Spec.ExtraConfig,
	}
	return
}

// buildMysqlVolumeMounts generate pod volumeMouns
func (t *MysqlBuilder) buildMysqlVolumeMounts() (data []corev1.VolumeMount) {
	data = append(data, corev1.VolumeMount{Name: "mysql-sock", MountPath: "/var/run/mysqld"})
	data = append(data, corev1.VolumeMount{Name: "my-cnf", MountPath: "/etc/my.cnf", SubPath: "my.cnf"})
	data = append(data, corev1.VolumeMount{MountPath: "/etc/my.cnf.d", Name: "my-cnfd"})
	data = append(data, corev1.VolumeMount{MountPath: "/var/lib/mysql", Name: "data"})
	data = append(data, corev1.VolumeMount{MountPath: "/etc/localtime", Name: "localtime"})
	data = append(data, corev1.VolumeMount{MountPath: "/docker-entrypoint-initdb.d", Name: "init-sql"})
	return
}

// buildMysqlVolumes generate pod volumes
func (t *MysqlBuilder) buildMysqlVolumes(cr *rdsv1alpha1.Mysql) (data []corev1.Volume) {
	var mysqlConfigVolumeMode int32 = 0755
	secret := BuildSecret(cr)

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

	data = append(data, corev1.Volume{
		Name: "localtime",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: "/etc/localtime",
			},
		},
	})

	data = append(data, corev1.Volume{
		Name: "init-sql",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secret.Name,
				Items: []corev1.KeyToPath{
					{Key: "init.sql", Path: "init.sql"},
				},
			},
		},
	})

	return
}

// buildMysqlEnvs generate pod environments variables
func (t *MysqlBuilder) buildMysqlEnvs(cr *rdsv1alpha1.Mysql) (data []corev1.EnvVar) {
	data = []corev1.EnvVar{
		{Name: "POD_IP", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIP"}}},
		{Name: "POD_NAME", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "metadata.name"}}},
	}

	return data
}

// buildMysqlContainer generate mysql container spec
func (t *MysqlBuilder) buildMysqlContainer(cr *rdsv1alpha1.Mysql) (container corev1.Container) {
	secret := BuildSecret(cr)

	container.Image = cr.Spec.Image
	container.ImagePullPolicy = cr.Spec.ImagePullPolicy
	container.Name = "mysql"
	container.Env = t.buildMysqlEnvs(cr)
	container.VolumeMounts = t.buildMysqlVolumeMounts()
	container.EnvFrom = []corev1.EnvFromSource{{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name}}}}
	container.Resources = cr.Spec.Resources
	container.LivenessProbe = cr.Spec.LivenessProbe
	container.ReadinessProbe = cr.Spec.ReadinessProbe

	return container
}

// buildMysqlInitContainer generate mysql config render caontainer spec
func (t *MysqlBuilder) buildMysqlInitContainer(cr *rdsv1alpha1.Mysql) (container corev1.Container) {
	secret := BuildSecret(cr)

	container.Image = cr.Spec.ConfigImage
	container.ImagePullPolicy = cr.Spec.ImagePullPolicy
	container.Name = "init"
	container.Env = t.buildMysqlEnvs(cr)
	container.EnvFrom = []corev1.EnvFromSource{{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name}}}}
	container.VolumeMounts = t.buildMysqlVolumeMounts()
	container.Command = []string{"sidecar", "mysql", "cfg"}
	return container
}

// BuildSts generate mysql statefulset
func (t *MysqlBuilder) BuildSts() (sts *appsv1.StatefulSet, err error) {
	var spec appsv1.StatefulSetSpec
	var podTemplateSpec corev1.PodTemplateSpec
	var mysqlDataVolumeClaim corev1.PersistentVolumeClaim
	var shareProcessNamespace = true

	if t.CR.Spec.ClusterMode == rdsv1alpha1.ModeSemiSync && t.CR.Spec.SemiSync.DoubleMasterHA && t.CR.Spec.ClusterUser == nil {
		return nil, fmt.Errorf("cluster enabled but cluster user not configured. CR resource spec.clusterUser is nil")
	}

	sts = new(appsv1.StatefulSet)

	sts.ObjectMeta = metav1.ObjectMeta{
		Name:        t.CR.Name + "-mysql",
		Namespace:   t.CR.Namespace,
		Labels:      BuildMysqlLabels(t.CR),
		Annotations: BuildMysqlAnnotaions(t.CR),
	}

	sts.TypeMeta = metav1.TypeMeta{
		Kind:       "StatefulSet",
		APIVersion: "apps/v1",
	}

	spec.Replicas = t.CR.Spec.Replicas
	spec.ServiceName = t.CR.Name + "-mysql"
	spec.Selector = &metav1.LabelSelector{MatchLabels: BuildMysqlLabels(t.CR)}

	podTemplateSpec.ObjectMeta = metav1.ObjectMeta{Labels: BuildMysqlLabels(t.CR)}
	podTemplateSpec.Spec.Volumes = t.buildMysqlVolumes(t.CR)
	podTemplateSpec.Spec.ShareProcessNamespace = &shareProcessNamespace
	podTemplateSpec.Spec.InitContainers = []corev1.Container{t.buildMysqlInitContainer(t.CR)}
	podTemplateSpec.Spec.Containers = []corev1.Container{t.buildMysqlContainer(t.CR)}
	podTemplateSpec.Spec.PriorityClassName = t.CR.Spec.PriorityClassName
	podTemplateSpec.Spec.Affinity = t.CR.Spec.Affinity
	podTemplateSpec.Spec.Tolerations = t.CR.Spec.Tolerations

	if t.CR.Spec.Monitor != nil {
		podTemplateSpec.Spec.Containers = append(podTemplateSpec.Spec.Containers, buildMysqlExporter(t.CR))
	}

	quantity, err := resource.ParseQuantity(t.CR.Spec.StorageSize)
	if err != nil {
		return nil, err
	}

	mysqlDataVolumeClaim.ObjectMeta = metav1.ObjectMeta{Name: "data", Labels: reconciler.BuildCRPVCLabels(t.CR, t.CR), Annotations: BuildMysqlAnnotaions(t.CR)} // use labels for gc , gc date is annotation types.PVCDeleteDateAnnotationName
	mysqlDataVolumeClaim.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"}
	mysqlDataVolumeClaim.Spec.StorageClassName = &t.CR.Spec.StorageClassName
	mysqlDataVolumeClaim.Spec.Resources = corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceStorage: quantity}}

	spec.Template = podTemplateSpec
	spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{mysqlDataVolumeClaim}
	sts.Spec = spec
	return sts, nil
}

// BuildService generate mysql services
func (t *MysqlBuilder) BuildService(cr *rdsv1alpha1.Mysql) (svc *corev1.Service) {
	var spec corev1.ServiceSpec
	svc = new(corev1.Service)

	svc.ObjectMeta = metav1.ObjectMeta{
		Name:        cr.Name + "-mysql",
		Namespace:   cr.Namespace,
		Labels:      BuildMysqlLabels(t.CR),
		Annotations: BuildMysqlAnnotaions(t.CR),
	}

	svc.TypeMeta = metav1.TypeMeta{
		Kind:       "StatefulSet",
		APIVersion: "apps/v1",
	}

	spec.Selector = BuildMysqlLabels(t.CR)

	spec.Ports = []corev1.ServicePort{
		{Name: "metrics", Port: 9104},
		{Name: "mysql", Port: 3306},
		{Name: "mysql-mgr", Port: 33061},
		{Name: "galera-replication", Port: 4444},
		{Name: "galera-peers", Port: 4567},
	}

	svc.Spec = spec
	return svc
}

// BuildContainerServices generate mysql services for each mysql container
func (t *MysqlBuilder) BuildContainerServices(cr *rdsv1alpha1.Mysql) (services []*corev1.Service) {
	for i := 0; i < int(*cr.Spec.Replicas); i++ {
		var spec corev1.ServiceSpec
		svc := new(corev1.Service)

		svc.ObjectMeta = metav1.ObjectMeta{
			Name:      cr.Name + "-mysql-" + strconv.Itoa(i),
			Namespace: cr.Namespace,
			Labels:    BuildMysqlLabels(t.CR),
		}

		svc.TypeMeta = metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		}

		spec.Selector = BuildMysqlLabels(t.CR)
		spec.Selector["statefulset.kubernetes.io/pod-name"] = cr.Name + "-mysql-" + strconv.Itoa(i)

		spec.Ports = []corev1.ServicePort{
			{Name: "metrics", Port: 9104},
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

// BuildMysqlLabels generate labels from cr resource, used for pod list filter
func BuildMysqlLabels(cr *rdsv1alpha1.Mysql) (labels map[string]string) {
	labels = map[string]string{
		"app":       "mysql",
		"cr-name":   cr.Name,
		"api-group": rdsv1alpha1.GroupVersion.Group,
	}
	copier.CopyWithOption(labels, cr.Labels, copier.Option{DeepCopy: true})
	return labels
}

// BuildMysqlAnnotaions generate annoations from cr resource, used for pod list filter
func BuildMysqlAnnotaions(cr *rdsv1alpha1.Mysql) (annotations map[string]string) {
	annotations = map[string]string{}
	copier.CopyWithOption(annotations, cr.Annotations, copier.Option{DeepCopy: true})
	return annotations
}
