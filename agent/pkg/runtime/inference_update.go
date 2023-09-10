package runtime

import (
	"context"

	"github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	inferenceclientset "github.com/tensorchord/openmodelz/modelzetes/pkg/client/clientset/versioned"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/errdefs"
)

func (r generalRuntime) InferenceUpdate(ctx context.Context, namespace string,
	req types.InferenceDeployment, event string) (err error) {

	if r.eventEnabled {
		err := r.eventRecorder.CreateDeploymentEvent(namespace, req.Spec.Name, event, req.Status.EventMessage)
		if err != nil {
			return err
		}
	}

	if err = updateInference(ctx, namespace, r.inferenceClient, req); err != nil {
		return err
	}
	return nil
}

func updateInference(
	ctx context.Context,
	functionNamespace string,
	inferenceClient inferenceclientset.Interface,
	request types.InferenceDeployment) (err error) {

	actual, err := inferenceClient.TensorchordV2alpha1().
		Inferences(functionNamespace).Get(
		ctx, request.Spec.Name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return errdefs.NotFound(err)
		} else {
			return errdefs.System(err)
		}
	}

	expected := actual.DeepCopy()

	if request.Spec.Image != "" {
		expected.Spec.Image = request.Spec.Image
	}
	if request.Spec.Scaling != nil {
		expected.Spec.Scaling = &v2alpha1.ScalingConfig{
			MinReplicas:     request.Spec.Scaling.MinReplicas,
			MaxReplicas:     request.Spec.Scaling.MaxReplicas,
			TargetLoad:      request.Spec.Scaling.TargetLoad,
			ZeroDuration:    request.Spec.Scaling.ZeroDuration,
			StartupDuration: request.Spec.Scaling.StartupDuration,
		}
		if request.Spec.Scaling.Type != nil {
			expected.Spec.Scaling.Type = new(v2alpha1.ScalingType)
			*expected.Spec.Scaling.Type = v2alpha1.ScalingType(*request.Spec.Scaling.Type)
		}
	}
	if request.Spec.EnvVars != nil {
		expected.Spec.EnvVars = request.Spec.EnvVars
	}
	if request.Spec.Secrets != nil {
		expected.Spec.Secrets = request.Spec.Secrets
	}
	if request.Spec.Constraints != nil {
		expected.Spec.Constraints = request.Spec.Constraints
	}
	if request.Spec.Labels != nil {
		expected.Spec.Labels = request.Spec.Labels
	}
	if request.Spec.Annotations != nil {
		expected.Spec.Annotations = request.Spec.Annotations
	}
	if request.Spec.Resources != nil {
		rr, err := createResources(request)
		if err != nil {
			return errdefs.InvalidParameter(err)
		}
		expected.Spec.Resources = &rr
	}

	if _, err := inferenceClient.TensorchordV2alpha1().
		Inferences(functionNamespace).Update(
		ctx, expected, metav1.UpdateOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			return errdefs.NotFound(err)
		} else {
			return errdefs.System(err)
		}
	}

	return nil
}
