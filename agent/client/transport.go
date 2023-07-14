// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client // import "github.com/docker/docker/client"

import (
	"crypto/tls"
	"net/http"
)

// resolveTLSConfig attempts to resolve the TLS configuration from the
// RoundTripper.
func resolveTLSConfig(transport http.RoundTripper) *tls.Config {
	switch tr := transport.(type) {
	case *http.Transport:
		return tr.TLSClientConfig
	default:
		return nil
	}
}
