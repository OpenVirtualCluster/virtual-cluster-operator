package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VClusterSpec defines the desired state of VCluster
type VClusterSpec struct {
	Sleep          bool           `json:"sleep,omitempty"`
	VClusterValues VclusterValues `json:"vclusterValues,omitempty"`
}

type VclusterValues struct {
	ControlPlane ControlPlane `json:"controlPlane,omitempty"`
}

// VClusterStatus defines the observed state of VCluster
type VClusterStatus struct {
	KubeconfigSecretReference *v1.SecretReference `json:"kubeconfigSecretReference,omitempty"`
	KubeconfigCreated         bool                `json:"kubeconfigCreated,omitempty"`
	Conditions                []metav1.Condition  `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VCluster is the Schema for the vclusters API
type VCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VClusterSpec   `json:"spec,omitempty"`
	Status VClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VClusterList contains a list of VCluster
type VClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VCluster{}, &VClusterList{})
}
