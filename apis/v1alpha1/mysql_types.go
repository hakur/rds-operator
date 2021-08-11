package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackupPolicy mysql data buckup method
type BackupPolicy string

// ClusterMode mysql cluster mode type
type ClusterMode string

const (
	// BackupDataDir mysql data backup method is archive mysql --datadir
	BackupDataDir BackupPolicy = "BackupDataDir"
	// ModeMGRMP cluster mode is mysql group replication multi primary
	ModeMGRMP ClusterMode = "MGRMP"
	// ModeMGRSP cluster mode is  mysql group replication single primary
	ModeMGRSP ClusterMode = "MGRSP"
	// ModeSemiSync cluster mode is  mysql semi sync
	ModeSemiSync ClusterMode = "SemiSync"
	// ModeAsync cluster mode is mysql traditional aysnc replication
	ModeAsync ClusterMode = "Async"
)

// MysqlMGRSinglePrimaryOptions mysql multi group replication single primary mode options
type MysqlMGRSinglePrimaryOptions struct {
	// ApplierThreshold mysql mgr variable: loose-group_replication_flow_control_applier_threshold
	ApplierThreshold int `json:"applierThreshold,omitempty"`
	// MGRRtries mysql mgr variable: loose-group_replication_recovery_retry_count
	MGRRetries int `json:"mgrRetries,omitempty"`
}

// MysqlUser mysql user settings
type MysqlUser struct {
	// Username mysql login account name
	Username string `json:"username"`
	// Password mysql login password of this user
	Password string `json:"password"`
	// Privileges mysql grant sql privileges, for example : []stirng{ "SELECT" ,"REPLICATION CLIENT"} or []string{"ALL PRIVILEGES"}
	Privileges []string `json:"privileges"`
	// Domain user login domain , for example : '%'
	Domain string `json:"domain"`
	// DatabaseTarget which database or tables will granted privileges to this user.
	// for example : grant all privileges on *.* to user xxx@'%' indentified by 'xxxxx', in this case, DatabaseTarget value should be '*.*'
	DatabaseTarget string `json:"databaseTarget"`
}

// MysqlOptions mysql cluster options
type MysqlOptions struct {
	// Image mysql image
	Image string `json:"image"`
	// MasterReplicas master pod count
	MasterReplicas *int32 `json:"masterReplicas,omitempty"`
	// Replicas mysql cluster pod total count,contains master and slave
	Replicas *int32 `json:"replicas,omitempty"`
	// StorageSize pvc disk size
	StorageSize string `json:"storageSize"`
	// Whitelist most of time it's kuberenetes pod CIDR and service CIDR, for example []string{"10.24.0.0/16","10.25.0.0/16"}
	Whitelist []string `json:"whitelist"`
	// MGRSP mysql multi group replication single primary mode options
	MGRSP          *MysqlMGRSinglePrimaryOptions `json:"mgrsp,omitempty"`
	ExtraConfigDir *string                       `json:"extraConfigDir,omitempty"`
	// Compute Resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
	// MysqlUsers a list of users will be created when initialize cluster
	Users           []MysqlUser `json:"users,omitempty"`
	ReplicationUser string      `json:"replicationUser"`
	MaxConn         *int        `json:"maxConn,omitempty"`
}

// ProxySQLOptions proxysql 参数
type ProxySQLOptions struct {
	// Image oci image full url ,for exmaple 'docker.io/library/nginx:1.18'
	Image string `json:"image"`
	// MysqlVersion specific version text will return to mysql client that connected on proxysql
	MysqlVersion string `json:"mysqlVersion"`
	// MonitorUser mysql server side user for proxysql monitor module connect
	MonitorUser string `json:"monitorUser"`
	// MonitorPassword password of MonitorUser
	MonitorPassword string `json:"monitorPassword"`
	// AdminUser proxysql side user for management proxysql
	AdminUser string `json:"adminUser"`
	// AdminPassword admin password for AdminUser
	AdminPassword string `json:"adminPassword"`
	// NodePort  nodeport of proxysql service
	// if this value is nil, means no nodePort should be open
	// if this value is zero,means open random nodePort
	// if this value is zero,means open a specific nodePort
	NodePort *int32 `json:"nodePort"`
	// Replicas mysql cluster pod total count,contains master and slave
	Replicas *int32 `json:"replicas,omitempty"`
	// StorageSize pvc disk size
	StorageSize string `json:"storageSize"`
	// Compute Resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
}

// MysqlSpec defines the desired state of Mysql
type MysqlSpec struct {
	// BackupPolicy mysql data backup methos , values are [ BackupDataDir ]
	BackupPolicy *BackupPolicy `json:"backupPolicy,omitempty"`
	// ClusterMode mysql cluster mode,values are [ MGRMP MGRSP SemiSync Async ]
	ClusterMode ClusterMode `json:"clusterMode"`
	// RootPassword mysql root password, if empty, an will allow empty password login
	RootPassword *string `json:"rootPassword,omitempty"`
	// ImagePullPolicy all pods image pull policy，value should keep with corev1.PullPolicy
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// Tolerations all pods tolerations，should keep with corev1.Toleration
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`
	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`
	// ServiceAccountName is the name of the ServiceAccount to use to run this pod.
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty" protobuf:"bytes,8,opt,name=serviceAccountName"`
	// StorageClassName all pods storage class name
	StorageClassName string `json:"storageClassName"`
	// Mysql mysql cluset options
	Mysql MysqlOptions `json:"mysql"`
	// ProxySQL proxysql options
	ProxySQL ProxySQLOptions `json:"proxysql"`
	// TimeZone timezone string , for example Asia/Shanghai
	TimeZone string `json:"timeZone"`
	// PriorityClassName pod priority class name for all pods under CR resource
	PriorityClassName string `json:"priorityClassName,omitempty" protobuf:"bytes,24,opt,name=priorityClassName"`
	// ConfigImage mysql initContainer for render mysql/proxysql config and boostrap mysql cluster
	ConfigImage string `json:"configImage"`
	// PVCRetentionSeconds pvc retention seconds after CR has been deleted
	// after pvc deleted, a deadline annotations will add to pvc.
	// if deadline reached (default time.Now().Unix() + PVCRetentionSeconds), and CR not found(filtered by labels), pvc will be deleted by operator.
	// if before deadline, a new CR with same labels of pvc created. pvc deadline annotation will be removed.
	// if this field value is nil, types.PVCDeleteRetentionDays will be default value to this field
	// if this field value is zero, pvc will alive forever
	PVCRetentionSeconds *int `json:"pvcRetentionSeconds,omitempty"`
}

// MysqlStatus defines the observed state of Mysql
type MysqlStatus struct {
	// Masters current mysql cluster masters
	Masters []string `json:"masters"`
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=mc

// Mysql is the Schema for the mysqls API
type Mysql struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MysqlSpec   `json:"spec,omitempty"`
	Status MysqlStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MysqlList contains a list of Mysql
type MysqlList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mysql `json:"items"`
}
