package types

// InferenceDeployment represents a request to create or update a Model.
type InferenceDeployment struct {
	Spec   InferenceDeploymentSpec   `json:"spec"`
	Status InferenceDeploymentStatus `json:"status,omitempty"`
}

type InferenceDeploymentSpec struct {
	// Name is the name of the inference.
	Name string `json:"name"`

	// Namespace for the inference.
	Namespace string `json:"namespace,omitempty"`

	// Scaling is the scaling configuration for the inference.
	Scaling *ScalingConfig `json:"scaling,omitempty"`

	// Framework is the inference framework.
	Framework Framework `json:"framework,omitempty"`

	// Image is a fully-qualified container image
	Image string `json:"image"`

	// Port is the port exposed by the inference.
	Port *int32 `json:"port,omitempty"`

	// HTTPProbePath is the path of the http probe.
	HTTPProbePath *string `json:"http_probe_path,omitempty"`

	// Command to run when starting the
	Command *string `json:"command,omitempty"`

	// EnvVars can be provided to set environment variables for the inference runtime.
	EnvVars map[string]string `json:"envVars,omitempty"`

	// Constraints are the constraints for the inference.
	Constraints []string `json:"constraints,omitempty"`

	// Secrets list of secrets to be made available to inference.
	Secrets []string `json:"secrets,omitempty"`

	// Labels are key-value pairs that may be attached to the inference.
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations are key-value pairs that may be attached to the inference.
	Annotations map[string]string `json:"annotations,omitempty"`

	// Resources are the compute resource requirements.
	Resources *ResourceRequirements `json:"resources,omitempty"`
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
	// ZeroDuration is the duration (in seconds) of zero load before scaling down to zero. Default is 5 minutes.
	ZeroDuration *int32 `json:"zero_duration,omitempty"`
	// StartupDuration is the duration (in seconds) of startup time.
	StartupDuration *int32 `json:"startup_duration,omitempty"`
}

type ScalingType string

const (
	ScalingTypeCapacity ScalingType = "capacity"
	ScalingTypeRPS      ScalingType = "rps"
)

// ResourceRequirements describes the compute resource requirements.
type ResourceRequirements struct {
	// Limits describes the maximum amount of compute resources allowed.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Limits ResourceList `json:"limits,omitempty" protobuf:"bytes,1,rep,name=limits,casttype=ResourceList,castkey=ResourceName"`
	// Requests describes the minimum amount of compute resources required.
	// If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
	// otherwise to an implementation-defined value.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Requests ResourceList `json:"requests,omitempty" protobuf:"bytes,2,rep,name=requests,casttype=ResourceList,castkey=ResourceName"`
}

// ResourceList is a set of (resource name, quantity) pairs.
type ResourceList map[ResourceName]Quantity

type ResourceName string

const (
	ResourceCPU    ResourceName = "cpu"
	ResourceMemory ResourceName = "memory"
	ResourceGPU    ResourceName = "gpu"
)

type Quantity string

const (
	RuntimeClassNvidia string = "nvidia"
)

type ImageCache struct {
	// Name is the name of the inference.
	Name           string `json:"name"`
	Namespace      string `json:"namespace"`
	Image          string `json:"image"`
	ForceFullCache bool   `json:"force_full_cache"`
	NodeSelector   string `json:"node_selector"`
}
