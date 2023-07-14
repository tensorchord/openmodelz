package scaling

import "time"

type ScaleType string

const (
	// DefaultMinReplicas is the minimal amount of replicas for a service.
	DefaultMinReplicas = 1

	// DefaultMaxReplicas is the amount of replicas a service will auto-scale up to.
	DefaultMaxReplicas = 5

	DefaultZeroDuration = 3 * time.Minute

	// DefaultScalingFactor is the defining proportion for the scaling increments.
	DefaultScalingFactor = 10

	ScaleTypeRPS      ScaleType = "rps"
	ScaleTypeCapacity ScaleType = "capacity"

	// MinScaleLabel label indicating min scale for a Inference
	MinScaleLabel = "ai.tensorchord.scale.min"

	// MaxScaleLabel label indicating max scale for a Inference
	MaxScaleLabel = "ai.tensorchord.scale.max"

	// ScalingFactorLabel label indicates the scaling factor for a Inference
	ScalingFactorLabel = "ai.tensorchord.scale.factor"

	// TargetLoadLabel label indicates the target load for a Inference
	TargetLoadLabel = "ai.tensorchord.scale.target"

	// ZeroDurationLabel label indicates the zero duration for a Inference
	ZeroDurationLabel = "ai.tensorchord.scale.zero-duration"

	// ScaleTypeLabel label indicates the scale type for a Inference
	ScaleTypeLabel = "ai.tensorchord.scale.type"

	FrameworkLabel = "ai.tensorchord.framework"
)
