package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// S3Config aws s3 object storage server config
type S3Config struct {
	AccessKey   string `json:"accessKey"`
	SecurityKey string `json:"securityKey"`
	Endpoint    string `json:"endpoint"`
	Bucket      string `json:"bucket"`
	Path        string `json:"path"`
}

// MysqlHost mysql back server connection settings
type MysqlHost struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// BackupWebHookPostData POST http body json Data
type BackupWebHookPostData struct {
	// Status backup CR status, values are [Pending Generating Done]
	// Pending mean pod is Created or Scheduling
	// Gernating mean pod is Running
	// Done mean pod is Completed
	Status string `json:"status"`
	// Path s3 path of this backup file
	Path string `json:"path"`
	// CreateTime create time of backup operation
	CreateTime string `json:"createTime"`
	// DoneTime backup oeration done time
	DoneTime string `json:"doneTime"`
	// backup file size bytes
	Size int64 `json:"size"`
	// CostSeconds how many seconds cost of backup operation, from create to finish
	CostSeconds int `json:"costSeconds"`
	// SourceServer backup file source server
	SourceServer string `json:"sourceServer"`
}

type MysqlBackupWebhook struct {
	// URL http request url
	URL string `json:"url"`
	// Password http basic auth username
	Username string `json:"username"`
	// Password http basic auth password
	Password string `json:"password"`
	// Headers http header fields, set cookie or some bearToken in this filed
	Headers map[string]string `json:"headers"`
	// DeleteResource if this field value is true, webhook post return http response code 200 and content is "ok", then delete MysqlBackup Custom Resource
	// when backup job type is cronjob , must set this field value to false
	DeleteResource bool `json:"deleteResource"`
}

// MysqlBackupSpec defines the desired state of Mysql
type MysqlBackupSpec struct {
	CommonField `json:",inline"`
	// S3 use aws s3 object storage service for store backup files
	S3 *S3Config `json:"s3,omitempty"`
	// Mysql host for backup
	Address []MysqlHost `json:"address,omitempty"`
	// ClusterMode mysql cluster mode
	ClusterMode ClusterMode `json:"clusterMode"`
	// PVCName if pvc name is empty, a emptydir will be used as tmp storage for mysql backup files
	PVCName *string `json:"pvcName,omitempty"`
	// StorageSize mysql backup files tmp storage dir max size
	StorageSize string `json:"storageSize"`
	// Username username of all mysql hosts, used for this backup operation
	Username string `json:"username"`
	// Password password of all mysql hosts, used for this backup operation
	Password string `json:"password"`
	// Schedule k8s/linux cronjob schedule
	Schedule             string   `json:"schedule"`
	InitContainerCommand []string `json:"initContainerCommand"`
	InitContainerArgs    []string `json:"initContainerArgs"`
	InitContainerImage   string   `json:"initContainerImage"`
	// UseZlibCompress use zlib compress for mysqlpump command
	// how to extra zlib compressed mysql backup file, see ???
	UseZlibCompress *bool `json:"useZlibCompress,omitempty"`
	// Webhook send backup file info POST to webhook url
	Webhook MysqlBackupWebhook `json:"webhook,omitempty"`
}

// MysqlBackupStatus defines the observed state of Mysql
type MysqlBackupStatus struct {
	LastErrMsg string `json:"lastErrMsg,omitempty"`
	Phase      string `json:"phase,omitempty"`
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=mcb

// MysqlBackup is the Schema for the mysqls API
type MysqlBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MysqlBackupSpec   `json:"spec,omitempty"`
	Status MysqlBackupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MysqlBackupList contains a list of Mysql
type MysqlBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MysqlBackup `json:"items"`
}
