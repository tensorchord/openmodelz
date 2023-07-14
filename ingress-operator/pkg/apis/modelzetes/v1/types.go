package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:name="Domain",type=string,JSONPath=`.spec.domain`

// InferenceIngress describes an OpenFaaS function
type InferenceIngress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InferenceIngressSpec `json:"spec"`
}

// InferenceIngressSpec is the spec for a InferenceIngressSpec resource. It must
// be created in the same namespace as the gateway, i.e. openfaas.
type InferenceIngressSpec struct {
	// Domain such as "api.example.com"
	Domain string `json:"domain"`

	// Function such as "nodeinfo"
	Function string `json:"function"`

	Framework string `json:"framework"`

	// Path such as "/v1/profiles/view/(.*)", or leave empty for default
	// +optional
	Path string `json:"path"`

	// IngressType such as "nginx"
	// +optional
	IngressType string `json:"ingressType,omitempty"`

	// Enable TLS via cert-manager
	// +optional
	TLS *InferenceIngressTLS `json:"tls,omitempty"`

	// BypassGateway, when true creates an Ingress record
	// directly for the Function name without using the gateway
	// in the hot path
	// +optional
	BypassGateway bool `json:"bypassGateway,omitempty"`
}

// InferenceIngressSpec TLS options
type InferenceIngressTLS struct {
	// +optional
	Enabled bool `json:"enabled"`

	// +optional
	IssuerRef ObjectReference `json:"issuerRef"`
}

// UseTLS if TLS is enabled
func (f *InferenceIngressSpec) UseTLS() bool {
	return f.TLS != nil && f.TLS.Enabled
}

// ObjectReference is a reference to an object with a given name and kind.
type ObjectReference struct {
	Name string `json:"name"`

	// +optional
	Kind string `json:"kind,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InferenceIngress is a list of Function resources
type InferenceIngressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []InferenceIngress `json:"items"`
}
