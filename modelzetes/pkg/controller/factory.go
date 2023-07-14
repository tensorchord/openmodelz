package controller

import (
	"github.com/tensorchord/openmodelz/agent/api/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	v2alpha1 "github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/k8s"
)

// FunctionFactory wraps modelzetes factory
type FunctionFactory struct {
	Factory k8s.FunctionFactory
}

func NewFunctionFactory(clientset kubernetes.Interface, config k8s.DeploymentConfig) FunctionFactory {
	return FunctionFactory{
		k8s.FunctionFactory{
			Client: clientset,
			Config: config,
		},
	}
}

func functionToResourceRequirements(in *v2alpha1.Inference) types.ResourceRequirements {
	resources := types.ResourceRequirements{}

	if in.Spec.Resources == nil {
		return resources
	}

	gpuLimit := in.Spec.Resources.Limits[consts.ResourceNvidiaGPU]
	gpuLimitPtr := &gpuLimit

	gpuRequest := in.Spec.Resources.Requests[consts.ResourceNvidiaGPU]
	gpuRequestsPtr := &gpuRequest

	resources = types.ResourceRequirements{
		Limits: types.ResourceList{
			types.ResourceCPU: types.Quantity(
				in.Spec.Resources.Limits.Cpu().String()),
			types.ResourceMemory: types.Quantity(
				in.Spec.Resources.Limits.Memory().String()),
			types.ResourceGPU: types.Quantity(gpuLimitPtr.String()),
		},
		Requests: types.ResourceList{
			types.ResourceCPU: types.Quantity(
				in.Spec.Resources.Requests.Cpu().String()),
			types.ResourceMemory: types.Quantity(
				in.Spec.Resources.Requests.Memory().String()),
			types.ResourceGPU: types.Quantity(gpuRequestsPtr.String()),
		},
	}

	return resources
}

func (f *FunctionFactory) MakeHuggingfacePullThroughCacheEnvVar() *corev1.EnvVar {
	if f.Factory.Config.HuggingfacePullThroughCache {
		return &corev1.EnvVar{
			Name:  "HF_ENDPOINT",
			Value: f.Factory.Config.HuggingfacePullThroughCacheEndpoint,
		}
	}

	return nil
}

func (f *FunctionFactory) MakeProbes(function *v2alpha1.Inference, port int) (
	*k8s.FunctionProbes, error) {
	// For old version inference without HTTPProbePath
	httpProbePath := consts.DefaultHTTPProbePath
	if (function.Spec.HTTPProbePath != nil) && (*function.Spec.HTTPProbePath != "") {
		httpProbePath = *function.Spec.HTTPProbePath
	}

	return f.Factory.MakeProbes(port, httpProbePath)
}

func (f *FunctionFactory) ConfigureReadOnlyRootFilesystem(function *v2alpha1.Inference, deployment *appsv1.Deployment) {
	f.Factory.ConfigureReadOnlyRootFilesystem(deployment)
}

func (f *FunctionFactory) ConfigureContainerUserID(deployment *appsv1.Deployment) {
	f.Factory.ConfigureContainerUserID(deployment)
}
