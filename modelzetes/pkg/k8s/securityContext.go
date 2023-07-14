// Copyright 2020 OpenFaaS Authors
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// nonRootFunctionuserID is the user id that is set when DeployHandlerConfig.SetNonRootUser is true.
// value >10000 per the suggestion from https://kubesec.io/basics/containers-securitycontext-runasuser/
const SecurityContextUserID = int64(12000)

// ConfigureContainerUserID sets the UID to 12000 for the function Container.  Defaults to user
// specified in image metadata if `SetNonRootUser` is `false`. Root == 0.
func (f *FunctionFactory) ConfigureContainerUserID(deployment *appsv1.Deployment) {
	userID := SecurityContextUserID
	var functionUser *int64

	if f.Config.SetNonRootUser {
		functionUser = &userID
	}

	if deployment.Spec.Template.Spec.Containers[0].SecurityContext == nil {
		deployment.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{}
	}

	deployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser = functionUser
}

// ConfigureReadOnlyRootFilesystem will create or update the required settings and mounts to ensure
// that the ReadOnlyRootFilesystem setting works as expected, meaning:
//  1. when ReadOnlyRootFilesystem is true, the security context of the container will have ReadOnlyRootFilesystem also
//     marked as true and a new `/tmp` folder mount will be added to the deployment spec
//  2. when ReadOnlyRootFilesystem is false, the security context of the container will also have ReadOnlyRootFilesystem set
//     to false and there will be no mount for the `/tmp` folder
//
// This method is safe for both create and update operations.
func (f *FunctionFactory) ConfigureReadOnlyRootFilesystem(deployment *appsv1.Deployment) {
	readonly := false
	if deployment.Spec.Template.Spec.Containers[0].SecurityContext != nil {
		deployment.Spec.Template.Spec.Containers[0].SecurityContext.ReadOnlyRootFilesystem = &readonly
	} else {
		deployment.Spec.Template.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
			ReadOnlyRootFilesystem: &readonly,
		}
	}

	existingVolumes := removeVolume("temp", deployment.Spec.Template.Spec.Volumes)
	deployment.Spec.Template.Spec.Volumes = existingVolumes

	existingMounts := removeVolumeMount("temp", deployment.Spec.Template.Spec.Containers[0].VolumeMounts)
	deployment.Spec.Template.Spec.Containers[0].VolumeMounts = existingMounts
}
