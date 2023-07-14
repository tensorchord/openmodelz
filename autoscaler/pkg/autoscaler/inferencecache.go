package autoscaler

import (
	"time"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

type Inference struct {
	Deployment types.InferenceDeployment
	Timestamp  time.Time
}

type InferenceCache struct {
	inference map[string]Inference
}

func newInferenceCache() *InferenceCache {
	return &InferenceCache{
		inference: make(map[string]Inference),
	}
}

func (i *InferenceCache) Set(key string, inference Inference) {
	i.inference[key] = inference
}

func (i *InferenceCache) Get(key string, expireTime time.Duration) (types.InferenceDeployment, bool) {
	inference, ok := i.inference[key]

	// expired
	if !ok || time.Since(inference.Timestamp) > expireTime {
		return types.InferenceDeployment{}, false
	}
	return inference.Deployment, ok
}
