package runtime

import (
	"github.com/sirupsen/logrus"
	ingressclient "github.com/tensorchord/openmodelz/ingress-operator/pkg/client/clientset/versioned"
	clientset "github.com/tensorchord/openmodelz/modelzetes/pkg/client/clientset/versioned"
	modelzv2alpha1 "github.com/tensorchord/openmodelz/modelzetes/pkg/client/informers/externalversions/modelzetes/v2alpha1"
	appsv1 "k8s.io/client-go/informers/apps/v1"
	corev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/tensorchord/openmodelz/agent/pkg/event"
)

type Runtime struct {
	endpointsInformer  corev1.EndpointsInformer
	deploymentInformer appsv1.DeploymentInformer
	inferenceInformer  modelzv2alpha1.InferenceInformer
	podInformer        corev1.PodInformer

	kubeClient      kubernetes.Interface
	ingressClient   ingressclient.Interface
	inferenceClient clientset.Interface

	logger        *logrus.Entry
	eventRecorder event.Interface

	ingressEnabled bool
	eventEnabled   bool
}

func New(
	endpointsInformer corev1.EndpointsInformer,
	deploymentInformer appsv1.DeploymentInformer,
	inferenceInformer modelzv2alpha1.InferenceInformer,
	podInformer corev1.PodInformer,
	kubeClient kubernetes.Interface,
	ingressClient ingressclient.Interface,
	inferenceClient clientset.Interface,
	eventRecorder event.Interface,
	ingressEnabled, eventEnabled bool,
) (Runtime, error) {
	return Runtime{
		endpointsInformer:  endpointsInformer,
		deploymentInformer: deploymentInformer,
		inferenceInformer:  inferenceInformer,
		podInformer:        podInformer,
		kubeClient:         kubeClient,
		ingressClient:      ingressClient,
		inferenceClient:    inferenceClient,
		logger:             logrus.WithField("component", "runtime"),
		eventRecorder:      eventRecorder,
		ingressEnabled:     ingressEnabled,
		eventEnabled:       eventEnabled,
	}, nil
}
