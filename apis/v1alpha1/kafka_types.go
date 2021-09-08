package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Kafka Kafka cluster CRD
type Kafka struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              KafkaSpec   `json:"spec,omitempty"`
	Status            KafkaStatus `json:"status,omitempty"`
}

type KafkaServer struct {
	CommonField `json:",inline"`
	// StorageSize pvc disk size
	StorageSize string `json:"storageSize"`
}

type KafkaClusterProxy struct {
	// Image Kafka cluster proxy image
	Image string `json:"image"`
	// Replicas Kafka cluster proxy pod relicas
	Replicas *int32 `json:"replicas,omitempty"`
	// Command Kafka cluster proxy container command
	Command []string `json:"command,omitempty"`
	// Command Kafka cluster proxy container command args
	Args []string `json:"args,omitempty"`
	// NodePort nodePort of Kafka cluster proxy service
	// if value is nil, no nodePort will be created
	// if value is 0, nodePort number is allocted by kuberentes with random number
	// if value is not nil and not equal 0, nodePort number is your specific number
	NodePort *int32 `json:"nodePort,omitempty"`
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
}

// KafkaSpec Kafka cluster spec
type KafkaSpec struct {
	// Password 密码
	Password *string `json:"password,omitempty"`
	// Replicas Kafka副本数量
	Replicas *int32 `json:"replicas"`
	// DataReplicas 数据副本数
	DataReplicas int `json:"dataReplicas"`
	// StorageClassName all pods storage class name
	StorageClassName string `json:"storageClassName"`
	// ImagePullPolicy all pods image pull policy，value should keep with corev1.PullPolicy
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	Kafka           KafkaServer       `json:"Kafka"`
	Proxy           KafkaClusterProxy `json:"proxy"`
	// TimeZone TZ envirtoment virable for all pods
	TimeZone string `json:"timeZone"`
	// Tolerations all pods tolerations，should keep with corev1.Toleration
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`
	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`
	// ServiceAccountName is the name of the ServiceAccount to use to run this pod.
	// More info: https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty" protobuf:"bytes,8,opt,name=serviceAccountName"`
	// Compute Resources required by this container.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty" protobuf:"bytes,8,opt,name=resources"`
	// PriorityClassName Kafka and Kafka-cluster-proxy pods pod priority class name
	PriorityClassName string `json:"priorityClassName,omitempty" protobuf:"bytes,24,opt,name=priorityClassName"`
}

// KafkaStatus bootstrap process status
type KafkaStatus struct {
	// Masters current Kafka cluster masters
	Masters []string `json:"masters"`
}

//+kubebuilder:object:root=true

// KafkaList a list of Kafka
type KafkaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kafka `json:"items"`
}
