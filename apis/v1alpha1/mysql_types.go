package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterMode mysql cluster mode type
type ClusterMode string

// ClusterPhase mysql cluster status
type ClusterPhase string

const (
	// ModeMGRMP cluster mode is mysql group replication multi primary
	ModeMGRMP ClusterMode = "MGRMP"
	// ModeMGRSP cluster mode is  mysql group replication single primary
	ModeMGRSP ClusterMode = "MGRSP"
	// ModeSemiSync cluster mode is  mysql semi sync
	ModeSemiSync ClusterMode = "SemiSync"

	MysqlPhaseNotReady    ClusterPhase = "NotReady"
	MysqlPhaseRunning     ClusterPhase = "Running"
	MysqlPhaseTerminating ClusterPhase = "Terminating"
)

// MysqlMGRSinglePrimaryOptions mysql multi group replication single primary mode options
type MysqlMGRSinglePrimaryOptions struct {
	// ApplierThreshold mysql mgr variable: loose-group_replication_flow_control_applier_threshold
	ApplierThreshold int `json:"applierThreshold,omitempty"`
	// MGRRtries mysql mgr variable: loose-group_replication_recovery_retry_count
	MGRRetries int `json:"mgrRetries,omitempty"`
}

type MysqlSemiSyncOptions struct {
	// DoubleMasterHA if true , mysql-0 and mysql-1 will be cluster masters,they copy data from each other
	DoubleMasterHA bool `json:"doubleMasterHA,omitempty"`
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

// MysqlSpec defines the desired state of Mysql
type MysqlSpec struct {
	CommonField `json:",inline"`
	// ClusterMode mysql cluster mode,values are [ MGRMP MGRSP SemiSync Async ]
	ClusterMode ClusterMode `json:"clusterMode"`
	// RootPassword mysql root password, if empty, an will allow empty password login
	RootPassword *string `json:"rootPassword,omitempty"`
	// StorageClassName kuberentes storage class name of this mysql pod
	StorageClassName string `json:"storageClassName"`
	// ConfigImage mysql initContainer for render mysql/proxysql config and boostrap mysql cluster
	ConfigImage string `json:"configImage"`
	// Replicas mysql cluster pod total count,contains master and slave
	Replicas *int32 `json:"replicas,omitempty"`
	// StorageSize pvc disk size
	StorageSize string `json:"storageSize"`
	// Whitelist most of time it's kuberenetes pod CIDR and service CIDR, for example []string{"10.24.0.0/16","10.25.0.0/16"}
	Whitelist []string `json:"whitelist"`
	// MGRSP mysql multi group replication single primary mode options
	MGRSP *MysqlMGRSinglePrimaryOptions `json:"mgrsp,omitempty"`
	// SemiSync mysql semi sync replication options
	SemiSync *MysqlSemiSyncOptions `json:"semiSync,omitempty"`
	// ExtraConfig write your own mysql config to override operator nested mysql config.
	// content will merge into ${extraConfigDir}/my.cnf
	ExtraConfig string `json:"extraConfig,omitempty"`
	// ExtraConfigDir my.cnf include dir
	ExtraConfigDir *string `json:"extraConfigDir,omitempty"`
	// ClusterUser mysql cluster replication user
	ClusterUser *MysqlUser `json:"clusterUser,omitempty"`
	MaxConn     *int       `json:"maxConn,omitempty"`
}

// MysqlStatus defines the observed state of Mysql
type MysqlStatus struct {
	// Masters current mysql cluster masters
	Masters        []string     `json:"masters,omitempty"`
	Members        []string     `json:"members,omitempty"`
	HealthyMembers []string     `json:"healthyMembers,omitempty"`
	Phase          ClusterPhase `json:"phase,omitempty"`
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=mc
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:JSONPath=".status.phase",name=phase,type=string
//+kubebuilder:printcolumn:JSONPath=".spec.clusterMode",name=cluster_mode,type=string

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
