package runtime

import (
	"github.com/sirupsen/logrus"
	apicorev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/client-go/informers/apps/v1"
	corev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	kubefledged "github.com/senthilrch/kube-fledged/pkg/client/clientset/versioned"
	"github.com/tensorchord/openmodelz/agent/pkg/event"
	ingressclient "github.com/tensorchord/openmodelz/ingress-operator/pkg/client/clientset/versioned"
	clientset "github.com/tensorchord/openmodelz/modelzetes/pkg/client/clientset/versioned"
	modelzv2alpha1 "github.com/tensorchord/openmodelz/modelzetes/pkg/client/informers/externalversions/modelzetes/v2alpha1"
)

type Runtime struct {
	endpointsInformer  corev1.EndpointsInformer
	deploymentInformer appsv1.DeploymentInformer
	inferenceInformer  modelzv2alpha1.InferenceInformer
	podInformer        corev1.PodInformer

	kubeClient        kubernetes.Interface
	clientConfig      *rest.Config
	restClient        *rest.RESTClient
	ingressClient     ingressclient.Interface
	inferenceClient   clientset.Interface
	kubefledgedClient kubefledged.Interface

	logger        *logrus.Entry
	eventRecorder event.Interface

	ingressEnabled       bool
	ingressAnyIPToDomain bool
	eventEnabled         bool
	buildEnabled         bool
}

func New(clientConfig *rest.Config,
	endpointsInformer corev1.EndpointsInformer,
	deploymentInformer appsv1.DeploymentInformer,
	inferenceInformer modelzv2alpha1.InferenceInformer,
	podInformer corev1.PodInformer,
	kubeClient kubernetes.Interface,
	ingressClient ingressclient.Interface,
	kubefledgedClient kubefledged.Interface,
	inferenceClient clientset.Interface,
	eventRecorder event.Interface,
	ingressEnabled bool,
	eventEnabled bool,
	buildEnabled bool,
	ingressAnyIPToDomain bool,
) (Runtime, error) {
	r := Runtime{
		endpointsInformer:    endpointsInformer,
		deploymentInformer:   deploymentInformer,
		inferenceInformer:    inferenceInformer,
		podInformer:          podInformer,
		kubeClient:           kubeClient,
		kubefledgedClient:    kubefledgedClient,
		clientConfig:         clientConfig,
		ingressClient:        ingressClient,
		inferenceClient:      inferenceClient,
		logger:               logrus.WithField("component", "runtime"),
		eventRecorder:        eventRecorder,
		ingressEnabled:       ingressEnabled,
		ingressAnyIPToDomain: ingressAnyIPToDomain,
		eventEnabled:         eventEnabled,
		buildEnabled:         buildEnabled,
	}
	// Ref https://github.com/operator-framework/operator-sdk/issues/1570
	clientConfig.APIPath = "api"
	clientConfig.GroupVersion = &apicorev1.SchemeGroupVersion
	clientConfig.NegotiatedSerializer = clientsetscheme.Codecs
	r.clientConfig = clientConfig
	restClient, err := rest.RESTClientFor(clientConfig)
	if err != nil {
		return r, err
	}
	r.restClient = restClient
	return r, nil
}
