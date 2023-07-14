// Copyright 2020 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
)

func readOnlyRootDisabled(t *testing.T, deployment *appsv1.Deployment) {
	if len(deployment.Spec.Template.Spec.Volumes) != 0 {
		t.Error("Volumes should be empty if ReadOnlyRootFilesystem is false")
	}

	if len(deployment.Spec.Template.Spec.Containers[0].VolumeMounts) != 0 {
		t.Error("VolumeMounts should be empty if ReadOnlyRootFilesystem is false")
	}
	functionContatiner := deployment.Spec.Template.Spec.Containers[0]

	if functionContatiner.SecurityContext != nil {
		if *functionContatiner.SecurityContext.ReadOnlyRootFilesystem != false {
			t.Error("ReadOnlyRootFilesystem should be false on the container SecurityContext")
		}
	}
}

func Test_configureReadOnlyRootFilesystem_Disabled_To_Disabled(t *testing.T) {
	f := mockFactory()
	deployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{Name: "testfunc", Image: "alpine:latest"},
					},
				},
			},
		},
	}

	f.ConfigureReadOnlyRootFilesystem(deployment)
	readOnlyRootDisabled(t, deployment)
}
