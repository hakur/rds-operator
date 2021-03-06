package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Redis redis cluster CRD
type Redis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RedisSpec   `json:"spec,omitempty"`
	Status            RedisStatus `json:"status,omitempty"`
}

type RedisServer struct {
	// Image redis server image
	Image string `json:"image"`
	// StorageSize pvc disk size
	StorageSize string `json:"storageSize"`
	// BackupMethod data backup method, valid value is [ AOF RDB ]
	BackupMethod string `json:"backupMethod,omitempty"`
	// Command redis container command
	Command []string `json:"command,omitempty"`
	// Command redis container command args
	Args []string `json:"args,omitempty"`
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

type Predixy struct {
	// Image redis cluster proxy image
	Image string `json:"image"`
	// Replicas redis cluster proxy pod relicas
	Replicas *int32 `json:"replicas,omitempty"`
	// Command redis cluster proxy container command
	Command []string `json:"command,omitempty"`
	// Command redis cluster proxy container command args
	Args []string `json:"args,omitempty"`
	// NodePort nodePort of redis cluster proxy service
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

type RedisClusterProxy struct {
	// Image redis cluster proxy image
	Image string `json:"image"`
	// Replicas redis cluster proxy pod relicas
	Replicas *int32 `json:"replicas,omitempty"`
	// Command redis cluster proxy container command
	Command []string `json:"command,omitempty"`
	// Command redis cluster proxy container command args
	Args []string `json:"args,omitempty"`
	// NodePort nodePort of redis cluster proxy service
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

type RedisMonitor struct {
	// Image oliver006/redis_exporter image
	Image string `json:"image"`
	// Args container run args
	Args []string `json:"args,omitempty"`
	// Args
	// Interval service monitor interval
	Interval string `json:"interval,omitempty"`
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
	// PVCRetentionSeconds pvc retention seconds after CR has been deleted
	// after pvc deleted, a deadline annotations will add to pvc.
	// if deadline reached (default time.Now().Unix() + PVCRetentionSeconds), and CR not found(filtered by labels), pvc will be deleted by operator.
	// if before deadline, a new CR with same labels of pvc created. pvc deadline annotation will be removed.
	// if this field value is nil, types.PVCDeleteRetentionSeconds will be default value to this field
	// if this field value is zero, pvc will alive forever
}

// RedisSpec redis cluster spec
type RedisSpec struct {
	// Password ??????
	Password *string `json:"password,omitempty"`
	// Replicas redis????????????
	MasterReplicas int `json:"masterReplicas"`
	// DataReplicas ???????????????
	DataReplicas int `json:"dataReplicas"`
	// StorageClassName all pods storage class name
	StorageClassName string `json:"storageClassName"`
	// ImagePullPolicy all pods image pull policy???value should keep with corev1.PullPolicy
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	Redis           RedisServer       `json:"redis"`
	// RedisClusterProxy use redislabs redis cluster proxy as cluster proxy
	RedisClusterProxy *RedisClusterProxy `json:"redisClusterProxy,omitempty"`
	// Predixy use predixy as redis cluster proxy
	Predixy *Predixy `json:"predixy,omitempty"`
	// TimeZone TZ envirtoment virable for all pods
	TimeZone string `json:"timeZone"`
	// Tolerations all pods tolerations???should keep with corev1.Toleration
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
	// PriorityClassName redis and redis-cluster-proxy pods pod priority class name
	PriorityClassName string        `json:"priorityClassName,omitempty" protobuf:"bytes,24,opt,name=priorityClassName"`
	Monitor           *RedisMonitor `json:"monitor,omitempty"`
}

// RedisStatus bootstrap process status
type RedisStatus struct {
	// Masters current redis cluster masters
	Masters []string `json:"masters"`
}

//+kubebuilder:object:root=true

// RedisList a list of Redis
type RedisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Redis `json:"items"`
}
