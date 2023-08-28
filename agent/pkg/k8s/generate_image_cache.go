package k8s

import (
	"time"

	kubefledged "github.com/senthilrch/kube-fledged/pkg/apis/kubefledged/v1alpha3"
	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/pkg/consts"
	modelzetes "github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func MakeImageCache(req types.ImageCache, inference *modelzetes.Inference) *kubefledged.ImageCache {
	nodeSlector := map[string]string{
		consts.LabelServerResource: string(req.NodeSelector),
	}
	cache := &kubefledged.ImageCache{
		ObjectMeta: v1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(inference, schema.GroupVersionKind{
					Group:   modelzetes.SchemeGroupVersion.Group,
					Version: modelzetes.SchemeGroupVersion.Version,
					Kind:    modelzetes.Kind,
				}),
			},
		},
		Spec: kubefledged.ImageCacheSpec{
			CacheSpec: []kubefledged.CacheSpecImages{
				{
					Images: []kubefledged.Image{
						{
							Name:           req.Image,
							ForceFullCache: req.ForceFullCache,
						},
					},
					NodeSelector: nodeSlector,
				},
			},
		},
		Status: kubefledged.ImageCacheStatus{
			StartTime: &metav1.Time{Time: time.Now()},
		},
	}
	return cache
}
