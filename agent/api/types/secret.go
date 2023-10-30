package types

// Secret for underlying orchestrator
type Secret struct {
	// Name of the secret
	Name string `json:"name"`

	// Namespace if applicable for the secret
	Namespace string `json:"namespace,omitempty"`

	// Data contains the secret data. Each key must consist of alphanumeric
	// characters, '-', '_' or '.'. The serialized form of the secret data is a
	// base64 encoded string, representing the arbitrary (possibly non-string)
	// data value here. Described in https://tools.ietf.org/html/rfc4648#section-4
	Data map[string][]byte `json:"data,omitempty"`

	// stringData allows specifying non-binary secret data in string form.
	// It is provided as a write-only input field for convenience.
	// All keys and values are merged into the data field on write, overwriting any existing values.
	// The stringData field is never output when reading from the API.
	StringData map[string]string `json:"stringData,omitempty"`
}
