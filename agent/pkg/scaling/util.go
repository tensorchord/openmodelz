package scaling

import (
	"time"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

func AsServerQueryResponse(inf *types.InferenceDeployment) (*ServiceQueryResponse, error) {
	if inf == nil {
		return nil, nil
	}
	res := ServiceQueryResponse{}

	res.Replicas = uint64(inf.Status.Replicas)
	res.Annotations = inf.Spec.Annotations
	res.AvailableReplicas = uint64(inf.Status.AvailableReplicas)
	res.Framework = string(inf.Spec.Framework)
	res.MinReplicas = uint64(*inf.Spec.Scaling.MinReplicas)
	res.MaxReplicas = uint64(*inf.Spec.Scaling.MaxReplicas)
	res.TargetLoad = uint64(*inf.Spec.Scaling.TargetLoad)
	res.ZeroDuration = time.Duration(*inf.Spec.Scaling.ZeroDuration) * time.Second
	return &res, nil
}
