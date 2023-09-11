package runtime

import (
	"context"
	"strings"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/errdefs"
	"github.com/tensorchord/openmodelz/agent/pkg/k8s"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r generalRuntime) ServerList(ctx context.Context) ([]types.Server, error) {
	nodes, err := r.kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, errdefs.NotFound(err)
		} else {
			return nil, errdefs.System(err)
		}
	}

	if len(nodes.Items) == 0 {
		return nil, nil
	}

	return getServers(nodes.Items), nil
}

func getServers(nodes []v1.Node) []types.Server {
	res := []types.Server{}
	for _, n := range nodes {
		res = append(res, getServer(n))
	}
	return res
}

func getServer(n v1.Node) types.Server {
	node := types.Server{
		Spec: types.ServerSpec{
			Name:   n.Name,
			Labels: make(map[string]string),
		},
		Status: types.ServerStatus{
			Allocatable: k8s.AsResourceList(n.Status.Allocatable),
			Capacity:    k8s.AsResourceList(n.Status.Capacity),
			System: types.NodeSystemInfo{
				MachineID:       n.Status.NodeInfo.MachineID,
				KernelVersion:   n.Status.NodeInfo.KernelVersion,
				OSImage:         n.Status.NodeInfo.OSImage,
				OperatingSystem: n.Status.NodeInfo.OperatingSystem,
				Architecture:    n.Status.NodeInfo.Architecture,
			},
		},
	}

	for k, v := range n.Labels {
		if strings.HasPrefix(k, "tensorchord.ai/") {
			node.Spec.Labels[strings.TrimPrefix(k, "tensorchord.ai/")] = v
		}
	}

	phase := "Ready"
	for _, c := range n.Status.Conditions {
		if c.Type == v1.NodeReady && c.Status != v1.ConditionTrue {
			phase = "NotReady"
		} else if c.Type == v1.NodeDiskPressure && c.Status != v1.ConditionFalse {
			phase = "DiskPressure"
		} else if c.Type == v1.NodeMemoryPressure && c.Status != v1.ConditionFalse {
			phase = "MemoryPressure"
		} else if c.Type == v1.NodePIDPressure && c.Status != v1.ConditionFalse {
			phase = "PIDPressure"
		} else if c.Type == v1.NodeNetworkUnavailable && c.Status != v1.ConditionFalse {
			phase = "NetworkUnavailable"
		}
	}
	node.Status.Phase = phase
	return node
}
