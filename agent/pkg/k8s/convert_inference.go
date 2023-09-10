package k8s

import (
	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func AsInferenceDeployment(inf *v2alpha1.Inference, item *appsv1.Deployment) *types.InferenceDeployment {
	if inf == nil {
		return nil
	}

	res := &types.InferenceDeployment{
		Spec: types.InferenceDeploymentSpec{
			Name:        inf.Name,
			Framework:   types.Framework(inf.Spec.Framework),
			Image:       inf.Spec.Image,
			Namespace:   inf.Namespace,
			EnvVars:     inf.Spec.EnvVars,
			Secrets:     inf.Spec.Secrets,
			Constraints: inf.Spec.Constraints,
			Labels:      inf.Spec.Labels,
			Annotations: inf.Spec.Annotations,
		},
		Status: types.InferenceDeploymentStatus{
			Phase: types.PhaseNoReplicas,
		},
	}

	if inf.Spec.Scaling != nil {
		res.Spec.Scaling = &types.ScalingConfig{
			MinReplicas:     inf.Spec.Scaling.MinReplicas,
			MaxReplicas:     inf.Spec.Scaling.MaxReplicas,
			TargetLoad:      inf.Spec.Scaling.TargetLoad,
			ZeroDuration:    inf.Spec.Scaling.ZeroDuration,
			StartupDuration: inf.Spec.Scaling.StartupDuration,
		}
		if inf.Spec.Scaling.Type != nil {
			typ := types.ScalingType(*inf.Spec.Scaling.Type)
			res.Spec.Scaling.Type = &typ
		}
	}

	if inf.Spec.Port != nil {
		res.Spec.Port = inf.Spec.Port
	}

	var replicas int32 = 0
	// Get status according to the deployment.
	if item != nil {
		if item.Spec.Replicas != nil {
			replicas = *item.Spec.Replicas
		}
		res.Status.Replicas = replicas
		res.Status.CreatedAt = &item.CreationTimestamp.Time
		res.Status.InvocationCount = 0
		res.Status.AvailableReplicas = item.Status.AvailableReplicas

		res.Status.Phase = types.PhaseNotReady
		for _, c := range item.Status.Conditions {
			if c.Type == appsv1.DeploymentAvailable && c.Status == v1.ConditionTrue {
				res.Status.Phase = types.PhaseReady
			} else if c.Type == appsv1.DeploymentProgressing && c.Status == v1.ConditionFalse {
				res.Status.Phase = types.PhaseScaling
			}
		}

		if item.Spec.Replicas != nil && *item.Spec.Replicas == 0 {
			res.Status.Phase = types.PhaseNoReplicas
		}

		if item.DeletionTimestamp != nil {
			res.Status.Phase = types.PhaseTerminating
		}
	}
	return res
}

func AsResourceList(resources v1.ResourceList) types.ResourceList {
	res := types.ResourceList{}
	gpuResource := resources[consts.ResourceNvidiaGPU]
	gpuPtr := &gpuResource

	if !resources.Cpu().IsZero() {
		res[types.ResourceCPU] = types.Quantity(
			resources.Cpu().String())
	}
	if !resources.Memory().IsZero() {
		res[types.ResourceMemory] = types.Quantity(
			resources.Memory().String())
	}
	if !gpuPtr.IsZero() {
		res[types.ResourceGPU] = types.Quantity(
			gpuPtr.String())
	}
	return res
}
