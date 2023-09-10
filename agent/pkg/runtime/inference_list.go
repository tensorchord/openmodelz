package runtime

import (
	"sort"

	mv2alpha1 "github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	modelzetesv2alpha1 "github.com/tensorchord/openmodelz/modelzetes/pkg/client/listers/modelzetes/v2alpha1"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	appsv1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	v1 "k8s.io/client-go/listers/apps/v1"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/errdefs"
	"github.com/tensorchord/openmodelz/agent/pkg/k8s"
)

func (r generalRuntime) InferenceList(namespace string) ([]types.InferenceDeployment, error) {
	infLister := r.inferenceInformer.Lister()
	deploymentLister := r.deploymentInformer.Lister()

	functions, err := inferenceList(namespace, infLister,
		deploymentLister)
	if err != nil {
		return nil, err
	}

	return functions, nil
}

func inferenceList(functionNamespace string,
	infLister modelzetesv2alpha1.InferenceLister,
	deploymentLister v1.DeploymentLister) ([]types.InferenceDeployment, error) {
	functions := []types.InferenceDeployment{}

	sel := labels.NewSelector()
	req, err := labels.NewRequirement(consts.LabelInferenceName, selection.Exists, []string{})
	if err != nil {
		return functions, errdefs.NotFound(err)
	}
	onlyFunctions := sel.Add(*req)

	inferences, err := infLister.Inferences(functionNamespace).
		List(labels.Everything())
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return functions, nil
		} else {
			return functions, errdefs.System(err)
		}
	}

	deploys, err := deploymentLister.Deployments(functionNamespace).List(onlyFunctions)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return getInferences(inferences, deploys)
		} else {
			return functions, errdefs.System(err)
		}
	}

	return getInferences(inferences, deploys)
}

func getInferences(inferences []*mv2alpha1.Inference, deploys []*appsv1.Deployment) ([]types.InferenceDeployment, error) {
	sort.Slice(inferences, func(i, j int) bool {
		return (*inferences[i]).Name < (*inferences[j]).Name
	})
	sort.Slice(deploys, func(i, j int) bool {
		return (*deploys[i]).Name < (*deploys[j]).Name
	})

	res := []types.InferenceDeployment{}
	j := 0
	for i := range inferences {
		if j >= len(deploys) {
			res = append(res, *k8s.AsInferenceDeployment(inferences[i], nil))
		} else if inferences[i].Name != deploys[j].Name {
			res = append(res, *k8s.AsInferenceDeployment(inferences[i], nil))
		} else {
			res = append(res, *k8s.AsInferenceDeployment(inferences[i], deploys[j]))
			j++
		}
	}

	return res, nil
}
