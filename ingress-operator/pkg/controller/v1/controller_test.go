package v1

import (
	"reflect"
	"testing"

	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	faasv1 "github.com/tensorchord/openmodelz/ingress-operator/pkg/apis/modelzetes/v1"
	"github.com/tensorchord/openmodelz/ingress-operator/pkg/controller"
)

func Test_makeRules_Nginx_RootPath_HasRegex(t *testing.T) {
	ingress := faasv1.InferenceIngress{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
		},
		Spec: faasv1.InferenceIngressSpec{
			IngressType: "nginx",
		},
	}

	rules := makeRules(&ingress, "apiserver")

	if len(rules) == 0 {
		t.Errorf("Ingress should give at least one rule")
		t.Fail()
	}

	wantPath := "/(.*)"
	gotPath := rules[0].HTTP.Paths[0].Path

	if gotPath != wantPath {
		t.Errorf("want path %s, but got %s", wantPath, gotPath)
	}

	gotPort := rules[0].HTTP.Paths[0].Backend.Service.Port.Number

	if gotPort != controller.OpenfaasWorkloadPort {
		t.Errorf("want port %d, but got %d", controller.OpenfaasWorkloadPort, gotPort)
	}
}

func Test_makeRules_Nginx_RootPath_IsRootWithBypassMode(t *testing.T) {
	wantFunction := "apiserver"
	ingress := faasv1.InferenceIngress{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
		},
		Spec: faasv1.InferenceIngressSpec{
			BypassGateway: true,
			IngressType:   "nginx",
			Function:      "nodeinfo",
			// Path:          "/",
		},
	}

	rules := makeRules(&ingress, "apiserver")

	if len(rules) == 0 {
		t.Errorf("Ingress should give at least one rule")
		t.Fail()
	}

	wantPath := "/"
	gotPath := rules[0].HTTP.Paths[0].Path

	if gotPath != wantPath {
		t.Errorf("want path %s, but got %s", wantPath, gotPath)
	}

	gotHost := rules[0].HTTP.Paths[0].Backend.Service.Name

	if gotHost != wantFunction {
		t.Errorf("want host to be function: %s, but got %s", wantFunction, gotHost)
	}
}

func Test_makeRules_Nginx_PathOverride(t *testing.T) {
	ingress := faasv1.InferenceIngress{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
		},
		Spec: faasv1.InferenceIngressSpec{
			IngressType: "nginx",
			Path:        "/v1/profiles/view/(.*)",
		},
	}

	rules := makeRules(&ingress, "apiserver")

	if len(rules) == 0 {
		t.Errorf("Ingress should give at least one rule")
		t.Fail()
	}

	wantPath := ingress.Spec.Path
	gotPath := rules[0].HTTP.Paths[0].Path

	if gotPath != wantPath {
		t.Errorf("want path %s, but got %s", wantPath, gotPath)
	}
}

func Test_makeRules_Traefik_RootPath_TrimsRegex(t *testing.T) {
	ingress := faasv1.InferenceIngress{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
		},
		Spec: faasv1.InferenceIngressSpec{
			IngressType: "traefik",
		},
	}

	rules := makeRules(&ingress, "apiserver")

	if len(rules) == 0 {
		t.Errorf("Ingress should give at least one rule")
		t.Fail()
	}

	wantPath := "/"
	gotPath := rules[0].HTTP.Paths[0].Path
	if gotPath != wantPath {
		t.Errorf("want path %s, but got %s", wantPath, gotPath)
	}
}

func Test_makeRules_Traefik_NestedPath_TrimsRegex_And_TrailingSlash(t *testing.T) {
	ingress := faasv1.InferenceIngress{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
		},
		Spec: faasv1.InferenceIngressSpec{
			IngressType: "traefik",
			Path:        "/v1/profiles/view/(.*)",
		},
	}

	rules := makeRules(&ingress, "apiserver")

	if len(rules) == 0 {
		t.Errorf("Ingress should give at least one rule")
		t.Fail()
	}

	wantPath := "/v1/profiles/view"
	gotPath := rules[0].HTTP.Paths[0].Path
	if gotPath != wantPath {
		t.Errorf("want path %s, but got %s", wantPath, gotPath)
	}
}

func Test_makeTLS(t *testing.T) {

	cases := []struct {
		name     string
		fni      *faasv1.InferenceIngress
		expected []netv1.IngressTLS
	}{
		{
			name: "tls disabled results in empty tls config",
			fni: &faasv1.InferenceIngress{
				Spec: faasv1.InferenceIngressSpec{
					TLS: &faasv1.InferenceIngressTLS{
						Enabled: false,
					},
				},
			},
			expected: []netv1.IngressTLS{},
		},
		{
			name: "tls enabled creates TLS object with correct host and secret with matching the host",
			fni: &faasv1.InferenceIngress{
				Spec: faasv1.InferenceIngressSpec{
					Domain: "foo.example.com",
					TLS: &faasv1.InferenceIngressTLS{
						Enabled: true,
						IssuerRef: faasv1.ObjectReference{
							Name: "test-issuer",
							Kind: "ClusterIssuer",
						},
					},
				},
			},
			expected: []netv1.IngressTLS{
				{
					Hosts: []string{
						"foo.example.com",
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := makeTLS(tc.fni)
			if !reflect.DeepEqual(tc.expected, got) {
				t.Fatalf("want tls config %v, got %v", tc.expected, got)
			}
		})
	}
}
