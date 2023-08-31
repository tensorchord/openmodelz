package k8s

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	types "github.com/tensorchord/openmodelz/agent/api/types"
	. "github.com/tensorchord/openmodelz/modelzetes/pkg/pointer"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	mock_time, _ = time.Parse("2006-01-02", "2023-08-31")
)

func Test_InstanceFromPod(t *testing.T) {
	scenarios := []struct {
		name     string
		pod      v1.Pod
		expected types.InferenceDeploymentInstance
	}{
		{
			"basic pod",
			v1.Pod{
				Status: v1.PodStatus{
					StartTime: Ptr(metav1.NewTime(mock_time)),
					ContainerStatuses: []v1.ContainerStatus{
						{},
					},
				},
			},
			types.InferenceDeploymentInstance{
				Status: types.InferenceDeploymentInstanceStatus{
					StartTime: mock_time,
				},
			},
		},
		{
			"phase running pod",
			v1.Pod{
				Status: v1.PodStatus{
					Phase:     v1.PodRunning,
					StartTime: Ptr(metav1.NewTime(mock_time)),
					ContainerStatuses: []v1.ContainerStatus{
						{},
					},
				},
			},
			types.InferenceDeploymentInstance{
				Status: types.InferenceDeploymentInstanceStatus{
					Phase:     types.InstancePhaseRunning,
					StartTime: mock_time,
				},
			},
		},
		{
			"phase pending pod",
			v1.Pod{
				Status: v1.PodStatus{
					Phase:     v1.PodPending,
					StartTime: Ptr(metav1.NewTime(mock_time)),
					ContainerStatuses: []v1.ContainerStatus{
						{},
					},
				},
			},
			types.InferenceDeploymentInstance{
				Status: types.InferenceDeploymentInstanceStatus{
					Phase:     types.InstancePhasePending,
					StartTime: mock_time,
				},
			},
		},
		{
			"phase scheduling pod",
			v1.Pod{
				Status: v1.PodStatus{
					Phase:     v1.PodPending,
					StartTime: Ptr(metav1.NewTime(mock_time)),
					ContainerStatuses: []v1.ContainerStatus{
						{},
					},
					Conditions: []v1.PodCondition{
						{
							Type:   v1.PodScheduled,
							Status: v1.ConditionFalse,
						},
					},
				},
			},
			types.InferenceDeploymentInstance{
				Status: types.InferenceDeploymentInstanceStatus{
					Phase:     types.InstancePhaseScheduling,
					StartTime: mock_time,
				},
			},
		},
		{
			"phase failed pod",
			v1.Pod{
				Status: v1.PodStatus{
					Phase:     v1.PodFailed,
					StartTime: Ptr(metav1.NewTime(mock_time)),
					ContainerStatuses: []v1.ContainerStatus{
						{},
					},
				},
			},
			types.InferenceDeploymentInstance{
				Status: types.InferenceDeploymentInstanceStatus{
					Phase:     types.InstancePhaseFailed,
					StartTime: mock_time,
				},
			},
		},
		{
			"phase succeed pod",
			v1.Pod{
				Status: v1.PodStatus{
					Phase:     v1.PodSucceeded,
					StartTime: Ptr(metav1.NewTime(mock_time)),
					ContainerStatuses: []v1.ContainerStatus{
						{},
					},
				},
			},
			types.InferenceDeploymentInstance{
				Status: types.InferenceDeploymentInstanceStatus{
					Phase:     types.InstancePhaseSucceeded,
					StartTime: mock_time,
				},
			},
		},
		{
			"phase unknown pod",
			v1.Pod{
				Status: v1.PodStatus{
					Phase:     v1.PodUnknown,
					StartTime: Ptr(metav1.NewTime(mock_time)),
					ContainerStatuses: []v1.ContainerStatus{
						{},
					},
				},
			},
			types.InferenceDeploymentInstance{
				Status: types.InferenceDeploymentInstanceStatus{
					Phase:     types.InstancePhaseUnknown,
					StartTime: mock_time,
				},
			},
		},
		{
			"phase creating pod",
			v1.Pod{
				Status: v1.PodStatus{
					Phase:     v1.PodUnknown,
					StartTime: Ptr(metav1.NewTime(mock_time)),
					ContainerStatuses: []v1.ContainerStatus{
						{
							Started: Ptr(false),
						},
					},
				},
			},
			types.InferenceDeploymentInstance{
				Status: types.InferenceDeploymentInstanceStatus{
					Phase:     types.InstancePhaseCreating,
					StartTime: mock_time,
				},
			},
		},
		{
			"phase initializing pod",
			v1.Pod{
				Status: v1.PodStatus{
					Phase:     v1.PodUnknown,
					StartTime: Ptr(metav1.NewTime(mock_time)),
					ContainerStatuses: []v1.ContainerStatus{
						{
							Started: Ptr(false),
							State: v1.ContainerState{
								Running: Ptr(v1.ContainerStateRunning{
									StartedAt: metav1.NewTime(mock_time),
								}),
							},
						},
					},
				},
			},
			types.InferenceDeploymentInstance{
				Status: types.InferenceDeploymentInstanceStatus{
					Phase:     types.InstancePhaseInitializing,
					StartTime: mock_time,
				},
			},
		},
		{
			"phase waiting pod",
			v1.Pod{
				Status: v1.PodStatus{
					Phase:     v1.PodUnknown,
					StartTime: Ptr(metav1.NewTime(mock_time)),
					ContainerStatuses: []v1.ContainerStatus{
						{
							Started: Ptr(false),
							State: v1.ContainerState{
								Waiting: Ptr(v1.ContainerStateWaiting{
									Reason:  "mock-reason",
									Message: "mock-message",
								}),
							},
						},
					},
				},
			},
			types.InferenceDeploymentInstance{
				Status: types.InferenceDeploymentInstanceStatus{
					Phase:     types.InstancePhase("mock-reason"),
					Reason:    "mock-reason",
					Message:   "mock-message",
					StartTime: mock_time,
				},
			},
		},
	}
	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			instance := InstanceFromPod(s.pod)
			if diff := cmp.Diff(s.expected, *instance); diff != "" {
				t.Errorf("Create instance from pod: expected %v, got %v", s.expected, instance)
				t.Fail()
			}
		})
	}
}
