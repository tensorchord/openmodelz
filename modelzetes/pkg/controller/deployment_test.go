package controller

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	v2alpha1 "github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/k8s"
)

func Test_newDeployment(t *testing.T) {
	var httpProbePath string = "/"
	function := &v2alpha1.Inference{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubesec",
		},
		Spec: v2alpha1.InferenceSpec{
			Name:          "kubesec",
			Image:         "docker.io/kubesec/kubesec",
			HTTPProbePath: &httpProbePath,
			Annotations:   map[string]string{},
		},
	}
	k8sConfig := k8s.DeploymentConfig{
		HTTPProbe:      true,
		SetNonRootUser: true,
		LivenessProbe: &k8s.ProbeConfig{
			PeriodSeconds:       1,
			TimeoutSeconds:      3,
			InitialDelaySeconds: 0,
		},
		ReadinessProbe: &k8s.ProbeConfig{
			PeriodSeconds:       1,
			TimeoutSeconds:      3,
			InitialDelaySeconds: 0,
		},
		StartupProbe: &k8s.ProbeConfig{
			PeriodSeconds:       1,
			TimeoutSeconds:      3,
			InitialDelaySeconds: 0,
		},
	}
	factory := NewFunctionFactory(fake.NewSimpleClientset(), k8sConfig)

	secrets := map[string]*corev1.Secret{}

	deployment := newDeployment(function, nil, secrets, factory)

	if deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Path != "/" {
		t.Errorf("Readiness probe should have HTTPGet handler set to %s", "/")
		t.Fail()
	}

	if deployment.Spec.Template.Spec.Containers[0].StartupProbe.InitialDelaySeconds != 0 {
		t.Errorf("Startup probe should have initial delay seconds set to %s", "0")
		t.Fail()
	}

	if deployment.Spec.Template.Spec.Containers[0].LivenessProbe.InitialDelaySeconds != 0 {
		t.Errorf("Liveness probe should have initial delay seconds set to %s", "0")
		t.Fail()
	}

	if *(deployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser) != k8s.SecurityContextUserID {
		t.Errorf("RunAsUser should be %v", k8s.SecurityContextUserID)
		t.Fail()
	}
}

func TestNewDeploymentWithStartupDurationLabel(t *testing.T) {
	var duration int32 = 600
	var httpProbePath string = "/"
	inf := &v2alpha1.Inference{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubesec",
		},
		Spec: v2alpha1.InferenceSpec{
			Name:          "kubesec",
			Image:         "docker.io/kubesec/kubesec",
			HTTPProbePath: &httpProbePath,
			Annotations: map[string]string{
				"prometheus.io.scrape": "true",
			},
			Scaling: &v2alpha1.ScalingConfig{
				StartupDuration: &duration,
			},
		},
	}

	factory := NewFunctionFactory(fake.NewSimpleClientset(),
		k8s.DeploymentConfig{
			HTTPProbe:      true,
			SetNonRootUser: true,
			LivenessProbe: &k8s.ProbeConfig{
				PeriodSeconds:       1,
				TimeoutSeconds:      3,
				InitialDelaySeconds: 0,
			},
			ReadinessProbe: &k8s.ProbeConfig{
				PeriodSeconds:       1,
				TimeoutSeconds:      3,
				InitialDelaySeconds: 0,
			},
			StartupProbe: &k8s.ProbeConfig{
				PeriodSeconds:       10,
				TimeoutSeconds:      3,
				InitialDelaySeconds: 0,
			},
		})

	secrets := map[string]*corev1.Secret{}

	expectedPeriodSeconds := int32(10)
	expectedFailureThreshold := int32(60)
	deployment := newDeployment(inf, nil, secrets, factory)
	if len(deployment.Spec.Template.Spec.Containers) == 0 {
		t.Errorf("Deployment should have at least one container")
		t.Fail()
	}
	if deployment.Spec.Template.Spec.Containers[0].StartupProbe == nil {
		t.Errorf("Deployment should have a startup probe")
		t.Fail()
	}
	if deployment.Spec.Template.Spec.Containers[0].StartupProbe.PeriodSeconds != expectedPeriodSeconds {
		t.Errorf("Startup probe should have timeout seconds set to %d", expectedPeriodSeconds)
		t.Fail()
	}
	if deployment.Spec.Template.Spec.Containers[0].StartupProbe.FailureThreshold != expectedFailureThreshold {
		t.Errorf("Startup probe should have failure threshold set to %d", expectedFailureThreshold)
		t.Fail()
	}
}

func Test_newDeployment_PrometheusScrape_NotOverridden(t *testing.T) {
	function := &v2alpha1.Inference{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubesec",
		},
		Spec: v2alpha1.InferenceSpec{
			Name:  "kubesec",
			Image: "docker.io/kubesec/kubesec",
			Annotations: map[string]string{
				"prometheus.io.scrape": "true",
			},
		},
	}

	factory := NewFunctionFactory(fake.NewSimpleClientset(),
		k8s.DeploymentConfig{
			HTTPProbe:      false,
			SetNonRootUser: true,
			LivenessProbe: &k8s.ProbeConfig{
				PeriodSeconds:       1,
				TimeoutSeconds:      3,
				InitialDelaySeconds: 0,
			},
			ReadinessProbe: &k8s.ProbeConfig{
				PeriodSeconds:       1,
				TimeoutSeconds:      3,
				InitialDelaySeconds: 0,
			},
			StartupProbe: &k8s.ProbeConfig{
				PeriodSeconds:       1,
				TimeoutSeconds:      3,
				InitialDelaySeconds: 0,
			},
		})

	secrets := map[string]*corev1.Secret{}

	deployment := newDeployment(function, nil, secrets, factory)

	want := "true"

	if deployment.Spec.Template.Annotations["prometheus.io.scrape"] != want {
		t.Errorf("Annotation prometheus.io.scrape should be %s, was: %s", want, deployment.Spec.Template.Annotations["prometheus.io.scrape"])
	}
}
