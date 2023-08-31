package controller

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	v2alpha1 "github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/k8s"
	. "github.com/tensorchord/openmodelz/modelzetes/pkg/pointer"
)

var defaultK8sConfig = k8s.DeploymentConfig{
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

func assertEnv(t *testing.T, expect map[string]string, real []v1.EnvVar) {
	for _, env := range real {
		value, exist := expect[env.Name]
		if exist == false || value != env.Value {
			t.Errorf("Environment variables contains unexpected %s:%s", env.Name, env.Value)
			t.Fail()
		}
		delete(expect, env.Name)
	}
	if len(expect) != 0 {
		t.Errorf("Environment variables should contain %v", expect)
		t.Fail()
	}
}

func Test_newDeployment(t *testing.T) {
	inference := &v2alpha1.Inference{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubesec",
		},
		Spec: v2alpha1.InferenceSpec{
			Name:          "kubesec",
			Image:         "docker.io/kubesec/kubesec",
			HTTPProbePath: Ptr("/"),
			Annotations:   map[string]string{},
		},
	}

	factory := NewFunctionFactory(fake.NewSimpleClientset(), defaultK8sConfig)

	secrets := map[string]*corev1.Secret{}

	deployment := newDeployment(inference, nil, secrets, factory)

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
	inf := &v2alpha1.Inference{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubesec",
		},
		Spec: v2alpha1.InferenceSpec{
			Name:          "kubesec",
			Image:         "docker.io/kubesec/kubesec",
			HTTPProbePath: Ptr("/"),
			Annotations: map[string]string{
				"prometheus.io.scrape": "true",
			},
			Scaling: &v2alpha1.ScalingConfig{
				StartupDuration: Ptr(int32(600)),
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
	inference := &v2alpha1.Inference{
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

	deployment := newDeployment(inference, nil, secrets, factory)

	want := "true"

	if deployment.Spec.Template.Annotations["prometheus.io.scrape"] != want {
		t.Errorf("Annotation prometheus.io.scrape should be %s, was: %s", want, deployment.Spec.Template.Annotations["prometheus.io.scrape"])
	}
}

func Test_newDeployment_WithZeroResource(t *testing.T) {
	quantity, _ := resource.ParseQuantity("0")
	inference := &v2alpha1.Inference{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubesec",
		},
		Spec: v2alpha1.InferenceSpec{
			Name:          "kubesec",
			Image:         "docker.io/kubesec/kubesec",
			HTTPProbePath: Ptr("/"),
			Annotations:   map[string]string{},
			Resources: &v1.ResourceRequirements{
				Limits: v1.ResourceList{consts.ResourceNvidiaGPU: quantity},
			},
		},
	}

	factory := NewFunctionFactory(fake.NewSimpleClientset(), defaultK8sConfig)

	secrets := map[string]*corev1.Secret{}

	deployment := newDeployment(inference, nil, secrets, factory)

	if deployment.Spec.Template.Spec.Containers[0].Env[0].Name != "CUDA_VISIBLE_DEVICES" {
		t.Errorf("CUDA_VISIBLE_DEVICES should be set to environment variables")
		t.Fail()
	}

	if deployment.Spec.Template.Spec.Containers[0].Env[0].Value != "" {
		t.Errorf("Empty value should be set to CUDA_VISIBLE_DEVICES")
		t.Fail()
	}
}

func Test_newDeployment_WithNonZeroResource(t *testing.T) {
	quantity, _ := resource.ParseQuantity("1")
	inference := &v2alpha1.Inference{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubesec",
		},
		Spec: v2alpha1.InferenceSpec{
			Name:          "kubesec",
			Image:         "docker.io/kubesec/kubesec",
			HTTPProbePath: Ptr("/"),
			Annotations:   map[string]string{},
			Resources: &v1.ResourceRequirements{
				Limits: v1.ResourceList{consts.ResourceNvidiaGPU: quantity},
			},
		},
	}

	factory := NewFunctionFactory(fake.NewSimpleClientset(), defaultK8sConfig)

	secrets := map[string]*corev1.Secret{}

	deployment := newDeployment(inference, nil, secrets, factory)

	if deployment.Spec.Template.Spec.Tolerations[0].Key != consts.TolerationGPU {
		t.Errorf("Tolerations should contain %s", consts.TolerationGPU)
		t.Fail()
	}

	if deployment.Spec.Template.Spec.Tolerations[1].Key != consts.TolerationNvidiaGPUPresent {
		t.Errorf("Tolerations should contain %s", consts.TolerationNvidiaGPUPresent)
		t.Fail()
	}
}

func Test_newDeployment_WithCommandsAndEnvVars(t *testing.T) {
	expectEnv := map[string]string{"MOCK": "TEST"}
	expectCommand := "python main.py"

	inference := &v2alpha1.Inference{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubesec",
		},
		Spec: v2alpha1.InferenceSpec{
			Name:          "kubesec",
			Image:         "docker.io/kubesec/kubesec",
			HTTPProbePath: Ptr("/"),
			Annotations:   map[string]string{},
			Command:       Ptr(expectCommand),
			EnvVars:       expectEnv,
		},
	}

	factory := NewFunctionFactory(fake.NewSimpleClientset(), defaultK8sConfig)

	secrets := map[string]*corev1.Secret{}

	deployment := newDeployment(inference, nil, secrets, factory)

	assertEnv(t, expectEnv, deployment.Spec.Template.Spec.Containers[0].Env)

	if strings.Join(deployment.Spec.Template.Spec.Containers[0].Command, " ") != expectCommand {
		t.Errorf("Command should contain value %s", expectCommand)
		t.Fail()
	}
}
