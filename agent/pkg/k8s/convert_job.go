package k8s

import (
	"fmt"

	v1 "k8s.io/api/batch/v1"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/pkg/consts"
)

func AsBuild(job v1.Job) (types.Build, error) {
	build := types.Build{
		Spec: types.BuildSpec{
			Name:      job.Name,
			Namespace: job.Namespace,
			ProjectID: job.Labels[consts.LabelProjectID],
		},
	}

	build.Status.Image = fmt.Sprintf("%s:%s",
		build.Spec.BuildSource.ArtifactImage,
		build.Spec.BuildSource.ArtifactImageTag)

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
