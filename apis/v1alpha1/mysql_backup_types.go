package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// S3Config aws s3 object storage server config
type S3Config struct {
	AccessKey   string
	SecurityKey string
	Endpoint    string
	Bucket      string
}

// ScpConfig scp linux server config
type ScpConfig struct {
	Host       string
	Port       int
	Username   string
	Password   string
	RemotePath string
}

// MysqlBackupStorage mysql backup files storage backend
type MysqlBackupStorage struct {
	// S3 use aws s3 object storage service for store backup files
	S3 *S3Config `json:"s3,omitempty"`
	// Pvc use a pvc for store backup files
	Pvc *string `json:"pvc,omitempty"`
	// Scp use a linux scp server for store backup files
	Scp *ScpConfig `json:"scp,omitempty"`
}

// MysqlBackupSpec defines the desired state of Mysql
type MysqlBackupSpec struct {
	// MysqlName mysql custom resource name
	MysqlName string `json:"mysqlName"`
	Storage   MysqlBackupStorage
}

// MysqlBackupStatus defines the observed state of Mysql
type MysqlBackupStatus struct {
	LastErrMsg string
	Phase      string
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=mcb

// MysqlBackup is the Schema for the mysqls API
type MysqlBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MysqlSpec   `json:"spec,omitempty"`
	Status MysqlStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MysqlBackupList contains a list of Mysql
type MysqlBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mysql `json:"items"`
}
