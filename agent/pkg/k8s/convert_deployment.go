package k8s

// Copyright 2020 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

import (
	"fmt"
	"sort"

	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

const (
	secretsMountPath             = "/var/modelz/secrets"
	secretLabel                  = "app.kubernetes.io/managed-by"
	secretLabelValue             = "modelz"
	secretsProjectVolumeNameTmpl = "%s-projected-secrets"
)

// AsInference reads a Deployment object into an InferenceDeployment, parsing the
// Deployment and Container spec into a simplified summary of the Inference.
func AsInference(item appsv1.Deployment) *types.InferenceDeployment {
	var replicas int32 = 0
	if item.Spec.Replicas != nil {
		replicas = *item.Spec.Replicas
	}

	functionContainer := item.Spec.Template.Spec.Containers[0]

	labels := item.Spec.Template.Labels
	created := item.CreationTimestamp.Time
	annotations := item.Spec.Template.Annotations

	inference := types.InferenceDeployment{
		Spec: types.InferenceDeploymentSpec{
			Name:        item.Name,
			Image:       functionContainer.Image,
			Labels:      labels,
			Annotations: annotations,
			Namespace:   item.Namespace,
			Secrets:     ReadFunctionSecretsSpec(item),
		},
		Status: types.InferenceDeploymentStatus{
			Replicas:          replicas,
			AvailableReplicas: item.Status.AvailableReplicas,
			InvocationCount:   0,
			CreatedAt:         &created,
		},
	}

	inference.Spec.Resources = AsResources(functionContainer.Resources)

	inference.Spec.EnvVars = AsEnvVar(functionContainer.Env)

	inference.Status.Phase = types.PhaseNotReady
	for _, c := range item.Status.Conditions {
		if c.Type == appsv1.DeploymentAvailable && c.Status == v1.ConditionTrue {
			inference.Status.Phase = types.PhaseReady
		} else if c.Type == appsv1.DeploymentProgressing && c.Status == v1.ConditionFalse {
			inference.Status.Phase = types.PhaseScaling
		}
	}

	if item.Spec.Replicas != nil && *item.Spec.Replicas == 0 {
		inference.Status.Phase = types.PhaseNoReplicas
	}

	if item.DeletionTimestamp != nil {
		inference.Status.Phase = types.PhaseTerminating
	}

	return &inference
}

func AsEnvVar(envs []v1.EnvVar) map[string]string {
	res := make(map[string]string)
	for _, env := range envs {
		res[env.Name] = env.Value
	}
	return res
}

func AsResources(
	requirements v1.ResourceRequirements) *types.ResourceRequirements {

	gpuLimit := requirements.Limits[consts.ResourceNvidiaGPU]
	gpuLimitPtr := &gpuLimit

	gpuRequest := requirements.Requests[consts.ResourceNvidiaGPU]
	gpuRequestsPtr := &gpuRequest

	resources := types.ResourceRequirements{
		Limits:   types.ResourceList{},
		Requests: types.ResourceList{},
	}

	if !requirements.Limits.Cpu().IsZero() {
		resources.Limits[types.ResourceCPU] = types.Quantity(
			requirements.Limits.Cpu().String())
	}
	if !requirements.Limits.Memory().IsZero() {
		resources.Limits[types.ResourceMemory] = types.Quantity(
			requirements.Limits.Memory().String())
	}
	if !gpuLimitPtr.IsZero() {
		resources.Limits[types.ResourceGPU] = types.Quantity(
			gpuLimitPtr.String())
	}
	if !requirements.Requests.Cpu().IsZero() {
		resources.Requests[types.ResourceCPU] = types.Quantity(
			requirements.Requests.Cpu().String())
	}
	if !requirements.Requests.Memory().IsZero() {
		resources.Requests[types.ResourceMemory] = types.Quantity(
			requirements.Requests.Memory().String())
	}
	if !gpuRequestsPtr.IsZero() {
		resources.Requests[types.ResourceGPU] = types.Quantity(
			gpuRequestsPtr.String())
	}

	return &resources
}

// ReadFunctionSecretsSpec parses the name of the required function secrets. This is the inverse of ConfigureSecrets.
func ReadFunctionSecretsSpec(item appsv1.Deployment) []string {
	secrets := []string{}

	for _, s := range item.Spec.Template.Spec.ImagePullSecrets {
		secrets = append(secrets, s.Name)
	}

	volumeName := fmt.Sprintf(secretsProjectVolumeNameTmpl, item.Name)
	var sourceSecrets []v1.VolumeProjection
	for _, v := range item.Spec.Template.Spec.Volumes {
		if v.Name == volumeName {
			sourceSecrets = v.Projected.Sources
			break
		}
	}

	for _, s := range sourceSecrets {
		if s.Secret == nil {
			continue
		}
		secrets = append(secrets, s.Secret.Name)
	}

	sort.Strings(secrets)
	return secrets
}
