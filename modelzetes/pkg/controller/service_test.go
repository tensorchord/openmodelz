package controller

import (
	"strings"
	"testing"

	v2alpha1 "github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	. "github.com/tensorchord/openmodelz/modelzetes/pkg/pointer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_newService(t *testing.T) {
	inference := &v2alpha1.Inference{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubesec",
			Namespace: "mock-space",
		},
		Spec: v2alpha1.InferenceSpec{
			Name:          "kubesec",
			Image:         "docker.io/kubesec/kubesec",
			HTTPProbePath: Ptr("/"),
			Annotations:   map[string]string{},
		},
	}

	service := newService(inference)

	if !strings.Contains(service.ObjectMeta.Name, inference.ObjectMeta.Name) {
		t.Errorf("Service name %s should contains inference name %s",
			service.ObjectMeta.Name, inference.ObjectMeta.Name)
		t.Fail()
	}
	if service.ObjectMeta.Namespace != inference.ObjectMeta.Namespace {
		t.Errorf("Service namespace %s should be equal to inference namespace %s",
			service.ObjectMeta.Namespace, inference.ObjectMeta.Namespace)
		t.Fail()
	}
}
