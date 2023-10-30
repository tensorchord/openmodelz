package controller

import (
	"github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	glog "k8s.io/klog"
)

const (
	capacity = "100Gi"
)

func makePersistentVolumeName(name string) string {
	return name + "-pv"
}

func newPersistentVolume(function *v2alpha1.Inference, volume v2alpha1.VolumeConfig) *corev1.PersistentVolume {
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: makePersistentVolumeName(volume.Name),
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(function, schema.GroupVersionKind{
					Group:   v2alpha1.SchemeGroupVersion.Group,
					Version: v2alpha1.SchemeGroupVersion.Version,
					Kind:    v2alpha1.Kind,
				}),
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
		},
	}

	switch volume.Type {
	case v2alpha1.VolumeTypeGCSFuse:
		pv.Spec.Capacity = corev1.ResourceList{
			corev1.ResourceStorage: resource.MustParse(capacity),
		}
		pv.Spec.StorageClassName = consts.GCSStorageClassName
		csi := &corev1.CSIPersistentVolumeSource{
			Driver:       consts.GCSCSIDriverName,
			VolumeHandle: consts.GCSVolumeHandle,
		}
		if volume.SecretName != nil {
			csi.NodePublishSecretRef = &corev1.SecretReference{
				Name:      *volume.SecretName,
				Namespace: function.Namespace,
			}
		}
		pv.Spec.CSI = csi
	case v2alpha1.VolumeTypeLocal:
		pv.Spec.Capacity = corev1.ResourceList{
			corev1.ResourceStorage: resource.MustParse(capacity),
		}
		mode := corev1.PersistentVolumeFilesystem
		pv.Spec.VolumeMode = &mode
		pv.Spec.StorageClassName = consts.LocalStorageClassName
		pv.Spec.Local = &corev1.LocalVolumeSource{
			Path: *volume.SubPath,
		}
		pv.Spec.NodeAffinity = &corev1.VolumeNodeAffinity{
			Required: &corev1.NodeSelector{
				NodeSelectorTerms: []corev1.NodeSelectorTerm{
					{
						MatchExpressions: []corev1.NodeSelectorRequirement{
							{
								Key:      corev1.LabelHostname,
								Operator: corev1.NodeSelectorOpIn,
								Values:   volume.NodeNames,
							},
						},
					},
				},
			},
		}
	default:
		glog.Errorf("unknown volume type: %s", volume.Type)
		return nil
	}

	return pv
}
