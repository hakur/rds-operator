// Package v1alpha1 contains API Schema definitions for the rds v1alpha1 API group
//+kubebuilder:object:generate=true
//+groupName=rds.hakurei.cn
package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is group version used to register these objects
	GroupVersion = schema.GroupVersion{Group: "rds.hakurei.cn", Version: "v1alpha1"}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// PodSettings pod settings
type PodSettings struct {
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type CommonField struct {
	// Pod pod settings
	Pod PodSettings `json:"pod,omitempty"`
	// Image main container image
	Image string `json:"image"`
	// ImagePullPolicy all pods image pull policy，value should keep with corev1.PullPolicy
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
	// Command Kafka container command
	Command []string `json:"command,omitempty"`
	// Command Kafka container command args
	Args []string `json:"args,omitempty"`
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
	// TimeZone timezone string , for example Asia/Shanghai
	TimeZone string `json:"timeZone"`
	// PriorityClassName pod priority class name for all pods under CR resource
	PriorityClassName string `json:"priorityClassName,omitempty" protobuf:"bytes,24,opt,name=priorityClassName"`
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
	PVCRetentionSeconds *int `json:"pvcRetentionSeconds,omitempty"`
}
