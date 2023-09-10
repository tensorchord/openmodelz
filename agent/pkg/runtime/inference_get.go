package runtime

import (
	"fmt"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/client-go/listers/apps/v1"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/errdefs"
	"github.com/tensorchord/openmodelz/agent/pkg/k8s"
	apis "github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/client/listers/modelzetes/v2alpha1"
)

func (r generalRuntime) InferenceGet(namespace, inferenceName string) (
	*types.InferenceDeployment, error) {
	return inferenceGet(namespace, inferenceName,
		r.inferenceInformer.Lister(), r.deploymentInformer.Lister())
}

func (r generalRuntime) InferenceGetCRD(namespace, name string) (*apis.Inference, error) {
	inference, err := r.inferenceInformer.Lister().Inferences(namespace).Get(name)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, errdefs.NotFound(err)
		}
		return nil, err
	}
	return inference, nil
}

// inferenceGet returns a inference or nil if not found
func inferenceGet(namespace string, inferenceName string,
	infLister v2alpha1.InferenceLister,
	lister v1.DeploymentLister) (*types.InferenceDeployment, error) {

	inference, err := infLister.Inferences(namespace).Get(inferenceName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, errdefs.NotFound(err)
		}
		return nil, err
	}

	item, err := lister.Deployments(namespace).
		Get(inferenceName)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return nil, err
		}
	}

	inf := k8s.AsInferenceDeployment(inference, item)
	if inf != nil {
		return inf, nil
	}

	return nil, fmt.Errorf("inference: %s not found", inferenceName)
}
