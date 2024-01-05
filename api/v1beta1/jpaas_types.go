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

type Type string

const (
	Application Type = "App"
	BaseApp     Type = "Base"
	Platform    Type = "Platform"
)

func (t Type) String() string {
	return string(t)
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// JPaasSpec defines the desired state of JPaas
type JPaasSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//应用基础属性
	//image当应用类型为Platform时，image可以忽略
	Namespace string          `json:"namespace"`
	Image     string          `json:"image,omitempty"`
	Replicas  int32           `json:"replicas"`
	Version   string          `json:"version"`
	Env       []corev1.EnvVar `json:"env,omitempty"`
	//应用是否强依赖
	VersionAligned bool `json:"version_aligned"`
	//crd 类型 Init代表JPaas初始化 App代表是具体应用的crd
	//如果是Platform类型，AppRefs为依赖基础应用，如果是Application类型，AppRefs是对应app以来的产品以及版本
	Type    Type     `json:"type,omitempty"`
	AppRefs []AppRef `json:"app_refs,omitempty"`
}

// JPaasStatus defines the observed state of JPaas
type JPaasStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

type AppRef struct {
	Name    string `json:"name"`
	Image   string `json:"image"`
	Version string `json:"version"`
	Status  bool   `json:"status"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="namespace",type="string",JSONPath=".spec.namespace",description="The namespace of Paas"
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type",description="The type of JPaas"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.conditions[].type",description="The status of Redis Cluster"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="The age of the resource"

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
