package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProxySQLSpec defines the desired state of ProxySQL
type ProxySQLSpec struct {
	CommonField `json:",inline"`
	// StorageClassName all pods storage class name
	StorageClassName string `json:"storageClassName"`
	// MysqlVersion specific version text will return to mysql client that connected on proxysql
	MysqlVersion string `json:"mysqlVersion"`
	// NodePort  nodeport of proxysql service
	// if this value is nil, means no nodePort should be open
	// if this value is zero,means open random nodePort
	// if this value is zero,means open a specific nodePort
	NodePort *int32 `json:"nodePort"`
	// Replicas proxysql pod total count,contains master and slave
	Replicas *int32 `json:"replicas,omitempty"`
	// StorageSize pvc disk size
	StorageSize string `json:"storageSize"`
	// ConfigImage mysql initContainer for render mysql/proxysql config and boostrap mysql cluster
	ConfigImage string `json:"configImage"`
}

// ProxySQLStatus defines the observed state of ProxySQL
type ProxySQLStatus struct {
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ProxySQL is the Schema for the ProxySQLs API
type ProxySQL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProxySQLSpec   `json:"spec,omitempty"`
	Status ProxySQLStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ProxySQLList contains a list of ProxySQL
type ProxySQLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProxySQL `json:"items"`
}
