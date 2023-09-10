package runtime

import (
	"context"
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/errdefs"
	"github.com/tensorchord/openmodelz/agent/pkg/k8s"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
)

func (r generalRuntime) BuildList(ctx context.Context, namespace string) (
	[]types.Build, error) {
	res := []types.Build{}
	jobs, err := r.kubeClient.BatchV1().Jobs(namespace).
		List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=true", consts.AnnotationBuilding),
		})
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, errdefs.System(err)
		}
	}

	if jobs != nil {
		for _, job := range jobs.Items {
			build, err := k8s.AsBuild(job)
			if err != nil {
				return nil, errdefs.System(err)
			}

			res = append(res, build)
		}
	}
	return res, nil
}

func (r generalRuntime) BuildCreate(ctx context.Context,
	req types.Build, inference *v2alpha1.Inference, builderImage, buildkitdAddress, buildCtlBin, secret string) error {
	buildJob, err := k8s.MakeBuild(req, inference, builderImage,
		buildkitdAddress, buildCtlBin, secret)
	if err != nil {
		return errdefs.System(err)
	}

	if _, err := r.kubeClient.BatchV1().Jobs(req.Spec.Namespace).
		Create(ctx, buildJob, metav1.CreateOptions{}); err != nil {
		return errdefs.System(err)
	}

	return nil
}

func (r generalRuntime) BuildGet(ctx context.Context, namespace, buildName string) (types.Build, error) {
	job, err := r.kubeClient.BatchV1().Jobs(namespace).Get(ctx,
		buildName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return types.Build{}, errdefs.NotFound(err)
		}
		return types.Build{}, errdefs.System(err)
	}

	res, err := k8s.AsBuild(*job)
	if err != nil {
		return types.Build{}, errdefs.System(err)
	}
	return res, nil
}
