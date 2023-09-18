package runtime

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	apicorev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/client-go/informers/apps/v1"
	corev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	clientsetscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	kubefledged "github.com/senthilrch/kube-fledged/pkg/client/clientset/versioned"
	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/pkg/config"
	"github.com/tensorchord/openmodelz/agent/pkg/event"
	ingressclient "github.com/tensorchord/openmodelz/ingress-operator/pkg/client/clientset/versioned"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	apis "github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	modelzetes "github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	clientset "github.com/tensorchord/openmodelz/modelzetes/pkg/client/clientset/versioned"
	modelzv2alpha1 "github.com/tensorchord/openmodelz/modelzetes/pkg/client/informers/externalversions/modelzetes/v2alpha1"
)

type Runtime interface {
	// build
	BuildList(ctx context.Context, namespace string) ([]types.Build, error)
	BuildCreate(ctx context.Context, req types.Build, inference *v2alpha1.Inference, builderImage,
		buildkitdAddress, buildCtlBin, secret string) error
	BuildGet(ctx context.Context, namespace, buildName string) (types.Build, error)
	// cache
	ImageCacheCreate(ctx context.Context, req types.ImageCache, inference *modelzetes.Inference) error
	// inference
	InferenceCreate(ctx context.Context,
		req types.InferenceDeployment, cfg config.IngressConfig, event string, serverPort int) error
	InferenceDelete(ctx context.Context, namespace, inferenceName, ingressNamespace, event string) error
	InferenceExec(ctx *gin.Context, namespace, instance string, commands []string, tty bool) error
	InferenceGet(namespace, inferenceName string) (*types.InferenceDeployment, error)
	InferenceGetCRD(namespace, name string) (*apis.Inference, error)
	InferenceInstanceList(namespace, inferenceName string) ([]types.InferenceDeploymentInstance, error)
	InferenceList(namespace string) ([]types.InferenceDeployment, error)
	InferenceScale(ctx context.Context, namespace string, req types.ScaleServiceRequest, inf *types.InferenceDeployment) error
	InferenceUpdate(ctx context.Context, namespace string, req types.InferenceDeployment, event string) (err error)
	// namespace
	NamespaceList(ctx context.Context) ([]string, error)
	NamespaceCreate(ctx context.Context, name string) error
	NamespaceGet(ctx context.Context, name string) bool
	NamespaceDelete(ctx context.Context, name string) error
	// server
	ServerDeleteNode(ctx context.Context, name string) error
	ServerLabelCreate(ctx context.Context, name string, spec types.ServerSpec) error
	ServerList(ctx context.Context) ([]types.Server, error)
	// managed cluster
	GetClusterInfo(cluster *types.ManagedCluster) error
}

type generalRuntime struct {
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
	r := generalRuntime{
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
