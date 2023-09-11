package k8s

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/apis/modelzetes/v2alpha1"
	"github.com/tensorchord/openmodelz/modelzetes/pkg/consts"
	. "github.com/tensorchord/openmodelz/modelzetes/pkg/pointer"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	v1types "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("agent/pkg/k8s/convert_inference", func() {
	It("function AsResourceList", func() {
		tcs := []struct {
			resource v1.ResourceList
			expect   types.ResourceList
		}{
			{
				resource: map[v1types.ResourceName]resource.Quantity{
					v1types.ResourceCPU:      resource.MustParse("0"),
					v1types.ResourceMemory:   resource.MustParse("0"),
					consts.ResourceNvidiaGPU: resource.MustParse("0"),
				},
				expect: types.ResourceList{},
			},
			{
				resource: map[v1types.ResourceName]resource.Quantity{
					v1types.ResourceCPU:      resource.MustParse("0"),
					v1types.ResourceMemory:   resource.MustParse("500m"),
					consts.ResourceNvidiaGPU: resource.MustParse("0"),
				},
				expect: types.ResourceList{
					types.ResourceMemory: types.Quantity("500m"),
				},
			},
			{
				resource: map[v1types.ResourceName]resource.Quantity{
					v1types.ResourceCPU:      resource.MustParse("0"),
					v1types.ResourceMemory:   resource.MustParse("0"),
					consts.ResourceNvidiaGPU: resource.MustParse("0.5"),
				},
				expect: types.ResourceList{
					types.ResourceGPU: types.Quantity("500m"),
				},
			},
			{
				resource: map[v1types.ResourceName]resource.Quantity{
					v1types.ResourceCPU:      resource.MustParse("0.1"),
					v1types.ResourceMemory:   resource.MustParse("0"),
					consts.ResourceNvidiaGPU: resource.MustParse("0"),
				},
				expect: types.ResourceList{
					types.ResourceCPU: types.Quantity("100m"),
				},
			},
		}
		for _, tc := range tcs {
			value := AsResourceList(tc.resource)
			Expect(value).To(Equal(tc.expect))
		}
	})
	It("function AsInferenceDeployment", func() {
		mockTime, _ := time.Parse("2006-01-02", "2023-09-07")
		tcs := []struct {
			inf        *v2alpha1.Inference
			deployment *appsv1.Deployment
			expect     *types.InferenceDeployment
		}{
			{
				inf:        nil,
				deployment: nil,
				expect:     nil,
			},
			{
				inf: Ptr(v2alpha1.Inference{}),
				deployment: Ptr(appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: metav1.Time{
							Time: mockTime,
						},
					},
				}),
				expect: Ptr(types.InferenceDeployment{
					Status: types.InferenceDeploymentStatus{
						Phase:     types.PhaseNotReady,
						CreatedAt: Ptr(mockTime),
					},
				}),
			},
			{
				inf: Ptr(v2alpha1.Inference{
					Spec: v2alpha1.InferenceSpec{
						Scaling: Ptr(v2alpha1.ScalingConfig{
							Type: Ptr(v2alpha1.ScalingTypeCapacity),
						}),
					},
				}),
				deployment: Ptr(appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						CreationTimestamp: metav1.Time{
							Time: mockTime,
						},
					},
				}),
				expect: Ptr(types.InferenceDeployment{
					Spec: types.InferenceDeploymentSpec{
						Scaling: Ptr(types.ScalingConfig{
							Type: Ptr(types.ScalingTypeCapacity),
						}),
					},
					Status: types.InferenceDeploymentStatus{
						Phase:     types.PhaseNotReady,
						CreatedAt: Ptr(mockTime),
					},
				}),
			},
		}
		for _, tc := range tcs {
			value := AsInferenceDeployment(tc.inf, tc.deployment)
			Expect(value).To(Equal(tc.expect))
		}
	})
})
