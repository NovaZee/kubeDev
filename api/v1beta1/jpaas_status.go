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
	ConditionRunning                 = "Running"
	ConditionCreating                = "Creating"
	ConditionUpgrading               = "Upgrading"
	ConditionFailed                  = "Failed"
)
