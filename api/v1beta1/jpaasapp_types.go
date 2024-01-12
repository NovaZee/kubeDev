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

// JPaasAppSpec defines the desired state of JPaasApp
type JPaasAppSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Image            ImageSpec            `json:"image"`
	EmbeddedResource EmbeddedResourceSpec `json:"embeddedResource"`
	Env              []corev1.EnvVar      `json:"env,omitempty"`
	//是否健康检查
	Health bool `json:"health,omitempty"`
	//基础应用类型,业务应用类型
	Type Type `json:"type,omitempty"`
}

// EmbeddedResourceSpec defines properties of EmbeddedResourceSpec.
type EmbeddedResourceSpec struct {
	// The number of replicas.
	Replicas *int32 `json:"replicas,omitempty"`

	// Access scope, enum("Cluster", "VPC", "External").
	AccessScope Scope `json:"accessScope"`

	// Ports.
	Ports Ports `json:"ports,omitempty"`

	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Volumes in the JobManager pod. todo：挂载路径后续统一
	Volumes []corev1.Volume `json:"volumes,omitempty"`

	// Volume mounts in the JobManager container.
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

// Ports defines ports of app.
type Ports struct {

	// Ws port, default: 443.
	Ws *int32 `json:"ws,omitempty"`

	// Front port, default: 8080.
	Default *int32 `json:"default,omitempty"`

	// UI port, default: 8081.
	UI *int32 `json:"ui,omitempty"`
}

// ImageSpec defines image ofcontainers.
type ImageSpec struct {
	// Flink image name.
	Name string `json:"name"`
	// Image pull policy. One of Always, Never, IfNotPresent. Defaults to Always
	// if :latest tag is specified, or IfNotPresent otherwise.
	PullPolicy corev1.PullPolicy `json:"pullPolicy,omitempty"`

	// Secrets for image pull.
	PullSecrets []corev1.LocalObjectReference `json:"pullSecrets,omitempty"`
}

// JPaasAppStatus defines the observed state of JPaasApp
type JPaasAppStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The overall state of the app
	State          string           `json:"state"`
	Components     ComponentsStatus `json:"components"`
	ComponentsInit bool             `json:"componentsInit"`
	// Last update timestamp for this status.
	LastUpdateTime string `json:"lastUpdateTime,omitempty"`
}

type ComponentsStatus struct {
	// The state of JobManager StatefulSet.
	AppDeployment AppComponentsStatus `json:"jobManagerStatefulSet"`

	// The state of JobManager service.
	AppService AppComponentServiceStatus `json:"jobManagerService"`

	// The state of JobManager ingress.
	// The state of TaskManager StatefulSet.
	// The state of configMap.
	// ConfigMap AppComponentsStatus `json:"configMap"`
}

type AppComponentsStatus struct {
	// The resource name of the component.
	Name string `json:"name"`

	// The state of the component.
	State string `json:"state"`
}

// AppComponentServiceStatus represents the state of a Kubernetes service.
type AppComponentServiceStatus struct {
	// The name of the Kubernetes jobManager service.
	Name string `json:"name"`

	// The state of the component.
	State string `json:"state"`

	// (Optional) The node port, present when `accessScope` is `NodePort`.
	NodePort int32 `json:"nodePort,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type",description="The type of client"

// JPaasApp is the Schema for the jpaasapps API
type JPaasApp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JPaasAppSpec   `json:"spec,omitempty"`
	Status JPaasAppStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// JPaasAppList contains a list of JPaasApp
type JPaasAppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JPaasApp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JPaasApp{}, &JPaasAppList{})
}
