package types

import "time"

// InferenceDeploymentStatus exported for system/inferences endpoint
type InferenceDeploymentStatus struct {
	Phase Phase `json:"phase,omitempty"`

	// InvocationCount count of invocations
	InvocationCount int32 `json:"invocationCount,omitempty"`

	// Replicas desired within the cluster
	Replicas int32 `json:"replicas,omitempty"`

	// AvailableReplicas is the count of replicas ready to receive
	// invocations as reported by the faas-provider
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// CreatedAt is the time read back from the faas backend's
	// data store for when the function or its container was created.
	CreatedAt *time.Time `json:"createdAt,omitempty"`

	// Usage represents CPU and RAM used by all of the
	// functions' replicas. Divide by AvailableReplicas for an
	// average value per replica.
	Usage *InferenceUsage `json:"usage,omitempty"`

	// EventMessage record human readable message indicating details about the event of deployment.
	EventMessage string `json:"eventMessage,omitempty"`
}

type Phase string

const (
	// PhaseReady is the state of an inference when it is ready to
	// receive invocations.
	PhaseReady Phase = "Ready"

	// PhaseScaling is the state of an inference when scales.
	PhaseScaling Phase = "Scaling"

	PhaseTerminating Phase = "Terminating"

	PhaseNoReplicas Phase = "NoReplicas"

	PhaseNotReady Phase = "NotReady"

	PhaseBuilding Phase = "Building"

	PhaseOptimizing Phase = "Optimizing"
)

// InferenceUsage represents CPU and RAM used by all of the
// functions' replicas.
//
// CPU is measured in seconds consumed since the last measurement
// RAM is measured in total bytes consumed
type InferenceUsage struct {
	// CPU is the increase in CPU usage since the last measurement
	// equivalent to Kubernetes' concept of millicores.
	CPU float64 `json:"cpu,omitempty"`

	//TotalMemoryBytes is the total memory usage in bytes.
	TotalMemoryBytes float64 `json:"totalMemoryBytes,omitempty"`

	GPU float64 `json:"gpu,omitempty"`
}
