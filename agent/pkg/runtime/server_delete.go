package runtime

import (
	"context"

	"github.com/tensorchord/openmodelz/agent/errdefs"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r generalRuntime) ServerDeleteNode(ctx context.Context, name string) error {
	err := r.kubeClient.CoreV1().Nodes().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return errdefs.NotFound(err)
		}
		return errdefs.System(err)
	}
	return nil
}
