package controller

import (
	"fmt"
	"testing"

	v2alpha1 "github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func Test_UpdateSecrets_DoesNotAddVolumeIfRequestSecretsIsNil(t *testing.T) {
	request := &v2alpha1.Inference{
		Spec: v2alpha1.InferenceSpec{
			Name:    "testfunc",
			Secrets: nil,
		},
	}
	existingSecrets := map[string]*corev1.Secret{
		"pullsecret": {Type: corev1.SecretTypeDockercfg},
		"testsecret": {Type: corev1.SecretTypeOpaque, Data: map[string][]byte{"filename": []byte("contents")}},
	}

	deployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "testfunc", Image: "alpine:latest"},
					},
				},
			},
		},
	}
	err := UpdateSecrets(request, deployment, existingSecrets)
	if err != nil {
		t.Errorf("unexpected error %s", err.Error())
	}

	validateEmptySecretVolumesAndMounts(t, deployment)

}

func Test_UpdateSecrets_DoesNotAddVolumeIfRequestSecretsIsEmpty(t *testing.T) {
	request := &v2alpha1.Inference{
		Spec: v2alpha1.InferenceSpec{
			Name:    "testfunc",
			Secrets: []string{},
		},
	}
	existingSecrets := map[string]*corev1.Secret{
		"pullsecret": {Type: corev1.SecretTypeDockercfg},
		"testsecret": {Type: corev1.SecretTypeOpaque, Data: map[string][]byte{"filename": []byte("contents")}},
	}

	deployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "testfunc", Image: "alpine:latest"},
					},
				},
			},
		},
	}
	err := UpdateSecrets(request, deployment, existingSecrets)
	if err != nil {
		t.Errorf("unexpected error %s", err.Error())
	}

	validateEmptySecretVolumesAndMounts(t, deployment)

}

func Test_UpdateSecrets_RemovesAllCopiesOfExitingSecretsVolumes(t *testing.T) {
	volumeName := "testfunc-projected-secrets"
	request := &v2alpha1.Inference{
		Spec: v2alpha1.InferenceSpec{
			Name:    "testfunc",
			Secrets: []string{},
		},
	}
	existingSecrets := map[string]*corev1.Secret{
		"pullsecret": {Type: corev1.SecretTypeDockercfg},
		"testsecret": {Type: corev1.SecretTypeOpaque, Data: map[string][]byte{"filename": []byte("contents")}},
	}

	deployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "testfunc",
							Image: "alpine:latest",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name: volumeName,
								},
								{
									Name: volumeName,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: volumeName,
						},
						{
							Name: volumeName,
						},
					},
				},
			},
		},
	}
	err := UpdateSecrets(request, deployment, existingSecrets)
	if err != nil {
		t.Errorf("unexpected error %s", err.Error())
	}

	validateEmptySecretVolumesAndMounts(t, deployment)

}

func Test_UpdateSecrets_AddNewSecretVolume(t *testing.T) {
	request := &v2alpha1.Inference{
		Spec: v2alpha1.InferenceSpec{
			Name:    "testfunc",
			Secrets: []string{"pullsecret", "testsecret"},
		},
	}
	existingSecrets := map[string]*corev1.Secret{
		"pullsecret": {Type: corev1.SecretTypeDockercfg},
		"testsecret": {Type: corev1.SecretTypeOpaque, Data: map[string][]byte{"filename": []byte("contents")}},
	}

	deployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "testfunc", Image: "alpine:latest"},
					},
				},
			},
		},
	}
	err := UpdateSecrets(request, deployment, existingSecrets)
	if err != nil {
		t.Errorf("unexpected error %s", err.Error())
	}

	validateNewSecretVolumesAndMounts(t, deployment)

}

func Test_UpdateSecrets_ReplacesPreviousSecretMountWithNewMount(t *testing.T) {
	request := &v2alpha1.Inference{
		Spec: v2alpha1.InferenceSpec{
			Name:    "testfunc",
			Secrets: []string{"pullsecret", "testsecret"},
		},
	}
	existingSecrets := map[string]*corev1.Secret{
		"pullsecret": {Type: corev1.SecretTypeDockercfg},
		"testsecret": {Type: corev1.SecretTypeOpaque, Data: map[string][]byte{"filename": []byte("contents")}},
	}

	deployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "testfunc", Image: "alpine:latest"},
					},
				},
			},
		},
	}
	err := UpdateSecrets(request, deployment, existingSecrets)
	if err != nil {
		t.Errorf("unexpected error %s", err.Error())
	}
	// mimic the deployment already existing and deployed with the same secrets by running
	// UpdateSecrets twice, the first run represents the original deployment, the second run represents
	// retrieving the deployment from the k8s api and applying the update to it
	err = UpdateSecrets(request, deployment, existingSecrets)
	if err != nil {
		t.Errorf("unexpected error %s", err.Error())
	}

	validateNewSecretVolumesAndMounts(t, deployment)

}

func Test_UpdateSecrets_RemovesSecretsVolumeIfRequestSecretsIsEmptyOrNil(t *testing.T) {
	request := &v2alpha1.Inference{
		Spec: v2alpha1.InferenceSpec{
			Name:    "testfunc",
			Secrets: []string{"pullsecret", "testsecret"},
		},
	}
	existingSecrets := map[string]*corev1.Secret{
		"pullsecret": {Type: corev1.SecretTypeDockercfg},
		"testsecret": {Type: corev1.SecretTypeOpaque, Data: map[string][]byte{"filename": []byte("contents")}},
	}

	deployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "testfunc", Image: "alpine:latest"},
					},
				},
			},
		},
	}
	err := UpdateSecrets(request, deployment, existingSecrets)
	if err != nil {
		t.Errorf("unexpected error %s", err.Error())
	}

	validateNewSecretVolumesAndMounts(t, deployment)

	request = &v2alpha1.Inference{
		Spec: v2alpha1.InferenceSpec{
			Name:    "testfunc",
			Secrets: []string{},
		},
	}
	err = UpdateSecrets(request, deployment, existingSecrets)
	if err != nil {
		t.Errorf("unexpected error %s", err.Error())
	}

	validateEmptySecretVolumesAndMounts(t, deployment)
}

func validateEmptySecretVolumesAndMounts(t *testing.T, deployment *appsv1.Deployment) {
	numVolumes := len(deployment.Spec.Template.Spec.Volumes)
	if numVolumes != 0 {
		fmt.Printf("%+v", deployment.Spec.Template.Spec.Volumes)
		t.Errorf("Incorrect number of volumes: expected 0, got %d", numVolumes)
	}

	c := deployment.Spec.Template.Spec.Containers[0]
	numVolumeMounts := len(c.VolumeMounts)
	if numVolumeMounts != 0 {
		t.Errorf("Incorrect number of volumes mounts: expected 0, got %d", numVolumeMounts)
	}
}

func validateNewSecretVolumesAndMounts(t *testing.T, deployment *appsv1.Deployment) {
	numVolumes := len(deployment.Spec.Template.Spec.Volumes)
	if numVolumes != 1 {
		t.Errorf("Incorrect number of volumes: expected 1, got %d", numVolumes)
	}

	volume := deployment.Spec.Template.Spec.Volumes[0]
	if volume.Name != "testfunc-projected-secrets" {
		t.Errorf("Incorrect volume name: expected \"testfunc-projected-secrets\", got \"%s\"", volume.Name)
	}

	if volume.VolumeSource.Projected == nil {
		t.Error("Secrets volume is not a projected volume type")
	}

	if volume.VolumeSource.Projected.Sources[0].Secret.Items[0].Key != "filename" {
		t.Error("Project secret not constructed correctly")
	}

	c := deployment.Spec.Template.Spec.Containers[0]
	numVolumeMounts := len(c.VolumeMounts)
	if numVolumeMounts != 1 {
		t.Errorf("Incorrect number of volumes mounts: expected 1, got %d", numVolumeMounts)
	}

	mount := c.VolumeMounts[0]
	if mount.Name != "testfunc-projected-secrets" {
		t.Errorf("Incorrect volume mounts: expected \"testfunc-projected-secrets\", got \"%s\"", mount.Name)
	}

	if mount.MountPath != secretsMountPath {
		t.Errorf("Incorrect volume mount path: expected \"%s\", got \"%s\"", secretsMountPath, mount.MountPath)
	}
}
