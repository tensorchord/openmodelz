package k8s

import (
	"time"

	"github.com/pkg/errors"
	mzconsts "github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/pkg/consts"
)

func MakeBuild(req types.Build, builderImage, buildkitdAddr, buildctlBin, register, registerToken string) (*batchv1.Job, error) {
	duration, err := time.ParseDuration(req.Spec.Duration)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse duration")
	}
	seconds := int64(duration.Seconds())
	defaultBackoffLimit := int32(0)
	defaultTTLSecondsAfterFinished := int32(60 * 60 * 24 * 7) // 7 days

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Spec.Name,
			Namespace: req.Spec.Namespace,
			Labels: map[string]string{
				consts.LabelProjectID:       req.Spec.ProjectID,
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
						consts.LabelProjectID: req.Spec.ProjectID,
						consts.LabelBuildName: req.Spec.Name,
					},
				},
				Spec: corev1.PodSpec{
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
							Name:  req.Spec.Name,
							Image: builderImage,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "workspace",
									MountPath: "/workspace",
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "MODELZ_BUILD_NAME",
									Value: req.Spec.Name,
								},
								{
									Name:  "MODELZ_BUILD_PROJECT_ID",
									Value: req.Spec.ProjectID,
								},
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
									Value: req.Spec.Directory,
								},
								{
									Name:  "MODELZ_BUILDER",
									Value: string(req.Spec.Builder),
								},
								{
									Name:  "MODELZ_BUILD_ARTIFACT_IMAGE",
									Value: req.Spec.ArtifactImage,
								},
								{
									Name:  "MODELZ_BUILD_ARTIFACT_IMAGE_TAG",
									Value: req.Spec.ArtifactImageTag,
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
								{
									Name:  "MODELZ_REGISTRY",
									Value: register,
								},
								{
									Name:  "MODELZ_REGISTRY_TOKEN",
									Value: registerToken,
								},
							},
						},
					},
				},
			},
		},
	}
	return job, nil
}
