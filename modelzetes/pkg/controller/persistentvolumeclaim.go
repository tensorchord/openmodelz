package controller

import (
	v2alpha1 "github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	glog "k8s.io/klog"
)

func makePersistentVolumeClaimName(name string) string {
	return name + "-pvc"
}

func makePersistentVolumeClaim(function *v2alpha1.Inference, volume v2alpha1.VolumeConfig) *corev1.PersistentVolumeClaim {
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      makePersistentVolumeClaimName(volume.Name),
			Namespace: function.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(function, schema.GroupVersionKind{
					Group:   v2alpha1.SchemeGroupVersion.Group,
					Version: v2alpha1.SchemeGroupVersion.Version,
					Kind:    v2alpha1.Kind,
				}),
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			VolumeName: makePersistentVolumeName(volume.Name),
		},
	}
	switch volume.Type {
	case v2alpha1.VolumeTypeGCSFuse:
		sc := consts.GCSStorageClassName
		pvc.Spec.StorageClassName = &sc
		pvc.Spec.Resources = corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(capacity),
			},
		}
	case v2alpha1.VolumeTypeLocal:
		sc := consts.LocalStorageClassName
		pvc.Spec.StorageClassName = &sc
		pvc.Spec.Resources = corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse(capacity),
			},
		}
	default:
		glog.Errorf("unknown volume type %s", volume.Type)
		return nil
	}
	return pvc
}
