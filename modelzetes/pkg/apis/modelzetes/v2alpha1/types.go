package v2alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`

// Inference describes an Inference
type Inference struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InferenceSpec `json:"spec"`
}

// InferenceSpec defines the desired state of Inference
type InferenceSpec struct {
	Name string `json:"name"`

	Image string `json:"image"`

	// Scaling is the scaling configuration for the inference.
	Scaling *ScalingConfig `json:"scaling,omitempty"`

	// Framework is the inference framework.
	Framework Framework `json:"framework,omitempty"`

	// Port is the port exposed by the inference.
	Port *int32 `json:"port,omitempty"`

	// HTTPProbePath is the path of the http probe.
	HTTPProbePath *string `json:"http_probe_path,omitempty"`

	// Command to run when starting the
	Command *string `json:"command,omitempty"`

	// EnvVars can be provided to set environment variables for the inference runtime.
	EnvVars map[string]string `json:"envVars,omitempty"`

	// Constraints are specific to the operator.
	Constraints []string `json:"constraints,omitempty"`

	// Secrets list of secrets to be made available to inference
	Secrets []string `json:"secrets,omitempty"`

	// Labels are metadata for inferences which may be used by the
	// faas-provider or the gateway
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations are metadata for inferences which may be used by the
	// faas-provider or the gateway
	Annotations map[string]string `json:"annotations,omitempty"`

	// Limits for inference
	Resources *v1.ResourceRequirements `json:"resources,omitempty"`
}

// Framework is the inference framework. It is only used to set the default port
// and command. For example, if the framework is "gradio", the default port is
// 7860 and the default command is "python app.py". You could override these
// defaults by setting the port and command fields and framework to `other`.
type Framework string

const (
	FrameworkGradio    Framework = "gradio"
	FrameworkStreamlit Framework = "streamlit"
	FrameworkMosec     Framework = "mosec"
	FrameworkOther     Framework = "other"
)

type ScalingConfig struct {
	// MinReplicas is the lower limit for the number of replicas to which the
	// autoscaler can scale down. It defaults to 0.
	MinReplicas *int32 `json:"min_replicas,omitempty"`
	// MaxReplicas is the upper limit for the number of replicas to which the
	// autoscaler can scale up. It cannot be less that minReplicas. It defaults
	// to 1.
	MaxReplicas *int32 `json:"max_replicas,omitempty"`
	// TargetLoad is the target load. In capacity mode, it is the expected number of the inflight requests per replica.
	TargetLoad *int32 `json:"target_load,omitempty"`
	// Type is the scaling type. It can be either "capacity" or "rps". Default is "capacity".
	Type *ScalingType `json:"type,omitempty"`
	// ZeroDuration is the duration of zero load before scaling down to zero. Default is 5 minutes.
	ZeroDuration *int32 `json:"zero_duration,omitempty"`
	// StartupDuration is the duration of startup time.
	StartupDuration *int32 `json:"startup_duration,omitempty"`
}

type ScalingType string

const (
	ScalingTypeCapacity ScalingType = "capacity"
	ScalingTypeRPS      ScalingType = "rps"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InferenceList is a list of inference resources
type InferenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []Inference `json:"items"`
}
