package controller

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/google/go-cmp/cmp"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	glog "k8s.io/klog"

	v2alpha1 "github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/k8s"
)

const (
	annotationInferenceSpec = "ai.tensorchord.inference.spec"
	defaultPort             = 8080
)

var runtimeClassNvidia = "nvidia"

// newDeployment creates a new Deployment for a Function resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Function resource that 'owns' it.
func newDeployment(
	inference *v2alpha1.Inference,
	existingDeployment *appsv1.Deployment,
	existingSecrets map[string]*corev1.Secret,
	factory FunctionFactory) *appsv1.Deployment {

	// Set replicas to 0 if the expected number of replicas is 0
	replicas := getReplicas(inference, existingDeployment)

	envVars := makeEnvVars(inference)
	labels := makeLabels(inference)
	nodeSelector := makeNodeSelector(inference.Spec.Constraints)

	port := makePort(inference)
	probes, err := factory.MakeProbes(inference, port)
	if err != nil {
		glog.Warningf("Function %s probes parsing failed: %v",
			inference.Spec.Name, err)
	}
	labelMap := k8s.MakeLabelSelector(inference.Spec.Name)
	// Add a new env var HF_ENDPOINT if enabled.
	hfEnvs := factory.MakeHuggingfacePullThroughCacheEnvVar()
	if hfEnvs != nil {
		envVars = addEnvVarIfNotExists(envVars, hfEnvs.Name, hfEnvs.Value)
	}

	annotations := makeAnnotations(inference)

	command := makeCommand(inference)

	allowPrivilegeEscalation := false

	deploymentSpec := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        inference.Spec.Name,
			Annotations: annotations,
			Namespace:   inference.Namespace,
			Labels:      labels,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(inference, schema.GroupVersionKind{
					Group:   v2alpha1.SchemeGroupVersion.Group,
					Version: v2alpha1.SchemeGroupVersion.Version,
					Kind:    v2alpha1.Kind,
				}),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: replicas,
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "10%",
					},
					MaxSurge: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "10%",
					},
				},
			},
			Selector: &metav1.LabelSelector{
				MatchLabels: labelMap,
			},
			RevisionHistoryLimit: int32p(5),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					NodeSelector: nodeSelector,
					Containers: []corev1.Container{
						{
							Name:  inference.Spec.Name,
							Image: inference.Spec.Image,
							Ports: []corev1.ContainerPort{
								{ContainerPort: int32(port), Protocol: corev1.ProtocolTCP},
							},
							Command:         command,
							ImagePullPolicy: corev1.PullPolicy(factory.Factory.Config.ImagePullPolicy),
							Env:             envVars,
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: &allowPrivilegeEscalation,
							},
							// TODO(xieydd): Add a function to set shm size
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "dshm",
									MountPath: "/dev/shm",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "dshm",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{
									Medium: corev1.StorageMediumMemory,
								},
							},
						},
					},
				},
			},
		},
	}

	if probes != nil {
		if probes.Liveness != nil {
			deploymentSpec.Spec.Template.Spec.Containers[0].LivenessProbe = probes.Liveness
		}
		if probes.Readiness != nil {
			deploymentSpec.Spec.Template.Spec.Containers[0].ReadinessProbe = probes.Readiness
		}
		if probes.Startup != nil {
			deploymentSpec.Spec.Template.Spec.Containers[0].StartupProbe = probes.Startup
			if inference.Spec.Scaling != nil &&
				inference.Spec.Scaling.StartupDuration != nil {
				// Set the failure threshold to the number of seconds in the duration.
				deploymentSpec.Spec.Template.Spec.Containers[0].
					StartupProbe.FailureThreshold = int32(
					*inference.Spec.Scaling.StartupDuration / probes.Startup.PeriodSeconds)
			}
		}
	}

	if inference.Spec.Resources != nil {
		deploymentSpec.Spec.Template.Spec.Containers[0].Resources = *inference.Spec.Resources
		if q, ok := inference.Spec.Resources.Limits[consts.ResourceNvidiaGPU]; ok {
			if q.Value() > 0 {
				// If GPU is requested, add the GPU toleration.
				deploymentSpec.Spec.Template.Spec.Tolerations = makeTolerationGPU()
				if factory.Factory.Config.RuntimeClassNvidia {
					deploymentSpec.Spec.Template.Spec.RuntimeClassName = &runtimeClassNvidia
				}
			} else {
				// If GPU is not requested, set CUDA_VISIBLE_DEVICES to empty string.
				deploymentSpec.Spec.Template.Spec.Containers[0].Env = append(
					deploymentSpec.Spec.Template.Spec.Containers[0].Env,
					corev1.EnvVar{
						Name:  "CUDA_VISIBLE_DEVICES",
						Value: "",
					},
				)
			}
		}
	}

	factory.ConfigureReadOnlyRootFilesystem(inference, deploymentSpec)
	factory.ConfigureContainerUserID(deploymentSpec)

	return deploymentSpec
}

func makeTolerationGPU() []corev1.Toleration {
	res := []corev1.Toleration{
		{
			Key:      consts.TolerationGPU,
			Operator: corev1.TolerationOpEqual,
			Value:    "true",
		},
		{
			Key:      consts.TolerationNvidiaGPUPresent,
			Operator: corev1.TolerationOpEqual,
			Value:    "present",
		},
	}
	return res
}

func makeCommand(inference *v2alpha1.Inference) []string {
	if inference.Spec.Command != nil {
		res := strings.Split(*inference.Spec.Command, " ")
		return res
	}
	return nil
}

func makeEnvVars(inference *v2alpha1.Inference) []corev1.EnvVar {
	envVars := []corev1.EnvVar{}

	if inference.Spec.EnvVars != nil {
		for k, v := range inference.Spec.EnvVars {
			envVars = append(envVars, corev1.EnvVar{
				Name:  k,
				Value: v,
			})
		}
	}

	// Set environment variables for different frameworks.
	switch inference.Spec.Framework {
	case v2alpha1.FrameworkGradio:
		envVars = addEnvVarIfNotExists(envVars,
			"GRADIO_SERVER_NAME", "0.0.0.0")
		envVars = addEnvVarIfNotExists(envVars,
			"GRADIO_SERVER_PORT", "7860")
	case v2alpha1.FrameworkMosec:
		envVars = addEnvVarIfNotExists(envVars,
			"MOSEC_PORT", strconv.Itoa(defaultPort))
	case v2alpha1.FrameworkStreamlit:
		envVars = addEnvVarIfNotExists(envVars, "STREAMLIT_SERVER_ENABLE_CORS", "false")
		envVars = addEnvVarIfNotExists(envVars, "STREAMLIT_SERVER_ADDRESS", "0.0.0.0")
		envVars = addEnvVarIfNotExists(envVars, "STREAMLIT_SERVER_ENABLE_XSRF_PROTECTION", "false")
	}

	return envVars
}

func addEnvVarIfNotExists(envVars []corev1.EnvVar, name, value string) []corev1.EnvVar {
	for _, envVar := range envVars {
		if envVar.Name == name {
			return envVars
		}
	}

	return append(envVars, corev1.EnvVar{
		Name:  name,
		Value: value,
	})
}

func makeLabels(inference *v2alpha1.Inference) map[string]string {
	labels := map[string]string{
		consts.LabelInferenceName: inference.Spec.Name,
		"app":                     inference.Spec.Name,
		"controller":              inference.Name,
	}
	if inference.Spec.Labels != nil {
		for k, v := range inference.Spec.Labels {
			labels[k] = v
		}
	}

	return labels
}

func makePort(inference *v2alpha1.Inference) int {
	if inference.Spec.Port != nil {
		return int(*inference.Spec.Port)
	}

	return defaultPort
}

func makeAnnotations(inference *v2alpha1.Inference) map[string]string {
	annotations := make(map[string]string)

	// disable scraping since the watchdog doesn't expose a metrics endpoint
	annotations["prometheus.io.scrape"] = "false"

	// copy inference annotations
	if inference.Spec.Annotations != nil {
		for k, v := range inference.Spec.Annotations {
			annotations[k] = v
		}
	}

	// save inference spec in deployment annotations
	// used to detect changes in inference spec
	specJSON, err := json.Marshal(inference.Spec)
	if err != nil {
		glog.Errorf("Failed to marshal inference spec: %s", err.Error())
		return annotations
	}

	annotations[annotationInferenceSpec] = string(specJSON)
	return annotations
}

func makeNodeSelector(constraints []string) map[string]string {
	selector := make(map[string]string)

	if len(constraints) > 0 {
		for _, constraint := range constraints {
			parts := strings.Split(constraint, "=")

			if len(parts) == 2 {
				selector[parts[0]] = parts[1]
			}
		}
	}

	return selector
}

// deploymentNeedsUpdate determines if the inference spec is different from the deployment spec
func deploymentNeedsUpdate(
	inference *v2alpha1.Inference, deployment *appsv1.Deployment) bool {
	prevFnSpecJson := deployment.ObjectMeta.Annotations[annotationInferenceSpec]
	if prevFnSpecJson == "" {
		// is a new deployment or is an old deployment that is missing the annotation
		return true
	}

	prevFnSpec := &v2alpha1.InferenceSpec{}
	err := json.Unmarshal([]byte(prevFnSpecJson), prevFnSpec)
	if err != nil {
		glog.Errorf("Failed to parse previous inference spec: %s", err.Error())
		return true
	}
	prevFn := v2alpha1.Inference{
		Spec: *prevFnSpec,
	}

	if diff := cmp.Diff(prevFn.Spec, inference.Spec); diff != "" {
		glog.V(2).Infof("Change detected for %s diff\n%s", inference.Name, diff)
		return true
	} else {
		glog.V(3).Infof("No changes detected for %s", inference.Name)
	}

	return false
}

func int32p(i int32) *int32 {
	return &i
}
