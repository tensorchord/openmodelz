package runtime

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/errdefs"
)

func (r generalRuntime) NamespaceList(ctx context.Context) ([]string, error) {
	ns, err := r.kubeClient.CoreV1().Namespaces().List(ctx,
		metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=true", types.LabelNamespace),
		})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, nil
		} else {
			return nil, errdefs.System(err)
		}
	}

	res := make([]string, len(ns.Items))
	for i, n := range ns.Items {
		res[i] = n.Name
	}
	return res, nil
}

func (r generalRuntime) NamespaceCreate(ctx context.Context, name string) error {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Labels: map[string]string{
				types.LabelNamespace: "true",
			},
		},
	}

	_, err := r.kubeClient.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		if k8serrors.IsAlreadyExists(err) {
			return errdefs.Conflict(err)
		} else {
			return errdefs.System(err)
		}
	}

	return nil
}

func (r generalRuntime) NamespaceGet(ctx context.Context, name string) bool {
	_, err := r.kubeClient.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false
	}

	return true
}

func (r generalRuntime) NamespaceDelete(ctx context.Context, name string) error {
	err := r.kubeClient.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
