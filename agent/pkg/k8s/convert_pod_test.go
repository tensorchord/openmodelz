package k8s

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	. "github.com/tensorchord/openmodelz/modelzetes/pkg/pointer"

	"github.com/tensorchord/openmodelz/agent/api/types"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("agent/pkg/k8s/convert_pod", func() {
	It("function InstanceFromPod", func() {
		tcs := []struct {
			desc   string
			pod    v1.Pod
			expect *types.InferenceDeploymentInstance
		}{
			{
				desc: "empty pod",
				pod:  v1.Pod{},
				expect: Ptr(
					types.InferenceDeploymentInstance{},
				),
			},
			{
				desc: "running pod",
				pod: v1.Pod{
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				},
				expect: Ptr(
					types.InferenceDeploymentInstance{
						Status: types.InferenceDeploymentInstanceStatus{
							Phase: types.InstancePhaseRunning,
						},
					},
				),
			},
			{
				desc: "pending pod",
				pod: v1.Pod{
					Status: v1.PodStatus{
						Phase: v1.PodPending,
					},
				},
				expect: Ptr(
					types.InferenceDeploymentInstance{
						Status: types.InferenceDeploymentInstanceStatus{
							Phase: types.InstancePhasePending,
						},
					},
				),
			},
			{
				desc: "scheduling pod",
				pod: v1.Pod{
					Status: v1.PodStatus{
						Phase: v1.PodPending,
						Conditions: []v1.PodCondition{
							{
								Type:   v1.PodScheduled,
								Status: v1.ConditionFalse,
							},
						},
					},
				},
				expect: Ptr(
					types.InferenceDeploymentInstance{
						Status: types.InferenceDeploymentInstanceStatus{
							Phase: types.InstancePhaseScheduling,
						},
					},
				),
			},
			{
				desc: "failed pod",
				pod: v1.Pod{
					Status: v1.PodStatus{
						Phase: v1.PodFailed,
					},
				},
				expect: Ptr(
					types.InferenceDeploymentInstance{
						Status: types.InferenceDeploymentInstanceStatus{
							Phase: types.InstancePhaseFailed,
						},
					},
				),
			},
			{
				desc: "succeed pod",
				pod: v1.Pod{
					Status: v1.PodStatus{
						Phase: v1.PodSucceeded,
					},
				},
				expect: Ptr(
					types.InferenceDeploymentInstance{
						Status: types.InferenceDeploymentInstanceStatus{
							Phase: types.InstancePhaseSucceeded,
						},
					},
				),
			},
			{
				desc: "unknown pod",
				pod: v1.Pod{
					Status: v1.PodStatus{
						Phase: v1.PodUnknown,
					},
				},
				expect: Ptr(
					types.InferenceDeploymentInstance{
						Status: types.InferenceDeploymentInstanceStatus{
							Phase: types.InstancePhaseUnknown,
						},
					},
				),
			},
			{
				desc: "creating pod",
				pod: v1.Pod{
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							{
								Started: Ptr(false),
							},
						},
					},
				},
				expect: Ptr(
					types.InferenceDeploymentInstance{
						Status: types.InferenceDeploymentInstanceStatus{
							Phase: types.InstancePhaseCreating,
						},
					},
				),
			},
			{
				desc: "waiting pod",
				pod: v1.Pod{
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							{
								Started: Ptr(false),
								State: v1.ContainerState{
									Waiting: Ptr(v1.ContainerStateWaiting{
										Reason: "mock-status",
									}),
								},
							},
						},
					},
				},
				expect: Ptr(
					types.InferenceDeploymentInstance{
						Status: types.InferenceDeploymentInstanceStatus{
							Phase:  types.InstancePhase("mock-status"),
							Reason: "mock-status",
						},
					},
				),
			},
			{
				desc: "initializing pod",
				pod: v1.Pod{
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							{
								Started: Ptr(false),
								State: v1.ContainerState{
									Running: Ptr(v1.ContainerStateRunning{}),
								},
							},
						},
					},
				},
				expect: Ptr(
					types.InferenceDeploymentInstance{
						Status: types.InferenceDeploymentInstanceStatus{
							Phase: types.InstancePhaseInitializing,
						},
					},
				),
			},
			{
				desc: "terminated pod",
				pod: v1.Pod{
					Status: v1.PodStatus{
						ContainerStatuses: []v1.ContainerStatus{
							{
								Started: Ptr(false),
								State: v1.ContainerState{
									Terminated: Ptr(v1.ContainerStateTerminated{}),
								},
							},
						},
					},
				},
				expect: Ptr(
					types.InferenceDeploymentInstance{
						Status: types.InferenceDeploymentInstanceStatus{
							Phase: types.InstancePhaseFailed,
						},
					},
				),
			},
		}
		for _, tc := range tcs {
			logrus.Info(tc.desc)
			value := InstanceFromPod(tc.pod)
			Expect(value).To(Equal(tc.expect))
		}
	})
})
