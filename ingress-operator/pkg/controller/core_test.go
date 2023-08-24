package controller

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	faasv1 "github.com/tensorchord/openmodelz/ingress-operator/pkg/apis/modelzetes/v1"
)

func TestMakeAnnotations(t *testing.T) {
	cases := []struct {
		name     string
		ingress  faasv1.InferenceIngress
		expected map[string]string
		excluded []string
	}{
		{
			name: "base case, annotations are copied, default class is nginx",
			ingress: faasv1.InferenceIngress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"test":    "test",
						"example": "example",
					},
				},
			},
			expected: map[string]string{
				"test":                        "test",
				"example":                     "example",
				"kubernetes.io/ingress.class": "nginx",
			},
		},
		{
			name: "can override ingress class value",
			ingress: faasv1.InferenceIngress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "awesome-nginx",
					},
				},
				Spec: faasv1.InferenceIngressSpec{
					IngressType: "awesome-nginx",
				},
			},
			expected: map[string]string{
				"kubernetes.io/ingress.class": "awesome-nginx",
			},
		},
		{
			name: "bypass removes rewrite target",
			ingress: faasv1.InferenceIngress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "nginx",
					},
				},
				Spec: faasv1.InferenceIngressSpec{
					IngressType:   "nginx",
					Function:      "nodeinfo",
					BypassGateway: true,
					Domain:        "nodeinfo.example.com",
				},
			},
			excluded: []string{"nginx.ingress.kubernetes.io/rewrite-target"},
		},
		{
			name: "default annotations includes a rewrite-target",
			ingress: faasv1.InferenceIngress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: faasv1.InferenceIngressSpec{
					IngressType: "nginx",
				},
			},
			expected: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target": "/function//$1",
			},
		},
		{
			name: "creates required traefik annotations",
			ingress: faasv1.InferenceIngress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "traefik",
					},
				},
				Spec: faasv1.InferenceIngressSpec{
					IngressType:   "traefik",
					Function:      "nodeinfo",
					BypassGateway: false,
					Domain:        "nodeinfo.example.com",
				},
			},
			expected: map[string]string{
				"traefik.ingress.kubernetes.io/rewrite-target": "/function/nodeinfo",
				"traefik.ingress.kubernetes.io/rule-type":      "PathPrefix",
			},
		},
		{
			name: "creates required skipper annotations",
			ingress: faasv1.InferenceIngress{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"kubernetes.io/ingress.class": "skipper",
					},
				},
				Spec: faasv1.InferenceIngressSpec{
					IngressType:   "skipper",
					Function:      "nodeinfo",
					BypassGateway: false,
					Domain:        "nodeinfo.example.com",
				},
			},
			expected: map[string]string{
				"kubernetes.io/ingress.class": "skipper",
				"zalando.org/skipper-filter":  `setPath("/function/nodeinfo")`,
			},
		},
		// {
		// 	name: "creates tls issuer annotation",
		// 	ingress: faasv1.InferenceIngress{
		// 		ObjectMeta: metav1.ObjectMeta{
		// 			Annotations: map[string]string{
		// 				"kubernetes.io/ingress.class": "nginx",
		// 			},
		// 		},
		// 		Spec: faasv1.InferenceIngressSpec{
		// 			IngressType:   "nginx",
		// 			Function:      "nodeinfo",
		// 			BypassGateway: false,
		// 			Domain:        "nodeinfo.example.com",
		// 			TLS: &faasv1.InferenceIngressTLS{
		// 				IssuerRef: faasv1.ObjectReference{
		// 					Name: "clusterFoo",
		// 					Kind: "ClusterIssuer",
		// 				},
		// 				Enabled: true,
		// 			},
		// 		},
		// 	},
		// 	expected: map[string]string{
		// 		"cert-manager.io/cluster-issuer": "clusterFoo",
		// 	},
		// },
		// {
		// 	name: "default tls issuer is local",
		// 	ingress: faasv1.InferenceIngress{
		// 		ObjectMeta: metav1.ObjectMeta{
		// 			Annotations: map[string]string{
		// 				"kubernetes.io/ingress.class": "nginx",
		// 			},
		// 		},
		// 		Spec: faasv1.InferenceIngressSpec{
		// 			IngressType:   "nginx",
		// 			Function:      "nodeinfo",
		// 			BypassGateway: false,
		// 			Domain:        "nodeinfo.example.com",
		// 			TLS: &faasv1.InferenceIngressTLS{
		// 				IssuerRef: faasv1.ObjectReference{
		// 					Name: "clusterFoo",
		// 				},
		// 				Enabled: true,
		// 			},
		// 		},
		// 	},
		// 	expected: map[string]string{
		// 		"cert-manager.io/issuer": "clusterFoo",
		// 	},
		// },
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := MakeAnnotations(&tc.ingress, "apiserver")
			for key, value := range tc.expected {
				found, ok := result[key]
				if !ok {
					t.Fatalf("Failed to find expected annotation: %q", key)
				}

				if found != value {
					t.Fatalf("expected annotation value %q, got %q", value, found)
				}
			}

			for _, key := range tc.excluded {
				value, ok := result[key]
				if ok {
					t.Fatalf("annotations should not include %q, but it was found with value %q", key, value)
				}
			}
		})
	}
}
