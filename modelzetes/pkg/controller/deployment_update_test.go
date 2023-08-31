package controller

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v2alpha1 "github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	. "github.com/tensorchord/openmodelz/modelzetes/pkg/pointer"
)

func Test_Deployment_Need_Update(t *testing.T) {
	scenarios := []struct {
		name      string
		inference *v2alpha1.Inference
		deploy    *appsv1.Deployment
		expected  bool
	}{
		{
			"empty deployment need update",
			&v2alpha1.Inference{},
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annotationInferenceSpec: "",
					},
				},
			},
			true,
		},
		{
			"bad deployment need update",
			&v2alpha1.Inference{},
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{annotationInferenceSpec: "bad"},
				},
			},
			true,
		},
		{
			"equal deployment doesn't need update",
			&v2alpha1.Inference{},
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annotationInferenceSpec: "{\"metadata\":{\"creationTimestamp\":null},\"spec\":{\"name\":\"\",\"image\":\"\"}}",
					},
				},
			},
			false,
		},
		{
			"unequal deployment need update",
			&v2alpha1.Inference{
				Spec: v2alpha1.InferenceSpec{
					Scaling: &v2alpha1.ScalingConfig{
						MinReplicas: Ptr(int32(2)),
					},
				},
			},
			&appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						annotationInferenceSpec: "{\"metadata\":{\"creationTimestamp\":null},\"spec\":{\"name\":\"\",\"image\":\"\"}}",
					},
				}},
			true,
		},
	}

	for _, s := range scenarios {
		t.Run(s.name, func(t *testing.T) {
			needUpdate := deploymentNeedsUpdate(s.inference, s.deploy)
			if needUpdate != s.expected {
				t.Errorf("incorrect judgement of need update: expected %v, got %v", s.expected, needUpdate)
				t.Fail()
			}
		})
	}
}
