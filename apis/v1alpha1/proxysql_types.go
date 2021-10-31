package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProxySQLClientUser struct {
	MysqlSimpleUserInfo `json:",inline"`
	DefaultHostGroup    int `json:"defaultHostGroup"`
}

// ProxySQLSpec defines the desired state of ProxySQL
// mysql 8 admin-hash_passwords=false https://www.cnblogs.com/9527l/p/12435675.html https://kitcharoenp.github.io/mysql/2020/07/18/proxysql2_backend_users_config.html https://www.jianshu.com/p/e22b149ba270
type ProxySQLSpec struct {
	CommonField `json:",inline"`
	// StorageClassName all pods storage class name
	StorageClassName string `json:"storageClassName"`
	// MysqlVersion specific version text will return to mysql client that connected on proxysql
	MysqlVersion string `json:"mysqlVersion"`
	// NodePort  nodeport of proxysql service
	// if this value is nil, means no nodePort should be open
	// if this value is zero, means open random nodePort
	// if this value is greater than zero, means open a specific nodePort
	NodePort *int32 `json:"nodePort"`
	// Replicas proxysql pod total count,contains master and slave
	Replicas *int32 `json:"replicas,omitempty"`
	// StorageSize pvc disk size
	StorageSize string `json:"storageSize"`
	// ConfigImage proxysql initContainer for render proxysql config
	ConfigImage string `json:"configImage"`

	// Mysql mysql backend servers with host:port list
	Mysqls []MysqlHost `json:"mysqls"`
	// FrontendUsers proxysql use theese users to connect mysql server and exec sql query
	BackendUsers []ProxySQLClientUser `json:"backendUsers"`
	// FrontendUsers mysql client use theese users connect proxysql
	FrontendUsers []ProxySQLClientUser `json:"frontedUsers"`
	// AdminUsers proxysql admin users list
	AdminUsers []MysqlSimpleUserInfo `json:"adminUsers"`
	// ClusterUser proxysql cluster peers user, not mysql user
	ClusterUser MysqlSimpleUserInfo `json:"clusterUser"`
	// MonitorUser proxysql use this user to connect and monitor mysql server
	MonitorUser MysqlSimpleUserInfo `json:"monitorUser"`
	// ClusterMode mysql cluster mode
	ClusterMode ClusterMode `json:"clusterMode"`
	// MysqlMaxConn max connections per mysql instance
	MysqlMaxConn int `json:"mysqlMaxConn"`
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
