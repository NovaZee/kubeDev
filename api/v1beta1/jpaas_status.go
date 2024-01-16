package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
)

// Condition saves the state information of the paas
type Condition struct {
	// Status of cluster condition.
	Type ConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// The last time this condition was updated.
	LastUpdateTime string `json:"lastUpdateTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime string `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// ConditionType defines the condition
type ConditionType string

const (
	ConditionAvailable ConditionType = "Available"
	ConditionHealthy   ConditionType = "Healthy"
	ConditionRunning   ConditionType = "Running"
	ConditionCreating  ConditionType = "Creating"
	ConditionUpgrading ConditionType = "Upgrading"
	ConditionFailed    ConditionType = "Failed"
	ConditionDeleted   ConditionType = "Deleted"
	ConditionUnInit    ConditionType = "Uninitialized"
)

// ConditionType defines the condition
type Scope string

// AccessScope defines the access scope of service.
const (
	AccessScopeCluster  Scope = "Cluster"
	AccessScopeVPC      Scope = "VPC"
	AccessScopeExternal       = "External"
	AccessScopeNodePort       = "NodePort"
	AccessScopeHeadless       = "Headless"
)

type Healthy string

const (
	HealthTrue  Healthy = "healthy"
	HealthFalse Healthy = "unhealthy"
)
