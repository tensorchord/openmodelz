package controller

import (
	"strconv"
	"testing"

	v2alpha1 "github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	. "github.com/tensorchord/openmodelz/modelzetes/pkg/pointer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_newDeployment_FrameworkGradio(t *testing.T) {
	expectEnv := map[string]string{"GRADIO_SERVER_NAME": "0.0.0.0", "GRADIO_SERVER_PORT": "7860"}

	inference := &v2alpha1.Inference{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubesec",
		},
		Spec: v2alpha1.InferenceSpec{
			Name:          "kubesec",
			Image:         "docker.io/kubesec/kubesec",
			HTTPProbePath: Ptr("/"),
			Annotations:   map[string]string{},
			Framework:     v2alpha1.FrameworkGradio,
		},
	}

	factory := NewFunctionFactory(fake.NewSimpleClientset(), defaultK8sConfig)

	secrets := map[string]*corev1.Secret{}

	deployment := newDeployment(inference, nil, secrets, factory)

	assertEnv(t, expectEnv, deployment.Spec.Template.Spec.Containers[0].Env)
}

func Test_newDeployment_FrameworkMosec(t *testing.T) {
	expectEnv := map[string]string{"MOSEC_PORT": strconv.Itoa(defaultPort)}

	inference := &v2alpha1.Inference{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubesec",
		},
		Spec: v2alpha1.InferenceSpec{
			Name:          "kubesec",
			Image:         "docker.io/kubesec/kubesec",
			HTTPProbePath: Ptr("/"),
			Annotations:   map[string]string{},
			Framework:     v2alpha1.FrameworkMosec,
		},
	}

	factory := NewFunctionFactory(fake.NewSimpleClientset(), defaultK8sConfig)

	secrets := map[string]*corev1.Secret{}

	deployment := newDeployment(inference, nil, secrets, factory)

	assertEnv(t, expectEnv, deployment.Spec.Template.Spec.Containers[0].Env)
}

func Test_newDeployment_FrameworkStreamlit(t *testing.T) {
	expectEnv := map[string]string{
		"STREAMLIT_SERVER_ENABLE_CORS":            "false",
		"STREAMLIT_SERVER_ADDRESS":                "0.0.0.0",
		"STREAMLIT_SERVER_ENABLE_XSRF_PROTECTION": "false"}

	inference := &v2alpha1.Inference{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubesec",
		},
		Spec: v2alpha1.InferenceSpec{
			Name:          "kubesec",
			Image:         "docker.io/kubesec/kubesec",
			HTTPProbePath: Ptr("/"),
			Annotations:   map[string]string{},
			Framework:     v2alpha1.FrameworkStreamlit,
		},
	}

	factory := NewFunctionFactory(fake.NewSimpleClientset(), defaultK8sConfig)

	secrets := map[string]*corev1.Secret{}

	deployment := newDeployment(inference, nil, secrets, factory)

	assertEnv(t, expectEnv, deployment.Spec.Template.Spec.Containers[0].Env)
}
