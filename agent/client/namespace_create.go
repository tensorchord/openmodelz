// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"context"
	"net/url"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

// NamespaceCreate creates the namespace.
func (cli *Client) NamespaceCreate(ctx context.Context,
	namespace string) error {
	req := types.NamespaceRequest{
		Name: namespace,
	}

	urlValues := url.Values{}

	resp, err := cli.post(ctx, gatewayNamespaceControlPlanePath, urlValues, req, nil)
	defer ensureReaderClosed(resp)

	return wrapResponseError(err, resp, "namespace", namespace)
}
