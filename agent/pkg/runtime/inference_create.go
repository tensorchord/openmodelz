package runtime

import (
	"context"
	"fmt"

	ingressv1 "github.com/tensorchord/openmodelz/ingress-operator/pkg/apis/modelzetes/v1"
	v2alpha1 "github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/errdefs"
	localconsts "github.com/tensorchord/openmodelz/agent/pkg/consts"
)

func (r Runtime) InferenceCreate(ctx context.Context,
	req types.InferenceDeployment, ingressDomain, ingressNamespace, event string) error {

	namespace := req.Spec.Namespace

	if r.eventEnabled {
		err := r.eventRecorder.CreateDeploymentEvent(namespace, req.Spec.Name, event, "")
		if err != nil {
			return err
		}
	}

	inf, err := makeInference(req, ingressDomain)
	if err != nil {
		return err
	}

	// Create the ingress
	// TODO(gaocegege): Check if the domain is already used.
	// TODO: Move it to apiserver.
	if r.ingressEnabled {
		name := req.Spec.Labels[localconsts.LabelName]

		if r.ingressAnyIPToDomain {
			// Get the service with type=loadbalancer.
			svcs, err := r.kubeClient.CoreV1().Services("").List(ctx, metav1.ListOptions{})
			if err != nil {
				return errdefs.System(fmt.Errorf("failed to list services: %v", err))
			}

			if len(svcs.Items) == 0 {
				return errdefs.System(fmt.Errorf("no service with type=LoadBalancer"))
			}
			var externalIP string
			for _, s := range svcs.Items {
				if s.Spec.Type == v1.ServiceTypeLoadBalancer {
					if len(s.Status.LoadBalancer.Ingress) == 0 {
						continue
					}
					externalIP = s.Status.LoadBalancer.Ingress[0].IP
					break
				}
			}
			// Set the domain to
			ingressDomain = fmt.Sprintf("%s.%s", externalIP, localconsts.Domain)
		}

		domain, err := makeDomain(name, ingressDomain)
		if err != nil {
			return errdefs.InvalidParameter(err)
		}

		// Set the domain.
		// Create the inference with the ingress domain.
		if inf.Spec.Annotations == nil {
			inf.Spec.Annotations = make(map[string]string)
		}
		inf.Spec.Annotations[AnnotationDomain] = fmt.Sprintf("https://%s", domain)

		_, err = r.inferenceClient.TensorchordV2alpha1().
			Inferences(namespace).Create(
			ctx, inf, metav1.CreateOptions{})
		if err != nil {
			if k8serrors.IsAlreadyExists(err) {
				return errdefs.Conflict(err)
			} else {
				return errdefs.System(err)
			}
		}

		ingress, err := makeIngress(req, domain, ingressNamespace)
		if err != nil {
			return err
		}

		_, err = r.ingressClient.TensorchordV1().
			InferenceIngresses(ingressNamespace).
			Create(ctx, ingress, metav1.CreateOptions{})
		if err != nil {
			if k8serrors.IsAlreadyExists(err) {
				return errdefs.Conflict(err)
			} else {
				return errdefs.System(err)
			}
		}
	} else {
		_, err = r.inferenceClient.TensorchordV2alpha1().
			Inferences(namespace).Create(
			ctx, inf, metav1.CreateOptions{})
		if err != nil {
			if k8serrors.IsAlreadyExists(err) {
				return errdefs.Conflict(err)
			} else {
				return errdefs.System(err)
			}
		}
	}
	return nil
}

func makeInference(request types.InferenceDeployment,
	domain string) (*v2alpha1.Inference, error) {
	is := &v2alpha1.Inference{
		ObjectMeta: metav1.ObjectMeta{
			Name:      request.Spec.Name,
			Namespace: request.Spec.Namespace,
			Labels: map[string]string{
				consts.LabelInferenceName: request.Spec.Name,
			},
		},
		Spec: v2alpha1.InferenceSpec{
			Name:          request.Spec.Name,
			Image:         request.Spec.Image,
			Framework:     v2alpha1.Framework(request.Spec.Framework),
			Port:          request.Spec.Port,
			Command:       request.Spec.Command,
			EnvVars:       request.Spec.EnvVars,
			Secrets:       request.Spec.Secrets,
			Constraints:   request.Spec.Constraints,
			Labels:        request.Spec.Labels,
			Annotations:   request.Spec.Annotations,
			HTTPProbePath: request.Spec.HTTPProbePath,
		},
	}

	if request.Spec.Scaling != nil {
		is.Spec.Scaling = &v2alpha1.ScalingConfig{
			MinReplicas:     request.Spec.Scaling.MinReplicas,
			MaxReplicas:     request.Spec.Scaling.MaxReplicas,
			TargetLoad:      request.Spec.Scaling.TargetLoad,
			ZeroDuration:    request.Spec.Scaling.ZeroDuration,
			StartupDuration: request.Spec.Scaling.StartupDuration,
		}
		if request.Spec.Scaling.Type != nil {
			buf := v2alpha1.ScalingType(*request.Spec.Scaling.Type)
			is.Spec.Scaling.Type = &buf
		}
	}

	rr, err := createResources(request)
	if err != nil {
		return nil, errdefs.InvalidParameter(err)
	}

	is.Spec.Resources = &rr
	return is, nil
}

func makeIngress(request types.InferenceDeployment, domain, BaseNamespace string) (*ingressv1.InferenceIngress, error) {
	labels := map[string]string{
		consts.LabelInferenceName: request.Spec.Name,
	}

	if request.Spec.Labels == nil {
		return nil, errdefs.InvalidParameter(fmt.Errorf("labels is required"))
	}

	ingress := &ingressv1.InferenceIngress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      request.Spec.Name,
			Namespace: BaseNamespace,
			Labels:    labels,
		},
		Spec: ingressv1.InferenceIngressSpec{
			Domain:        domain,
			Framework:     string(request.Spec.Framework),
			IngressType:   "nginx",
			BypassGateway: false,
			Function:      request.Spec.Name,
			TLS: &ingressv1.InferenceIngressTLS{
				Enabled: true,
			},
		},
	}

	// Add HTTPS scheme to the domain.
	return ingress, nil
}
