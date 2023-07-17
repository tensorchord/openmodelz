package runtime

import (
	"context"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/pkg/k8s"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r Runtime) ServerList(ctx context.Context) ([]types.Server, error) {
	nodes, err := r.kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if len(nodes.Items) == 0 {
		return nil, nil
	}

	return getServers(nodes.Items), nil
}

func getServers(nodes []v1.Node) []types.Server {
	res := []types.Server{}
	for _, n := range nodes {
		node := types.Server{
			Spec: types.ServerSpec{
				Name: n.Name,
			},
			Status: types.ServerStatus{
				Allocatable: k8s.AsResourceList(n.Status.Allocatable),
				Capacity:    k8s.AsResourceList(n.Status.Capacity),
				Phase:       string(n.Status.Phase),
				System: types.NodeSystemInfo{
					MachineID:       n.Status.NodeInfo.MachineID,
					KernelVersion:   n.Status.NodeInfo.KernelVersion,
					OSImage:         n.Status.NodeInfo.OSImage,
					OperatingSystem: n.Status.NodeInfo.OperatingSystem,
					Architecture:    n.Status.NodeInfo.Architecture,
				},
			},
		}

		res = append(res, node)
	}
	return res
}
