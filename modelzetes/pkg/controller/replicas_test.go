package controller

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes/fake"

	v2alpha1 "github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/k8s"
)

func intPtr(n int32) *int32 {
	return &n
}

func Test_Replicas(t *testing.T) {
	scenarios := []struct {
		name     string
		function *v2alpha1.Inference
		deploy   *appsv1.Deployment
		expected *int32
	}{
		{
			"return nil replicas when label is missing and deployment does not exist",
			&v2alpha1.Inference{},
			nil,
			nil,
		},
		{
			"return nil replicas when label is missing and deployment has no replicas",
			&v2alpha1.Inference{},
			&appsv1.Deployment{},
			nil,
		},
		{
			"return min replicas when label is present and deployment has nil replicas",
			&v2alpha1.Inference{
				Spec: v2alpha1.InferenceSpec{
					Scaling: &v2alpha1.ScalingConfig{
						MinReplicas: intPtr(2),
					},
				},
			},
			&appsv1.Deployment{Spec: appsv1.DeploymentSpec{Replicas: nil}},
			int32p(2),
		},
		{
			"return min replicas when label is present and deployment has replicas less than min",
			&v2alpha1.Inference{
				Spec: v2alpha1.InferenceSpec{
					Scaling: &v2alpha1.ScalingConfig{
						MinReplicas: intPtr(2),
					},
				},
			},
			&appsv1.Deployment{Spec: appsv1.DeploymentSpec{Replicas: int32p(1)}},
			int32p(2),
		},
		{
			"return existing replicas when label is present and deployment has more replicas than min",
			&v2alpha1.Inference{
				Spec: v2alpha1.InferenceSpec{
					Scaling: &v2alpha1.ScalingConfig{
						MinReplicas: intPtr(2),
					},
				},
			},
			&appsv1.Deployment{Spec: appsv1.DeploymentSpec{Replicas: int32p(3)}},
			int32p(3),
		},
		{
			"return existing replicas when label is missing and deployment has replicas set by HPA",
			&v2alpha1.Inference{
				Spec: v2alpha1.InferenceSpec{
					Scaling: &v2alpha1.ScalingConfig{},
				},
			}, &appsv1.Deployment{Spec: appsv1.DeploymentSpec{Replicas: int32p(3)}},
			int32p(3),
		},
		{
			"return zero replicas when label is present and deployment has zero replicas",
			&v2alpha1.Inference{
				Spec: v2alpha1.InferenceSpec{
					Scaling: &v2alpha1.ScalingConfig{
						MinReplicas: intPtr(2),
					},
				},
			}, &appsv1.Deployment{Spec: appsv1.DeploymentSpec{Replicas: int32p(0)}},
			int32p(2),
		},
	}

	factory := NewFunctionFactory(fake.NewSimpleClientset(),
		k8s.DeploymentConfig{
			LivenessProbe:  &k8s.ProbeConfig{},
			ReadinessProbe: &k8s.ProbeConfig{},
			StartupProbe:   &k8s.ProbeConfig{},
		})

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			deploy := newDeployment(s.function, s.deploy, nil, factory)
			value := deploy.Spec.Replicas

			if s.expected != nil && value != nil {
				if *s.expected != *value {
					t.Errorf("incorrect replica count: expected %v, got %v", *s.expected, *value)
				}
			} else if s.expected != value {
				t.Errorf("incorrect replica count: expected %v, got %v", s.expected, value)
			}
		})
	}
}
