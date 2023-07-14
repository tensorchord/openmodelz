package runtime

import (
	"context"
	"strconv"

	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/errdefs"
)

func (r Runtime) InferenceScale(ctx context.Context, namespace string,
	req types.ScaleServiceRequest) (err error) {
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

	minReplicasStr := deployment.Annotations[consts.AnnotationMinReplicas]
	if minReplicasStr != "" {
		minReplicas, err := strconv.Atoi(minReplicasStr)
		if err != nil {
			return errdefs.InvalidParameter(err)
		}
		if replicas < int32(minReplicas) {
			replicas = int32(minReplicas)
		}
	}
	maxReplicasStr := deployment.Annotations[consts.AnnotationMaxReplicas]
	if maxReplicasStr != "" {
		maxReplicas, err := strconv.Atoi(maxReplicasStr)
		if err != nil {
			return errdefs.InvalidParameter(err)
		}
		if replicas > int32(maxReplicas) {
			replicas = int32(maxReplicas)
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
