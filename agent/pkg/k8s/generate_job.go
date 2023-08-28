package k8s

import (
	"time"

	"github.com/cockroachdb/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/pkg/consts"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	mzconsts "github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
)

func MakeBuild(req types.Build, inference *v2alpha1.Inference, builderImage, buildkitdAddr, buildctlBin, secret string) (*batchv1.Job, error) {
	job := &batchv1.Job{}
	duration, err := time.ParseDuration(req.Spec.BuildTarget.Duration)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse duration")
	}
	seconds := int64(duration.Seconds())
	defaultBackoffLimit := int32(0)
	defaultTTLSecondsAfterFinished := int32(60 * 60 * 24 * 7) // 7 days

	envs := []corev1.EnvVar{
		{
			Name:  "MODELZ_BUILD_NAME",
			Value: req.Spec.Name,
		},
		{
			Name:  "MODELZ_BUILDER",
			Value: string(req.Spec.BuildTarget.Builder),
		},
		{
			Name:  "MODELZ_BUILD_ARTIFACT_IMAGE",
			Value: req.Spec.BuildTarget.ArtifactImage,
		},
		{
			Name:  "MODELZ_BUILD_ARTIFACT_IMAGE_TAG",
			Value: req.Spec.BuildTarget.ArtifactImageTag,
		},
		{
			Name:  "MODELZ_REGISTRY",
			Value: req.Spec.BuildTarget.Registry,
		},
		{
			Name:  "MODELZ_REGISTRY_TOKEN",
			Value: req.Spec.BuildTarget.RegistryToken,
		},
	}
	if req.Spec.BuildTarget.Builder != types.BuilderTypeImage {
		envs = append(envs, buildEnvsForDockerfileOrEnvd(req, buildkitdAddr, buildctlBin)...)
	} else {
		envs = append(envs, buildEnvsForImage(req)...)
	}

	ownerReference := []metav1.OwnerReference{
		*metav1.NewControllerRef(inference, schema.GroupVersionKind{
			Group:   v2alpha1.SchemeGroupVersion.Group,
			Version: v2alpha1.SchemeGroupVersion.Version,
			Kind:    v2alpha1.Kind,
		}),
	}
	job = &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:            req.Spec.Name,
			Namespace:       req.Spec.Namespace,
			OwnerReferences: ownerReference,
			Labels: map[string]string{
				consts.LabelBuildName:       req.Spec.Name,
				mzconsts.AnnotationBuilding: "true",
			},
		},
		Spec: batchv1.JobSpec{
			ActiveDeadlineSeconds:   &seconds,
			BackoffLimit:            &defaultBackoffLimit,
			TTLSecondsAfterFinished: &defaultTTLSecondsAfterFinished,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						consts.LabelBuildName: req.Spec.Name,
					},
				},
				Spec: corev1.PodSpec{
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: secret},
					},
					Volumes: []corev1.Volume{
						{
							Name: "workspace",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:            req.Spec.Name,
							Image:           builderImage,
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "workspace",
									MountPath: "/workspace",
								},
							},
							Env: envs,
						},
					},
				},
			},
		},
	}

	return job, nil
}

func buildEnvsForImage(req types.Build) []corev1.EnvVar {
	envs := []corev1.EnvVar{}
	if req.Spec.DockerSource.AuthN.Username != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "MODELZ_SOURCE_REGISTRY_USERNAME",
			Value: req.Spec.DockerSource.AuthN.Username,
		})
	}

	if req.Spec.DockerSource.AuthN.Password != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "MODELZ_SOURCE_REGISTRY_PASSWORD",
			Value: req.Spec.DockerSource.AuthN.Password,
		})
	}

	if req.Spec.DockerSource.AuthN.Token != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "MODELZ_SOURCE_REGISTRY_TOKEN",
			Value: req.Spec.DockerSource.AuthN.Token,
		})
	}

	if req.Spec.DockerSource.ArtifactImage != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "MODELZ_SOURCE_REGISTRY_IMAGE",
			Value: req.Spec.DockerSource.ArtifactImage,
		})
	}

	if req.Spec.DockerSource.ArtifactImageTag != "" {
		envs = append(envs, corev1.EnvVar{
			Name:  "MODELZ_SOURCE_REGISTRY_IMAGE_TAG",
			Value: req.Spec.DockerSource.ArtifactImageTag,
		})
	}

	return envs
}

func buildEnvsForDockerfileOrEnvd(req types.Build, buildkitdAddr, buildctlBin string) []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  "MODELZ_BUILD_GIT_URL",
			Value: req.Spec.Repository,
		},
		{
			Name:  "MODELZ_BUILD_GIT_BRANCH",
			Value: req.Spec.Branch,
		},
		{
			Name:  "MODELZ_BUILD_GIT_COMMIT",
			Value: req.Spec.Revision,
		},
		{
			Name:  "MODELZ_BUILD_BASE_DIR",
			Value: req.Spec.BuildTarget.Directory,
		},
		{
			Name:  "MODELZ_WORKSPACE",
			Value: "/workspace",
		},
		{
			Name:  "MODELZ_BUILDKITD_ADDRESS",
			Value: buildkitdAddr,
		},
		{
			Name:  "MODELZ_BUILDER_BIN",
			Value: buildctlBin,
		},
	}
}
