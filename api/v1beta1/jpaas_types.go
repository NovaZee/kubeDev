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

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// JPaasSpec defines the desired state of JPaas
type JPaasSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of JPaas. Edit jpaas_types.go to remove/update
	image         string          `json:"image"`
	Replicas      int32           `json:"replicas"`
	EnableService bool            `json:"enable_service"`
	Env           []corev1.EnvVar `json:"env,omitempty"`
}

// JPaasStatus defines the observed state of JPaas
type JPaasStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// JPaas is the Schema for the jpaas API
type JPaas struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JPaasSpec   `json:"spec,omitempty"`
	Status JPaasStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// JPaasList contains a list of JPaas
type JPaasList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JPaas `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JPaas{}, &JPaasList{})
}
