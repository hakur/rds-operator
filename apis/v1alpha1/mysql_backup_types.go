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

// MysqlBackupSpec defines the desired state of Mysql
type MysqlBackupSpec struct {
	CommonField `json:",inline"`
	// S3 use aws s3 object storage service for store backup files
	S3          *S3Config   `json:"s3,omitempty"`
	Hosts       []MysqlHost `json:"mysql,omitempty"`
	ClusterMode ClusterMode `json:"clusterMode"`
	// PVCName if pvc name is empty, a emptydir will be used as tmp storage for mysql backup files
	PVCName *string `json:"pvcName,omitempty"`
	// StorageSize mysql backup files tmp storage dir max size
	StorageSize string `json:"storageSize"`
	// Username username of all mysql hosts
	Username string `json:"username"`
	// Password password of all mysql hosts
	Password string `json:"password"`
	// Schedule k8s/linux cronjob schedule
	Schedule             string   `json:"schedule"`
	InitContainerCommand []string `json:"initContainerCommand"`
	InitContainerArgs    []string `json:"initContainerArgs"`
	InitContainerImage   string   `json:"initContainerImage"`
	UseZlibCompress      *bool    `json:"useZlibCompress,omitempty"`
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
