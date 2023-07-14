package controller

import (
	"errors"
	"fmt"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/tensorchord/openmodelz/modelzetes/pkg/client/clientset/versioned"
	informers "github.com/tensorchord/openmodelz/modelzetes/pkg/client/informers/externalversions"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/config"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/k8s"
)

func New(c config.Config, stopCh <-chan struct{}) (*Controller, error) {
	clientCmdConfig, err := clientcmd.BuildConfigFromFlags(
		c.KubeConfig.MasterURL, c.KubeConfig.Kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubeconfig: %s", err.Error())
	}

	clientCmdConfig.QPS = float32(c.KubeConfig.QPS)
	clientCmdConfig.Burst = c.KubeConfig.Burst

	kubeClient, err := kubernetes.NewForConfig(clientCmdConfig)
	if err != nil {
		return nil, fmt.Errorf("error building Kubernetes clientset: %s", err.Error())
	}

	inferenceClient, err := clientset.NewForConfig(clientCmdConfig)
	if err != nil {
		return nil, fmt.Errorf("error building Inference clientset: %s", err.Error())
	}

	deployConfig := k8s.DeploymentConfig{
		HTTPProbe:      true,
		SetNonRootUser: false,
		ReadinessProbe: &k8s.ProbeConfig{
			InitialDelaySeconds: int32(c.Probes.Readiness.InitialDelaySeconds),
			TimeoutSeconds:      int32(c.Probes.Readiness.TimeoutSeconds),
			PeriodSeconds:       int32(c.Probes.Readiness.PeriodSeconds),
		},
		LivenessProbe: &k8s.ProbeConfig{
			InitialDelaySeconds: int32(c.Probes.Liveness.InitialDelaySeconds),
			TimeoutSeconds:      int32(c.Probes.Liveness.TimeoutSeconds),
			PeriodSeconds:       int32(c.Probes.Liveness.PeriodSeconds),
		},
		StartupProbe: &k8s.ProbeConfig{
			InitialDelaySeconds: int32(c.Probes.Startup.InitialDelaySeconds),
			TimeoutSeconds:      int32(c.Probes.Startup.TimeoutSeconds),
			PeriodSeconds:       int32(c.Probes.Startup.PeriodSeconds),
		},
		ImagePullPolicy:   c.Inference.ImagePullPolicy,
		ProfilesNamespace: "default",
	}

	if c.HuggingfaceProxy.Endpoint == "" {
		deployConfig.HuggingfacePullThroughCache = false
	} else {
		deployConfig.HuggingfacePullThroughCache = true
		deployConfig.HuggingfacePullThroughCacheEndpoint = c.HuggingfaceProxy.Endpoint
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(kubeClient, c.KubeConfig.ResyncPeriod)

	inferenceInformerFactory := informers.NewSharedInformerFactoryWithOptions(inferenceClient, c.KubeConfig.ResyncPeriod)

	inferences := inferenceInformerFactory.Tensorchord().V2alpha1().Inferences()
	go inferences.Informer().Run(stopCh)
	if ok := cache.WaitForNamedCacheSync(
		fmt.Sprintf("%s:inferences", consts.ProviderName),
		stopCh, inferences.Informer().HasSynced); !ok {
		return nil, errors.New("failed to wait for inference caches to sync")
	}

	deployments := kubeInformerFactory.Apps().V1().Deployments()
	go deployments.Informer().Run(stopCh)
	if ok := cache.WaitForNamedCacheSync(
		fmt.Sprintf("%s:deployments", consts.ProviderName),
		stopCh, deployments.Informer().HasSynced); !ok {
		return nil, errors.New("failed to wait for deployment caches to sync")
	}

	controllerFactory := NewFunctionFactory(kubeClient, deployConfig)

	ctr := NewController(
		kubeClient, inferenceClient, kubeInformerFactory,
		inferenceInformerFactory, controllerFactory)
	return ctr, nil
}
