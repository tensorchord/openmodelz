package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	pkgerrors "github.com/pkg/errors"

	faasv1 "github.com/tensorchord/openmodelz/ingress-operator/pkg/apis/modelzetes/v1"
	clientset "github.com/tensorchord/openmodelz/ingress-operator/pkg/client/clientset/versioned"
	informers "github.com/tensorchord/openmodelz/ingress-operator/pkg/client/informers/externalversions"
	listers "github.com/tensorchord/openmodelz/ingress-operator/pkg/client/listers/modelzetes/v1"
	"github.com/tensorchord/openmodelz/ingress-operator/pkg/controller"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	networkingv1 "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	klog "k8s.io/klog"
)

// SyncHandler is the controller implementation for Function resources
type SyncHandler struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface

	functionsLister listers.InferenceIngressLister

	ingressLister networkingv1.IngressLister

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new OpenFaaS controller
func NewController(
	kubeclientset kubernetes.Interface,
	faasclientset clientset.Interface,
	kubeInformerFactory kubeinformers.SharedInformerFactory,
	functionIngressFactory informers.SharedInformerFactory,
) controller.BaseController {

	recorder := controller.EventRecorder(kubeclientset)
	functionIngress := functionIngressFactory.Tensorchord().V1().InferenceIngresses()
	ingressInformer := kubeInformerFactory.Networking().V1().Ingresses()
	ingressLister := ingressInformer.Lister()

	syncer := SyncHandler{
		kubeclientset:   kubeclientset,
		functionsLister: functionIngress.Lister(),
		ingressLister:   ingressLister,
		recorder:        recorder,
	}

	ctrl := controller.BaseController{
		FunctionsLister: functionIngress.Lister(),
		FunctionsSynced: functionIngress.Informer().HasSynced,
		Workqueue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "FunctionIngresses"),
		SyncHandler:     syncer.handler,
	}
	klog.Info("Setting up event handlers")
	ctrl.SetupEventHandlers(functionIngress, kubeInformerFactory)
	ingressInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: ctrl.HandleObject,
	})

	return ctrl
}

// handler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the fni resource
// with the current status of the resource.
func (h SyncHandler) handler(ctx context.Context, key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		runtime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the fni resource with this namespace/name
	fni, err := h.functionsLister.InferenceIngresses(namespace).Get(name)
	if err != nil {
		// The fni resource may no longer exist, in which case we stop processing.
		if errors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("function ingress '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	fniName := fni.ObjectMeta.Name
	klog.Infof("FunctionIngress name: %v", fniName)

	ingresses := h.ingressLister.Ingresses(namespace)
	ingress, getIngressErr := ingresses.Get(fni.Name)
	createIngress := errors.IsNotFound(getIngressErr)
	if !createIngress && ingress == nil {
		klog.Errorf("cannot get ingress: %s in %s, error: %s", fni.Name, namespace, getIngressErr.Error())
	}

	klog.Info("fni.Spec.UseTLS() ", fni.Spec.UseTLS())
	klog.Info("createIngress ", createIngress)

	if createIngress {
		rules := makeRules(fni)
		tls := makeTLS(fni)

		ns := namespace
		if mns, exists := os.LookupEnv("MODELZ_NAMESPACE"); exists {
			ns = mns
		}

		newIngress := netv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{
				Name:            name,
				Namespace:       ns,
				Annotations:     controller.MakeAnnotations(fni),
				OwnerReferences: controller.MakeOwnerRef(fni),
			},
			Spec: netv1.IngressSpec{
				Rules: rules,
				TLS:   tls,
			},
		}

		_, createErr := h.kubeclientset.NetworkingV1().Ingresses(ns).Create(ctx, &newIngress, metav1.CreateOptions{})
		if createErr != nil {
			klog.Errorf("cannot create ingress: %v in %v, error: %v", name, namespace, createErr.Error())
		}

		h.recorder.Event(fni, corev1.EventTypeNormal, controller.SuccessSynced, controller.MessageResourceSynced)
		return nil
	}

	old := faasv1.InferenceIngress{}

	if val, ok := ingress.Annotations["com.openfaas.spec"]; ok && len(val) > 0 {
		unmarshalErr := json.Unmarshal([]byte(val), &old)
		if unmarshalErr != nil {
			return pkgerrors.Wrap(unmarshalErr, "unable to unmarshal from field com.openfaas.spec")
		}
	}

	// Update the Deployment resource if the fni definition differs
	if controller.IngressNeedsUpdate(&old, fni) {
		klog.Infof("Updating FunctionIngress: %s", fniName)

		if old.ObjectMeta.Name != fni.ObjectMeta.Name {
			return fmt.Errorf("cannot rename object")
		}

		updated := ingress.DeepCopy()

		rules := makeRules(fni)

		annotations := controller.MakeAnnotations(fni)
		for k, v := range annotations {
			updated.Annotations[k] = v
		}

		updated.Spec.Rules = rules
		updated.Spec.TLS = makeTLS(fni)

		_, updateErr := h.kubeclientset.NetworkingV1().Ingresses(namespace).Update(ctx, updated, metav1.UpdateOptions{})
		if updateErr != nil {
			klog.Errorf("error updating ingress: %v", updateErr)
			return updateErr
		}
	}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		return fmt.Errorf("transient error: %v", err)
	}

	h.recorder.Event(fni, corev1.EventTypeNormal, controller.SuccessSynced, controller.MessageResourceSynced)
	return nil
}

func makeRules(fni *faasv1.InferenceIngress) []netv1.IngressRule {
	path := "/(.*)"

	if fni.Spec.BypassGateway {
		path = "/"
	}

	if len(fni.Spec.Path) > 0 {
		path = fni.Spec.Path
	}

	if controller.GetClass(fni.Spec.IngressType) == "traefik" {
		// We have to trim the regex and the trailing slash for Traefik,
		// otherwise routing won't work
		path = strings.TrimRight(path, "/(.*)")
		if len(path) == 0 {
			path = "/"
		}
	}

	serviceHost := "apiserver"
	if fni.Spec.BypassGateway {
		serviceHost = fni.Spec.Function
	}

	pathType := netv1.PathTypeImplementationSpecific

	return []netv1.IngressRule{
		{
			Host: fni.Spec.Domain,
			IngressRuleValue: netv1.IngressRuleValue{
				HTTP: &netv1.HTTPIngressRuleValue{
					Paths: []netv1.HTTPIngressPath{
						{
							Path:     path,
							PathType: &pathType,
							Backend: netv1.IngressBackend{
								Service: &netv1.IngressServiceBackend{
									Name: serviceHost,
									Port: netv1.ServiceBackendPort{
										Number: controller.OpenfaasWorkloadPort,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func makeTLS(fni *faasv1.InferenceIngress) []netv1.IngressTLS {
	if !fni.Spec.UseTLS() {
		return []netv1.IngressTLS{}
	}

	return []netv1.IngressTLS{
		{
			// Use default secret name, thus no need to specify SecretName.
			Hosts: []string{
				fni.Spec.Domain,
			},
		},
	}
}
