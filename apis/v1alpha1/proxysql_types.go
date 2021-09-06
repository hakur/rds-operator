package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProxySQLSpec defines the desired state of ProxySQL
type ProxySQLSpec struct {
	// Image oci image full url ,for exmaple 'docker.io/library/nginx:1.18'
	Image string `json:"image"`
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
	// PVCRetentionSeconds pvc retention seconds after CR has been deleted
	// after pvc deleted, a deadline annotations will add to pvc.
	// if deadline reached (default time.Now().Unix() + PVCRetentionSeconds), and CR not found(filtered by labels), pvc will be deleted by operator.
	// if before deadline, a new CR with same labels of pvc created. pvc deadline annotation will be removed.
	// if this field value is nil, types.PVCDeleteRetentionSeconds will be default value to this field
	// if this field value is zero, pvc will alive forever
	PVCRetentionSeconds *int `json:"pvcRetentionSeconds,omitempty"`
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
	// Compute Resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
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
