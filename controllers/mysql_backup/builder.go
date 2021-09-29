package mysqlbackup

import (
	"strings"

	rdsv1alpha1 "github.com/hakur/rds-operator/apis/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CronJobBuilder struct {
	CR *rdsv1alpha1.MysqlBackup
}

func BuildSecret(cr *rdsv1alpha1.MysqlBackup) (secret *corev1.Secret) {
	var hosts []string

	for _, v := range cr.Spec.Hosts {
		hosts = append(hosts, v.Host)
	}

	secret = new(corev1.Secret)
	secret.ObjectMeta = metav1.ObjectMeta{
		Name:      cr.Name + "-backup-secret",
		Namespace: cr.Namespace,
		Labels:    BuildLabels(cr),
	}

	secret.Data = make(map[string][]byte)
	secret.Data["MYSQL_USER"] = []byte(cr.Spec.Username)
	secret.Data["MYSQL_PASSWORD"] = []byte(cr.Spec.Password)
	secret.Data["MYSQL_NODES"] = []byte(strings.Join(hosts, " "))
	secret.Data["MYSQL_PORT"] = []byte("3306")
	secret.Data["MYSQL_CLUSTER_MODE"] = []byte(cr.Spec.ClusterMode)
	secret.Data["S3_ENDPOINT"] = []byte(cr.Spec.S3.Endpoint)
	secret.Data["S3_BUCKET"] = []byte(cr.Spec.S3.Bucket)
	secret.Data["S3_PATH"] = []byte(strings.Trim(cr.Spec.S3.Path, "/"))
	secret.Data["S3_ACCESS_KEY"] = []byte(cr.Spec.S3.AccessKey)
	secret.Data["S3_SECURITY_KEY"] = []byte(cr.Spec.S3.SecurityKey)

	if cr.Spec.UseZlibCompress != nil && *cr.Spec.UseZlibCompress {
		secret.Data["MYSQL_BACKUP_USE_ZLIB_COMPRESS"] = []byte("true")
	} else {
		secret.Data["MYSQL_BACKUP_USE_ZLIB_COMPRESS"] = []byte("false")
	}

	return
}

func (t *CronJobBuilder) BuildCronJob() (job *batchv1.CronJob, err error) {
	job = new(batchv1.CronJob)
	job.APIVersion = "batch/v1"
	job.Kind = "CronJob"

	job.ObjectMeta = metav1.ObjectMeta{
		Name:      t.CR.Name + "-mysqlbackup",
		Namespace: t.CR.Namespace,
		Labels:    BuildLabels(t.CR),
	}

	jobSpec, err := t.buildJobSpec()
	if err != nil {
		return nil, err
	}

	job.Spec = batchv1.CronJobSpec{
		JobTemplate: batchv1.JobTemplateSpec{
			Spec: jobSpec,
		},
		Schedule: t.CR.Spec.Schedule,
	}
	return
}

func (t *CronJobBuilder) buildJobSpec() (spec batchv1.JobSpec, err error) {
	volumes, err := t.buildVolume()
	var parallel int32 = 1
	var ttlSeconds int32 = 300
	spec = batchv1.JobSpec{
		Parallelism:             &parallel,
		TTLSecondsAfterFinished: &ttlSeconds,
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				InitContainers: []corev1.Container{t.buildInitContainer()},
				Containers:     []corev1.Container{t.buildMainContainer()},
				Volumes:        volumes,
				RestartPolicy:  "OnFailure",
			},
		},
	}
	return
}

func (t *CronJobBuilder) buildInitContainer() (container corev1.Container) {
	secret := BuildSecret(t.CR)
	container = corev1.Container{
		Name:    "backup",
		Image:   t.CR.Spec.InitContainerImage,
		Command: t.CR.Spec.InitContainerCommand,
		Args:    t.CR.Spec.InitContainerArgs,
		EnvFrom: []corev1.EnvFromSource{
			{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name},
				},
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: "data", MountPath: "/data"},
		},
		Resources: t.CR.Spec.Resources,
	}
	return
}

func (t *CronJobBuilder) buildMainContainer() (container corev1.Container) {
	secret := BuildSecret(t.CR)
	container = corev1.Container{
		Name:    "upload",
		Image:   t.CR.Spec.Image,
		Command: t.CR.Spec.Command,
		Args:    t.CR.Spec.Args,
		EnvFrom: []corev1.EnvFromSource{
			{
				SecretRef: &corev1.SecretEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{Name: secret.Name},
				},
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: "data", MountPath: "/data"},
		},
		Resources: t.CR.Spec.Resources,
	}
	return
}

func (t *CronJobBuilder) buildVolume() (volumes []corev1.Volume, err error) {
	if t.CR.Spec.PVCName != nil {
		volumes = append(volumes, corev1.Volume{
			Name: "data",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: *t.CR.Spec.PVCName,
				},
			},
		})
	} else {
		quantity, err := resource.ParseQuantity(t.CR.Spec.StorageSize)
		if err != nil {
			return nil, err
		}
		volumes = append(volumes, corev1.Volume{
			Name: "data",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					SizeLimit: &quantity,
				},
			},
		})
	}

	var scriptsDefaultMode int32 = 0755
	volumes = append(volumes, corev1.Volume{
		Name: "scripts",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: BuildScriptsConfigMapName(t.CR),
				},
				DefaultMode: &scriptsDefaultMode,
			},
		},
	})
	return
}

func BuildLabels(cr *rdsv1alpha1.MysqlBackup) (labels map[string]string) {
	labels = map[string]string{
		"app":       "mysqlbackup",
		"cr-name":   cr.Name,
		"api-group": rdsv1alpha1.GroupVersion.Group,
	}
	return
}

func BuildScriptsConfigMapName(cr *rdsv1alpha1.MysqlBackup) string {
	return cr.Name + "-backup-scripts"
}
