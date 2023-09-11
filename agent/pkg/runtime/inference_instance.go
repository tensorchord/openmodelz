package runtime

import (
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	v1 "k8s.io/client-go/listers/core/v1"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/errdefs"
	"github.com/tensorchord/openmodelz/agent/pkg/k8s"
)

func (r generalRuntime) InferenceInstanceList(namespace, inferenceName string) (
	[]types.InferenceDeploymentInstance, error) {
	return getInstances(namespace, inferenceName, r.podInformer.Lister())
}

func getInstances(functionNamespace string, functionName string,
	lister v1.PodLister) ([]types.InferenceDeploymentInstance, error) {
	instances := make([]types.InferenceDeploymentInstance, 0)

	items, err := lister.List(
		labels.SelectorFromSet(k8s.MakeLabelSelector(functionName)))
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errdefs.System(err)
	}

	for _, item := range items {
		if item != nil {
			instance := k8s.InstanceFromPod(*item)
			if instance != nil {
				instances = append(instances, *instance)
			}
		}
	}

	return instances, nil
}
