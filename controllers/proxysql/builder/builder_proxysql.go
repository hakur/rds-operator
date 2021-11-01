package builder

import (
	"encoding/base64"
	"strconv"
	"strings"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	"github.com/hakur/rds-operator/pkg/reconciler"
	hutil "github.com/hakur/util"
	"github.com/jinzhu/copier"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// https://github.com/sysown/proxysql/issues/2464#issuecomment-633361992
// https://proxysql.com/documentation/main-runtime/

type ProxySQLBuilder struct {
	CR *rdsv1alpha1.ProxySQL
}

// buildProxySQLVolumeMounts generate pod volumeMouns
func (t *ProxySQLBuilder) buildProxySQLVolumeMounts() (data []corev1.VolumeMount) {
	data = append(data, corev1.VolumeMount{MountPath: "/var/lib/proxysql", Name: "data"})
	data = append(data, corev1.VolumeMount{MountPath: "/etc/proxysql.cnf.d", Name: "cnf"})
	data = append(data, corev1.VolumeMount{MountPath: "/etc/localtime", Name: "localtime"})
	return
}

// buildProxySQLVolume generate pod volumeMouns
func (t *ProxySQLBuilder) buildProxySQLVolume() (data []corev1.Volume) {
	data = append(data, corev1.Volume{
		Name: "cnf",
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

	return
}

// buildProxySQLEnvs generate pod environments variables
func (t *ProxySQLBuilder) buildProxySQLEnvs() (data []corev1.EnvVar) {
	var adminCredentials string
	var maxWriters = 1

	for _, adminUser := range t.CR.Spec.AdminUsers {
		adminCredentials += adminUser.Username + ":" + hutil.Base64Decode(adminUser.Password) + ";"
	}
	adminCredentials = base64.StdEncoding.EncodeToString([]byte(strings.Trim(adminCredentials, ";")))

	if t.CR.Spec.ClusterMode == rdsv1alpha1.ModeMGRMP {
		maxWriters = 3
	}

	data = []corev1.EnvVar{
		{Name: "POD_IP", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIP"}}},
		{Name: "POD_NAME", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "metadata.name"}}},
		{Name: "PROXYSQL_MAX_WRITE_NODES", Value: strconv.Itoa(maxWriters)},
		{Name: "PROXYSQL_CLUSTER_USERNAME", Value: t.CR.Spec.ClusterUser.Username},
		{Name: "PROXYSQL_CLUSTER_PASSWORD", Value: t.CR.Spec.ClusterUser.Password},
		{Name: "MYSQL_MAX_CONNS", Value: strconv.Itoa(t.CR.Spec.MysqlMaxConn)},
		{Name: "ADMIN_CREDENTIALS", Value: adminCredentials},
		{Name: "MYSQL_MONITOR_USERNAME", Value: t.CR.Spec.MonitorUser.Username},
		{Name: "MYSQL_MONITOR_PASSWORD", Value: t.CR.Spec.MonitorUser.Password},
		{Name: "MYSQL_CLUSTER_MODE", Value: string(t.CR.Spec.ClusterMode)},
	}

	return data
}

// buildProxySQLConfigContainer generate proxysql config render caontainer spec
func (t *ProxySQLBuilder) buildProxySQLConfigContainer() (container corev1.Container) {
	container.Image = t.CR.Spec.ConfigImage
	container.ImagePullPolicy = t.CR.Spec.ImagePullPolicy
	container.Name = "init"
	container.Env = t.buildProxySQLEnvs()
	container.VolumeMounts = t.buildProxySQLVolumeMounts()
	container.Command = []string{"sidecar", "proxysql", "cfg"}
	return container
}

// buildProxySQLContainer generate proxysql container spec
func (t *ProxySQLBuilder) buildProxySQLContainer() (container corev1.Container) {

	container.Image = t.CR.Spec.Image
	container.ImagePullPolicy = t.CR.Spec.ImagePullPolicy
	container.Name = "proxysql"
	container.Env = t.buildProxySQLEnvs()
	container.VolumeMounts = t.buildProxySQLVolumeMounts()
	container.Command = []string{"proxysql"}
	container.Args = []string{"--foreground", "--idle-threads", "--datadir", "/var/lib/proxysql", "--config", "/etc/proxysql.cnf.d/proxysql.cnf"}

	container.Resources = t.CR.Spec.Resources
	container.LivenessProbe = t.CR.Spec.LivenessProbe
	container.ReadinessProbe = t.CR.Spec.ReadinessProbe
	return container
}

// BuildSts generate proxysql statefulset spec
func (t *ProxySQLBuilder) BuildSts() (sts *appsv1.StatefulSet, err error) {
	var spec appsv1.StatefulSetSpec
	var podTemplateSpec corev1.PodTemplateSpec
	var dataVolumeClaim corev1.PersistentVolumeClaim
	var shareProcessNamespace = true

	sts = new(appsv1.StatefulSet)

	sts.ObjectMeta = metav1.ObjectMeta{
		Name:        t.CR.Name + "-proxysql",
		Namespace:   t.CR.Namespace,
		Labels:      BuildProxySQLLabels(t.CR),
		Annotations: BuildProxySQLAnnotations(t.CR),
	}

	sts.TypeMeta = metav1.TypeMeta{
		Kind:       "StatefulSet",
		APIVersion: "apps/v1",
	}

	spec.Replicas = t.CR.Spec.Replicas
	spec.ServiceName = t.CR.Name + "-proxysql"
	spec.Selector = &metav1.LabelSelector{MatchLabels: BuildProxySQLLabels(t.CR)}

	podTemplateSpec.ObjectMeta = metav1.ObjectMeta{Labels: BuildProxySQLLabels(t.CR)}
	podTemplateSpec.Spec.ShareProcessNamespace = &shareProcessNamespace
	podTemplateSpec.Spec.InitContainers = []corev1.Container{t.buildProxySQLConfigContainer()}
	podTemplateSpec.Spec.Containers = []corev1.Container{t.buildProxySQLContainer()}
	podTemplateSpec.Spec.PriorityClassName = t.CR.Spec.PriorityClassName
	podTemplateSpec.Spec.Volumes = t.buildProxySQLVolume()
	podTemplateSpec.Spec.Affinity = t.CR.Spec.Affinity
	podTemplateSpec.Spec.Tolerations = t.CR.Spec.Tolerations

	quantity, err := resource.ParseQuantity(t.CR.Spec.StorageSize)
	if err != nil {
		return nil, err
	}

	dataVolumeClaim.ObjectMeta = metav1.ObjectMeta{Name: "data", Labels: reconciler.BuildCRPVCLabels(t.CR, t.CR)}
	dataVolumeClaim.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"}
	dataVolumeClaim.Spec.StorageClassName = &t.CR.Spec.StorageClassName
	dataVolumeClaim.Spec.Resources = corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceStorage: quantity}}

	spec.Template = podTemplateSpec
	spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{dataVolumeClaim}
	sts.Spec = spec
	return sts, nil
}

// BuildService generate proxysql statefulset service
func (t *ProxySQLBuilder) BuildService() (svc *corev1.Service) {
	var spec corev1.ServiceSpec
	svc = new(corev1.Service)

	svc.ObjectMeta = metav1.ObjectMeta{
		Name:        t.CR.Name + "-proxysql",
		Namespace:   t.CR.Namespace,
		Labels:      BuildProxySQLLabels(t.CR),
		Annotations: BuildProxySQLAnnotations(t.CR),
	}

	svc.TypeMeta = metav1.TypeMeta{
		Kind:       "StatefulSet",
		APIVersion: "apps/v1",
	}

	spec.Selector = BuildProxySQLLabels(t.CR)

	spec.Ports = []corev1.ServicePort{
		{Name: "proxysql", Port: 6032},
	}

	if t.CR.Spec.NodePort != nil {
		spec.Ports = append(spec.Ports, corev1.ServicePort{
			Name: "mysql", Port: 3306, NodePort: *t.CR.Spec.NodePort,
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
func BuildProxySQLLabels(cr *rdsv1alpha1.ProxySQL) (labels map[string]string) {
	labels = map[string]string{
		"app":       "proxysql",
		"cr-name":   cr.Name,
		"api-group": rdsv1alpha1.GroupVersion.Group,
	}
	copier.CopyWithOption(labels, cr.Labels, copier.Option{DeepCopy: true})
	return labels
}

func BuildProxySQLAnnotations(cr *rdsv1alpha1.ProxySQL) (annoations map[string]string) {
	annoations = map[string]string{}
	copier.CopyWithOption(annoations, cr.Annotations, copier.Option{DeepCopy: true})
	return annoations
}
