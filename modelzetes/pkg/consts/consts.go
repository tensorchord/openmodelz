package consts

const (
	ResourceNvidiaGPU = "nvidia.com/gpu"

	LabelInferenceName      = "inference"
	LabelInferenceNamespace = "inference-namespace"
	LabelBuildName          = "ai.tensorchord.build"
	LabelName               = "ai.tensorchord.name"
	LabelNamespace          = "modelz.tensorchord.ai/namespace"
	LabelServerResource     = "ai.tensorchord.server-resource"

	AnnotationBuilding        = "ai.tensorchord.building"
	AnnotationDockerImage     = "ai.tensorchord.docker.image"
	AnnotationControlPlaneKey = "ai.tensorchord.control-plane"

	ModelzAnnotationValue = "modelz"

	TolerationGPU              = "ai.tensorchord.gpu"
	TolerationNvidiaGPUPresent = "nvidia.com/gpu"

	//OrchestrationIdentifier identifier string for provider orchestration
	OrchestrationIdentifier = "kubernetes"
	//ProviderName name of the provider
	ProviderName = "modelzetes"

	DefaultServicePrefix = "mdz-"

	DefaultHTTPProbePath = "/"

	// MaxReplicas is the maximum number of replicas that can be set for a inference.
	MaxReplicas = 5

	GCSCSIDriverName      = "gcs.csi.ofek.dev"
	GCSVolumeHandle       = "csi-gcs"
	GCSStorageClassName   = "csi-gcs-sc"
	LocalStorageClassName = "local-storage"
)
