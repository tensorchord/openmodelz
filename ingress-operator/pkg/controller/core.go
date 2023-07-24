package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-cmp/cmp"
	faasv1 "github.com/tensorchord/openmodelz/ingress-operator/pkg/apis/modelzetes/v1"
	"github.com/tensorchord/openmodelz/ingress-operator/pkg/client/clientset/versioned/scheme"
	faasscheme "github.com/tensorchord/openmodelz/ingress-operator/pkg/client/clientset/versioned/scheme"
	v1 "github.com/tensorchord/openmodelz/ingress-operator/pkg/client/informers/externalversions/modelzetes/v1"
	listers "github.com/tensorchord/openmodelz/ingress-operator/pkg/client/listers/modelzetes/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	klog "k8s.io/klog"
)

const AgentName = "ingress-operator"
const FaasIngressKind = "InferenceIngress"
const OpenfaasWorkloadPort = 8080

const (
	// SuccessSynced is used as part of the Event 'reason' when a Function is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Function fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"
	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by controller"
	// MessageResourceSynced is the message used for an Event fired when a Function
	// is synced successfully
	MessageResourceSynced = "FunctionIngress synced successfully"
)

// BaseController is the controller contains the common function ingress
// implementation that is shared between the various versions of k8s.
type BaseController struct {
	FunctionsLister listers.InferenceIngressLister
	FunctionsSynced cache.InformerSynced

	// Workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	Workqueue workqueue.RateLimitingInterface

	SyncHandler func(ctx context.Context, key string) error
}

func (c BaseController) Run(threadiness int, stopCh <-chan struct{}) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer runtime.HandleCrash()
	defer c.Workqueue.ShutDown()
	defer cancel()

	// Start the informer factories to begin populating the informer caches
	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.FunctionsSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Function resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker(ctx), time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the workqueue.
func (c BaseController) runWorker(ctx context.Context) func() {
	return func() {
		for c.processNextWorkItem(ctx) {
		}
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c BaseController) processNextWorkItem(ctx context.Context) bool {
	obj, shutdown := c.Workqueue.Get()

	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer c.Workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.Workqueue.Forget(obj)
			runtime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		if err := c.SyncHandler(ctx, key); err != nil {
			return fmt.Errorf("error syncing '%s': %s", key, err.Error())
		}
		c.Workqueue.Forget(obj)
		return nil
	}(obj)

	if err != nil {
		runtime.HandleError(err)
		return true
	}

	return true
}

// enqueueFunction takes a fni resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than fni.
func (c *BaseController) EnqueueFunction(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		runtime.HandleError(err)
		return
	}
	c.Workqueue.AddRateLimited(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the fni resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that fni resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c BaseController) HandleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			runtime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.V(4).Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}

	klog.V(4).Infof("Processing object: %s", object.GetName())
	if ownerRef := metav1.GetControllerOf(object); ownerRef != nil {
		// If this object is not owned by a fni, we should not do anything more
		// with it.
		if ownerRef.Kind != FaasIngressKind {
			return
		}

		fni, err := c.FunctionsLister.InferenceIngresses(object.GetNamespace()).Get(ownerRef.Name)
		if err != nil {
			klog.Infof("FunctionIngress '%s' deleted. Ignoring orphaned object '%s': %v", ownerRef.Name, object.GetSelfLink(), err)
			return
		}

		c.EnqueueFunction(fni)
		return
	}
}

func (c BaseController) SetupEventHandlers(
	functionIngress v1.InferenceIngressInformer,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
) {
	functionIngress.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: c.EnqueueFunction,
		UpdateFunc: func(old, new interface{}) {

			oldFn, ok := CheckCustomResourceType(old)
			if !ok {
				return
			}
			newFn, ok := CheckCustomResourceType(new)
			if !ok {
				return
			}
			diffSpec := cmp.Diff(oldFn.Spec, newFn.Spec)
			diffAnnotations := cmp.Diff(oldFn.ObjectMeta.Annotations, newFn.ObjectMeta.Annotations)

			if diffSpec != "" || diffAnnotations != "" {
				c.EnqueueFunction(new)
			}
		},
	})

	// Set up an event handler for when functions related resources like pods, deployments, replica sets
	// can't be materialized. This logs abnormal events like ImagePullBackOff, back-off restarting failed container,
	// failed to start container, oci runtime errors, etc
	// Enable this with -v=3
	kubeInformerFactory.Core().V1().Events().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				event := obj.(*corev1.Event)
				since := time.Since(event.LastTimestamp.Time)
				// log abnormal events occurred in the last minute
				if since.Seconds() < 61 && strings.Contains(event.Type, "Warning") {
					klog.V(3).Infof("Abnormal event detected on %s %s: %s", event.LastTimestamp, key, event.Message)
				}
			}
		},
	})
}

func GetClass(ingressType string) string {
	switch ingressType {
	case "":
	case "nginx":
		return "nginx"
	default:
		return ingressType
	}

	return "nginx"
}

func GetIssuerKind(issuerType string) string {
	switch issuerType {
	case "ClusterIssuer":
		return "cert-manager.io/cluster-issuer"
	default:
		return "cert-manager.io/issuer"
	}
}

func MakeAnnotations(fni *faasv1.InferenceIngress, host string) map[string]string {
	class := GetClass(fni.Spec.IngressType)
	specJSON, _ := json.Marshal(fni)
	annotations := make(map[string]string)

	annotations["ai.tensorchord.spec"] = string(specJSON)

	if !fni.Spec.BypassGateway {
		switch class {
		case "nginx":
			switch host {
			// TODO: make this configurable
			case "apiserver":
				annotations["nginx.ingress.kubernetes.io/rewrite-target"] = "/api/v1/" + fni.Spec.Framework +
					"/" + fni.Spec.Function + "/$1"
				annotations["nginx.ingress.kubernetes.io/use-regex"] = "true"
			default:
				annotations["nginx.ingress.kubernetes.io/rewrite-target"] = "/inference/" + fni.Name + ".default" + "/$1"
				annotations["nginx.ingress.kubernetes.io/ssl-redirect"] = "false"
				annotations["nginx.ingress.kubernetes.io/use-regex"] = "true"
			}

		}
	}

	annotations["nginx.ingress.kubernetes.io/proxy-send-timeout"] = "300"
	annotations["nginx.ingress.kubernetes.io/proxy-read-timeout"] = "300"

	// We use the default certificate for now.
	// if fni.Spec.UseTLS() {
	// 	issuerType := GetIssuerKind(fni.Spec.TLS.IssuerRef.Kind)
	// 	annotations[issuerType] = fni.Spec.TLS.IssuerRef.Name
	// }

	// Set annotations with overrides from FunctionIngress
	// annotations
	for k, v := range fni.ObjectMeta.Annotations {
		annotations[k] = v
	}

	return annotations
}

func MakeOwnerRef(fni *faasv1.InferenceIngress) []metav1.OwnerReference {
	ref := []metav1.OwnerReference{
		*metav1.NewControllerRef(fni, schema.GroupVersionKind{
			Group:   faasv1.SchemeGroupVersion.Group,
			Version: faasv1.SchemeGroupVersion.Version,
			Kind:    FaasIngressKind,
		}),
	}
	return ref
}

func CheckCustomResourceType(obj interface{}) (faasv1.InferenceIngress, bool) {
	var fn *faasv1.InferenceIngress
	var ok bool
	if fn, ok = obj.(*faasv1.InferenceIngress); !ok {
		klog.Errorf("Event Watch received an invalid object: %#v", obj)
		return faasv1.InferenceIngress{}, false
	}
	return *fn, true
}

func IngressNeedsUpdate(old, fni *faasv1.InferenceIngress) bool {
	return !cmp.Equal(old.Spec, fni.Spec) ||
		!cmp.Equal(old.ObjectMeta.Annotations, fni.ObjectMeta.Annotations)
}

func EventRecorder(client kubernetes.Interface) record.EventRecorder {
	// Create event broadcaster
	// Add o6s types to the default Kubernetes Scheme so Events can be
	// logged for faas-controller types.
	faasscheme.AddToScheme(scheme.Scheme)
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.V(4).Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: client.CoreV1().Events("")})
	return eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: AgentName})
}
