package runtime

import (
	"context"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/tensorchord/openmodelz/agent/api/types"
)

func (r *Runtime) InferenceList(ns string) ([]types.InferenceDeployment, error) {
	r.mutex.Lock()
	res := r.cache
	r.mutex.Unlock()

	ctrs, err := r.cli.ContainerList(context.TODO(), dockertypes.ContainerListOptions{
		Filters: filters.NewArgs(filters.Arg("label", labelVendor+"="+valueVendor)),
	})
	if err != nil {
		return nil, err
	}

	for _, ctr := range ctrs {
		inf := types.InferenceDeployment{
			Spec: types.InferenceDeploymentSpec{
				Name:      ctr.Labels[labelName],
				Image:     ctr.Image,
				Namespace: "default",
			},
			Status: types.InferenceDeploymentStatus{},
		}
		if ctr.State == "running" {
			inf.Status.Phase = types.PhaseReady
			inf.Status.AvailableReplicas = 1
			inf.Status.Replicas = 1
		} else {
			inf.Status.Phase = types.PhaseNotReady
			inf.Status.Replicas = 1
		}

		res[inf.Spec.Name] = inf
	}

	l := []types.InferenceDeployment{}
	for _, inf := range res {
		l = append(l, inf)
	}
	return l, nil
}
