package v1

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	clientset "github.com/tensorchord/openmodelz/ingress-operator/pkg/client/clientset/versioned"
	informers "github.com/tensorchord/openmodelz/ingress-operator/pkg/client/informers/externalversions"
	"github.com/tensorchord/openmodelz/ingress-operator/pkg/config"
	"github.com/tensorchord/openmodelz/ingress-operator/pkg/controller"
)

func New(c config.Config, stopCh <-chan struct{}) (*controller.BaseController, error) {
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

	ingressClient, err := clientset.NewForConfig(clientCmdConfig)
	if err != nil {
		return nil, fmt.Errorf("error building Inference clientset: %s", err.Error())
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactoryWithOptions(kubeClient, c.KubeConfig.ResyncPeriod)

	ingressInformerFactory := informers.NewSharedInformerFactoryWithOptions(ingressClient, c.KubeConfig.ResyncPeriod)

	capabilities, err := getPreferredAvailableAPIs(kubeClient, "Ingress")
	if err != nil {
		return nil, fmt.Errorf("error retrieving Kubernetes cluster capabilities: %s", err.Error())
	}
	logrus.Infof("cluster supports ingress in: %s", capabilities)

	if !capabilities.Has("networking.k8s.io/v1") {
		return nil, errors.New("networking.k8s.io/v1 is not available")
	}

	inferenceIngresses := ingressInformerFactory.Tensorchord().V1().InferenceIngresses()
	go inferenceIngresses.Informer().Run(stopCh)
	if ok := cache.WaitForNamedCacheSync(
		fmt.Sprintf("%s:inferenceingresses", "tensorchord"),
		stopCh, inferenceIngresses.Informer().HasSynced); !ok {
		return nil, errors.New("failed to wait for inferenceingresses caches to sync")
	}
	ingresses := kubeInformerFactory.Networking().V1().Ingresses()
	go ingresses.Informer().Run(stopCh)
	if ok := cache.WaitForNamedCacheSync(
		fmt.Sprintf("%s:ingresses", "networking"),
		stopCh, ingresses.Informer().HasSynced); !ok {
		return nil, errors.New("failed to wait for ingresses caches to sync")
	}

	ctr := NewController(c,
		kubeClient, ingressClient, kubeInformerFactory,
		ingressInformerFactory)
	return &ctr, nil
}

// getPreferredAvailableAPIs queries the cluster for the preferred resources information and returns a Capabilities
// instance containing those api groups that support the specified kind.
//
// kind should be the title case singular name of the kind. For example, "Ingress" is the kind for a resource "ingress".
func getPreferredAvailableAPIs(client kubernetes.Interface, kind string) (Capabilities, error) {
	discoveryclient := client.Discovery()
	lists, err := discoveryclient.ServerPreferredResources()
	if err != nil {
		return nil, err
	}

	caps := Capabilities{}
	for _, list := range lists {
		if len(list.APIResources) == 0 {
			continue
		}
		for _, resource := range list.APIResources {
			if len(resource.Verbs) == 0 {
				continue
			}
			if resource.Kind == kind {
				caps[list.GroupVersion] = true
			}
		}
	}

	return caps, nil
}

type Capabilities map[string]bool

func (c Capabilities) Has(wanted string) bool {
	return c[wanted]
}

func (c Capabilities) String() string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}
