package k8s

import (
	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	v1 "k8s.io/api/core/v1"

	"github.com/tensorchord/openmodelz/agent/api/types"
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
			Reason:  pod.Status.Reason,
			Message: pod.Status.Message,
		},
	}

	if pod.Status.StartTime != nil {
		i.Status.StartTime = pod.Status.StartTime.Time
	}

	switch pod.Status.Phase {
	case v1.PodRunning:
		i.Status.Phase = types.InstancePhaseRunning
	case v1.PodPending:
		i.Status.Phase = types.InstancePhasePending
	case v1.PodFailed:
		i.Status.Phase = types.InstancePhaseFailed
	case v1.PodSucceeded:
		i.Status.Phase = types.InstancePhaseSucceeded
	case v1.PodUnknown:
		i.Status.Phase = types.InstancePhaseUnknown
	}

	if pod.Status.Conditions != nil {
		for _, c := range pod.Status.Conditions {
			if c.Type == v1.PodScheduled && c.Status == v1.ConditionFalse {
				i.Status.Phase = types.InstancePhaseScheduling
				i.Status.Reason = c.Reason
				i.Status.Message = c.Message
				break
			}
		}
	}

	if len(pod.Status.ContainerStatuses) != 0 {
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
			} else if pod.Status.ContainerStatuses[0].State.Terminated != nil {
				i.Status.Phase = types.InstancePhaseFailed
				i.Status.Reason = pod.Status.ContainerStatuses[0].State.Terminated.Reason
				i.Status.Message = pod.Status.ContainerStatuses[0].State.Terminated.Message
			}
		}
	}
	return i
}
