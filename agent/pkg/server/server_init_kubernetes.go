package server

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	kubeinformersv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	kubefledged "github.com/senthilrch/kube-fledged/pkg/client/clientset/versioned"
	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/pkg/event"
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
	s.podStartWatch(pods, kubeClient)
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
		s.config.Ingress.IngressEnabled, s.config.ModelZCloud.EventEnabled,
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

// podStartWatch log event when pod start began and finished
func (s *Server) podStartWatch(pods kubeinformersv1.PodInformer, client *kubernetes.Clientset) {
	pods.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			new := obj.(*v1.Pod)
			controlPlane, exist := new.Annotations[consts.AnnotationControlPlaneKey]
			// for inference created by modelz apiserver
			if !exist || controlPlane != consts.ModelzAnnotationValue {
				return
			}
			podWatchEventLog(s.eventRecorder, new, types.PodCreateEvent)
			start := time.Now()

			// Ticker will keep watching until pod start or timeout
			ticker := time.NewTicker(time.Second * 2)
			timeout := time.After(5 * time.Minute)
			go func() {
				for {
					select {
					case <-timeout:
						podWatchEventLog(s.eventRecorder, new, types.PodTimeoutEvent)
						return
					case <-ticker.C:
						pod, err := client.CoreV1().Pods(new.Namespace).Get(context.TODO(), new.Name, metav1.GetOptions{})
						if err != nil {
							logrus.WithFields(logrus.Fields{
								"namespace":  pod.Namespace,
								"deployment": pod.Labels["app"],
								"name":       pod.Name,
							}).Errorf("failed to get pod: %s", err)
							return
						}
						for _, c := range pod.Status.Conditions {
							if c.Type == v1.PodReady && c.Status == v1.ConditionTrue {
								podWatchEventLog(s.eventRecorder, new, types.PodReadyEvent)
								label := prometheus.Labels{
									"inference_name": new.Labels["app"],
									"source_image":   new.Annotations[consts.AnnotationDockerImage]}
								s.metricsOptions.PodStartHistogram.With(label).
									Observe(time.Since(start).Seconds())
								return
							}
						}
					}
				}
			}()
		},
	})
}

// log status for pod watch status transfer
func podWatchEventLog(recorder event.Interface, obj *v1.Pod, event string) {
	deployment := obj.Labels["app"]
	err := recorder.CreateDeploymentEvent(obj.Namespace, deployment, event, obj.Name)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"namespace":  obj.Namespace,
			"deployment": deployment,
			"name":       obj.Name,
			"event":      event,
		}).Errorf("failed to create deployment event: %s", err)
	}
}
