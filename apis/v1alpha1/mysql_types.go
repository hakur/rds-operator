package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackupPolicy mysql data buckup method
type BackupPolicy string

// ClusterMode mysql cluster mode type
type ClusterMode string

// ClusterPhase mysql cluster status
type ClusterPhase string

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

	MysqlPhaseCreating    ClusterPhase = "Creating"
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
	// if this field value is nil, types.PVCDeleteRetentionSeconds will be default value to this field
	// if this field value is zero, pvc will alive forever
	PVCRetentionSeconds *int `json:"pvcRetentionSeconds,omitempty"`
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
	Users   []MysqlUser `json:"users,omitempty"`
	MaxConn *int        `json:"maxConn,omitempty"`
	// Periodic probe of container liveness.
	// Container will be restarted if the probe fails.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	// +optional
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty" protobuf:"bytes,10,opt,name=livenessProbe"`
	// Periodic probe of container service readiness.
	// Container will be removed from service endpoints if the probe fails.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#container-probes
	// +optional
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty" protobuf:"bytes,11,opt,name=readinessProbe"`
}

// MysqlStatus defines the observed state of Mysql
type MysqlStatus struct {
	// Masters current mysql cluster masters
	Masters []string     `json:"masters,omitempty"`
	Members []string     `json:"members,omitempty"`
	Phase   ClusterPhase `json:"phase,omitempty"`
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
