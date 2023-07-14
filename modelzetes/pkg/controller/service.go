package controller

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"

	v2alpha1 "github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
)

// newService creates a new ClusterIP Service for a Function resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Function resource that 'owns' it.
func newService(function *v2alpha1.Inference) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:        consts.DefaultServicePrefix + function.Spec.Name,
			Namespace:   function.Namespace,
			Annotations: map[string]string{"prometheus.io.scrape": "false"},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(function, schema.GroupVersionKind{
					Group:   v2alpha1.SchemeGroupVersion.Group,
					Version: v2alpha1.SchemeGroupVersion.Version,
					Kind:    v2alpha1.Kind,
				}),
			},
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: map[string]string{consts.LabelInferenceName: function.Spec.Name},
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Protocol: corev1.ProtocolTCP,
					Port:     functionPort,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: int32(makePort(function)),
					},
				},
			},
		},
	}
}
