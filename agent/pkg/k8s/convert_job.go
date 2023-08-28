package k8s

import (
	v1 "k8s.io/api/batch/v1"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

func AsBuild(job v1.Job) (types.Build, error) {
	build := types.Build{
		Spec: types.BuildSpec{
			Name:      job.Name,
			Namespace: job.Namespace,
		},
	}

	if job.Status.Succeeded > 0 {
		build.Status.Phase = types.BuildPhaseSucceeded
	} else if job.Status.Failed > 0 {
		build.Status.Phase = types.BuildPhaseFailed
	} else if job.Status.Active > 0 {
		build.Status.Phase = types.BuildPhaseRunning
	} else {
		build.Status.Phase = types.BuildPhasePending
	}

	return build, nil
}
