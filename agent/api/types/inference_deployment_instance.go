package types

import "time"

type InferenceDeploymentInstance struct {
	Spec   InferenceDeploymentInstanceSpec   `json:"spec,omitempty"`
	Status InferenceDeploymentInstanceStatus `json:"status,omitempty"`
}

type InferenceDeploymentInstanceSpec struct {
	Namespace      string `json:"namespace,omitempty"`
	Name           string `json:"name,omitempty"`
	OwnerReference string `json:"owner_reference,omitempty"`
}

type InferenceDeploymentInstanceStatus struct {
	Phase     InstancePhase `json:"phase,omitempty"`
	StartTime time.Time     `json:"createdAt,omitempty"`
	Reason    string        `json:"reason,omitempty"`
	Message   string        `json:"message,omitempty"`
}

type InstancePhase string

const (
	InstancePhaseScheduling   InstancePhase = "Scheduling"
	InstancePhasePending      InstancePhase = "Pending"
	InstancePhaseRunning      InstancePhase = "Running"
	InstancePhaseFailed       InstancePhase = "Failed"
	InstancePhaseSucceeded    InstancePhase = "Succeeded"
	InstancePhaseUnknown      InstancePhase = "Unknown"
	InstancePhaseCreating     InstancePhase = "Creating"
	InstancePhaseInitializing InstancePhase = "Initializing"
)
