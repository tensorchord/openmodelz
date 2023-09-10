package runtime

import (
	"context"
	"path/filepath"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/errdefs"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r generalRuntime) ServerLabelCreate(ctx context.Context, name string, spec types.ServerSpec) error {
	node, err := r.kubeClient.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return errdefs.NotFound(err)
		} else {
			return errdefs.System(err)
		}
	}

	if len(node.Labels) == 0 {
		node.Labels = map[string]string{}
	}

	for k, v := range spec.Labels {
		node.Labels[filepath.Join("tensorchord.ai", k)] = v
	}

	_, err = r.kubeClient.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return errdefs.NotFound(err)
		} else {
			return errdefs.System(err)
		}
	}

	return nil
}
