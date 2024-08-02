/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// VclusterSpec defines the desired state of Vcluster
type VclusterSpec struct {
	Config Config `json:"foo,omitempty"`

	Sleep bool `json:"sleep,omitempty"`
}

// VclusterStatus defines the observed state of Vcluster
type VclusterStatus struct {
	KubeconfigSecretReference *corev1.SecretReference `json:"kubeconfigSecretReference,omitempty"`
	KubeconfigCreated         bool                    `json:"kubeconfigCreated,omitempty"`
	Conditions                []metav1.Condition      `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Vcluster is the Schema for the vclusters API
type Vcluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VclusterSpec   `json:"spec,omitempty"`
	Status VclusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VclusterList contains a list of Vcluster
type VclusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Vcluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Vcluster{}, &VclusterList{})
}
