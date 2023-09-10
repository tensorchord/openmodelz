package runtime

import (
	"context"

	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/errdefs"
)

func (r generalRuntime) InferenceScale(ctx context.Context, namespace string,
	req types.ScaleServiceRequest, inf *types.InferenceDeployment) (err error) {
	options := metav1.GetOptions{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
	}

	deployment, err := r.kubeClient.AppsV1().Deployments(namespace).
		Get(ctx, req.ServiceName, options)
	if err != nil {
		return errdefs.InvalidParameter(err)
	}

	oldReplicas := *deployment.Spec.Replicas
	replicas := int32(req.Replicas)

	if inf.Spec.Scaling != nil {
		minReplicas := *inf.Spec.Scaling.MinReplicas
		if replicas < minReplicas {
			replicas = minReplicas
		}

		maxReplicas := *inf.Spec.Scaling.MaxReplicas
		if replicas > maxReplicas {
			replicas = maxReplicas
		}
	}

	if replicas >= consts.MaxReplicas {
		replicas = consts.MaxReplicas
	}

	if oldReplicas == replicas {
		return nil
	}
	event := types.DeploymentScaleDownEvent
	if oldReplicas < replicas {
		event = types.DeploymentScaleUpEvent
	}

	var building bool
	if r.buildEnabled {
		_, building = deployment.Annotations[consts.AnnotationBuilding]
	}

	if building {
		event = types.DeploymentScaleBlockEvent
		req.EventMessage = "Deployment is building image, scale is blocked"
		replicas = 0
	}

	if r.eventEnabled {
		err = r.eventRecorder.CreateDeploymentEvent(namespace, deployment.Name, event, req.EventMessage)
		if err != nil {
			return err
		}
	}

	deployment.Spec.Replicas = &replicas
	r.logger.WithField("deployment", deployment.Name).
		WithField("namespace", namespace).
		WithField("replicas", replicas).Debug("scaling deployment")

	if _, err = r.kubeClient.AppsV1().Deployments(namespace).
		Update(ctx, deployment, metav1.UpdateOptions{}); err != nil {
		return errdefs.System(err)
	}

	return nil
}
