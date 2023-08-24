package server

import (
	"fmt"

	"github.com/sirupsen/logrus"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	kubefledged "github.com/senthilrch/kube-fledged/pkg/client/clientset/versioned"
	"github.com/tensorchord/openmodelz/agent/pkg/k8s"
	"github.com/tensorchord/openmodelz/agent/pkg/log"
	"github.com/tensorchord/openmodelz/agent/pkg/runtime"
	"github.com/tensorchord/openmodelz/agent/pkg/scaling"
	ingressclient "github.com/tensorchord/openmodelz/ingress-operator/pkg/client/clientset/versioned"
	clientset "github.com/tensorchord/openmodelz/modelzetes/pkg/client/clientset/versioned"
	informers "github.com/tensorchord/openmodelz/modelzetes/pkg/client/informers/externalversions"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/signals"
)

func (s *Server) initKubernetesResources() error {
	clientCmdConfig, err := clientcmd.BuildConfigFromFlags(
		s.config.KubeConfig.MasterURL, s.config.KubeConfig.Kubeconfig)
	if err != nil {
		return err
	}

	clientCmdConfig.QPS = float32(s.config.KubeConfig.QPS)
	clientCmdConfig.Burst = s.config.KubeConfig.Burst

	kubeClient, err := kubernetes.NewForConfig(clientCmdConfig)
	if err != nil {
		return err
	}
	inferenceClient, err := clientset.NewForConfig(clientCmdConfig)
	if err != nil {
		return err
	}

	var ingressClient ingressclient.Interface
	if s.config.Ingress.IngressEnabled {
		ingressClient, err = ingressclient.NewForConfig(clientCmdConfig)
		if err != nil {
			return err
		}
	}
	kubefledgedClient, err := kubefledged.NewForConfig(clientCmdConfig)
	if err != nil {
		return err
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(
		kubeClient, s.config.KubeConfig.ResyncPeriod)

	inferenceInformerFactory := informers.NewSharedInformerFactoryWithOptions(
		inferenceClient, s.config.KubeConfig.ResyncPeriod)

	// set up signals so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()

	inferences := inferenceInformerFactory.Tensorchord().V2alpha1().Inferences()
	go inferences.Informer().Run(stopCh)
	if ok := cache.WaitForNamedCacheSync(
		fmt.Sprintf("%s:inferences", consts.ProviderName),
		stopCh, inferences.Informer().HasSynced); !ok {
		s.logger.Errorf("failed to wait for cache to sync")
	}

	deployments := kubeInformerFactory.Apps().V1().Deployments()
	go deployments.Informer().Run(stopCh)
	if ok := cache.WaitForNamedCacheSync(
		fmt.Sprintf("%s:deployments", consts.ProviderName),
		stopCh, deployments.Informer().HasSynced); !ok {
		s.logger.Errorf("failed to wait for cache to sync")
	}

	pods := kubeInformerFactory.Core().V1().Pods()
	go pods.Informer().Run(stopCh)
	if ok := cache.WaitForNamedCacheSync(
		fmt.Sprintf("%s:pods", consts.ProviderName),
		stopCh, pods.Informer().HasSynced); !ok {
		s.logger.Errorf("failed to wait for cache to sync")
	}

	endpoints := kubeInformerFactory.Core().V1().Endpoints()
	go endpoints.Informer().Run(stopCh)
	if ok := cache.WaitForNamedCacheSync(
		fmt.Sprintf("%s:endpoints", consts.ProviderName),
		stopCh, endpoints.Informer().HasSynced); !ok {
		s.logger.Errorf("failed to wait for cache to sync")
	}

	runtime, err := runtime.New(clientCmdConfig,
		endpoints, deployments, inferences, pods,
		kubeClient, ingressClient, kubefledgedClient, inferenceClient,
		s.eventRecorder,
		s.config.Ingress.IngressEnabled, s.config.DB.EventEnabled,
		s.config.Build.BuildEnabled, s.config.Ingress.AnyIPToDomain,
	)
	if err != nil {
		return err
	}
	s.runtime = runtime
	if s.config.Server.Dev {
		logrus.Warn("running in dev mode, using port forwarding to access pods, please do not use dev mode in production")
		s.endpointResolver = k8s.NewPortForwardingResolver(clientCmdConfig, kubeClient)
	} else {
		s.endpointResolver = k8s.NewEndpointResolver(endpoints.Lister())
	}
	s.deploymentLogRequester = log.NewK8sAPIRequestor(kubeClient)
	s.scaler, err = scaling.NewInferenceScaler(runtime, s.config.Inference.CacheTTL)
	if err != nil {
		return err
	}
	if s.scaler == nil {
		return fmt.Errorf("scaler is nil")
	}
	return nil
}
