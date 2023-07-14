package runtime

import (
	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

func createResources(request types.InferenceDeployment) (corev1.ResourceRequirements, error) {
	resources := corev1.ResourceRequirements{
		Limits:   corev1.ResourceList{},
		Requests: corev1.ResourceList{},
	}

	if request.Spec.Resources == nil {
		return resources, nil
	}

	// Set Memory limits
	if request.Spec.Resources.Limits[types.ResourceMemory] != "" {
		qty, err := resource.ParseQuantity(
			string(request.Spec.Resources.Limits[types.ResourceMemory]))
		if err != nil {
			return resources, err
		}
		resources.Limits[corev1.ResourceMemory] = qty
	}

	if request.Spec.Resources.Requests[types.ResourceMemory] != "" {
		qty, err := resource.ParseQuantity(
			string(request.Spec.Resources.Requests[types.ResourceMemory]))
		if err != nil {
			return resources, err
		}
		resources.Requests[corev1.ResourceMemory] = qty
	}

	// Set CPU limits
	if request.Spec.Resources.Limits[types.ResourceCPU] != "" {
		qty, err := resource.ParseQuantity(
			string(request.Spec.Resources.Limits[types.ResourceCPU]))
		if err != nil {
			return resources, err
		}
		resources.Limits[corev1.ResourceCPU] = qty
	}

	if request.Spec.Resources.Requests[types.ResourceCPU] != "" {
		qty, err := resource.ParseQuantity(
			string(request.Spec.Resources.Requests[types.ResourceCPU]))
		if err != nil {
			return resources, err
		}
		resources.Requests[corev1.ResourceCPU] = qty
	}

	// Set GPU limits
	if request.Spec.Resources.Limits[types.ResourceGPU] != "" {
		qty, err := resource.ParseQuantity(
			string(request.Spec.Resources.Limits[types.ResourceGPU]))
		if err != nil {
			return resources, err
		}
		resources.Limits[consts.ResourceNvidiaGPU] = qty
	}

	if request.Spec.Resources.Requests[types.ResourceGPU] != "" {
		qty, err := resource.ParseQuantity(
			string(request.Spec.Resources.Requests[types.ResourceGPU]))
		if err != nil {
			return resources, err
		}
		resources.Requests[consts.ResourceNvidiaGPU] = qty
	}

	return resources, nil
}
