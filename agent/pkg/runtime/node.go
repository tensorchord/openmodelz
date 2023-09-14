package runtime

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/agent/pkg/consts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r generalRuntime) ListServerResource() ([]string, error) {
	resources := []string{}
	listOptions := metav1.ListOptions{
		LabelSelector: consts.LabelServerResource,
	}

	nodes, err := r.kubeClient.CoreV1().Nodes().List(context.Background(), listOptions)
	if err != nil {
		logrus.Errorf("failed to list nodes: %v", err)
		return resources, err
	}
	for _, node := range nodes.Items {
		resources = append(resources, node.Labels[consts.LabelServerResource])
	}
	return resources, nil
}
