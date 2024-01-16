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
	v1 "k8s.io/api/core/v1"
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
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Status      bool   `json:"status,omitempty"`
	Initialized bool   `json:"Initialized"`
	NeedUpgrade bool   `json:"needUpgrade"`
	// 镜像前缀地址 除了版本号
	ImageUrl  string      `json:"imageUrl,omitempty"`
	CommonEnv []v1.EnvVar `json:"commonEnv,omitempty"`
	//应用是否强依赖
	VersionAligned bool      `json:"versionAligned"`
	AppRefs        []AppSpec `json:"appRefs,omitempty"`
}

// JPaasStatus defines the observed state of JPaas
type JPaasStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions []Condition `json:"conditions,omitempty"`
}

type AppSpec struct {
	Name        string `json:"name"`
	AccessScope Scope  `json:"accessScope,omitempty"`
	Version     string `json:"version"`
	//基础应用类型,业务应用类型
	Type Type `json:"type,omitempty"`
	// failed,uninitialized,deleted,available,upgrading
	AppCrStatus ConditionType `json:"appCrStatus"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type",description="The type of paasterm"
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
