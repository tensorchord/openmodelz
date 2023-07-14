package k8s

import (
	types "github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	v1 "k8s.io/api/core/v1"
)

func MakeLabelSelector(name string) map[string]string {
	return map[string]string{
		"app": name,
	}
}

func InstanceFromPod(pod v1.Pod) *types.InferenceDeploymentInstance {
	i := &types.InferenceDeploymentInstance{
		Spec: types.InferenceDeploymentInstanceSpec{
			Namespace:      pod.Namespace,
			Name:           pod.Name,
			OwnerReference: pod.Labels[consts.LabelInferenceName],
		},
		Status: types.InferenceDeploymentInstanceStatus{
			StartTime: pod.Status.StartTime.Time,
			Reason:    pod.Status.Reason,
			Message:   pod.Status.Message,
		},
	}

	switch pod.Status.Phase {
	case v1.PodRunning:
		i.Status.Phase = types.InstancePhaseRunning
	case v1.PodPending:
		for _, c := range pod.Status.Conditions {
			if c.Type == v1.PodScheduled && c.Status == v1.ConditionFalse {
				i.Status.Phase = types.InstancePhaseScheduling
				break
			}
		}
		i.Status.Phase = types.InstancePhasePending
	case v1.PodFailed:
		i.Status.Phase = types.InstancePhaseFailed
	case v1.PodSucceeded:
		i.Status.Phase = types.InstancePhaseSucceeded
	case v1.PodUnknown:
		i.Status.Phase = types.InstancePhaseUnknown
	}

	if pod.Status.ContainerStatuses[0].Started != nil &&
		!*pod.Status.ContainerStatuses[0].Started {
		i.Status.Phase = types.InstancePhaseCreating
		if pod.Status.ContainerStatuses[0].State.Waiting != nil {
			i.Status.Reason = pod.Status.ContainerStatuses[0].State.Waiting.Reason
			i.Status.Message = pod.Status.ContainerStatuses[0].State.Waiting.Message
			i.Status.Phase = types.InstancePhase(
				pod.Status.ContainerStatuses[0].State.Waiting.Reason)
		} else if pod.Status.ContainerStatuses[0].State.Running != nil {
			i.Status.Phase = types.InstancePhaseInitializing
		}
	}
	return i
}
