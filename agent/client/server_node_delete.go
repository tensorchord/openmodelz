// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"context"
	"fmt"
	"net/url"
)

// ServerLabelCreate create the labels for the servers.
func (cli *Client) ServerNodeDelete(ctx context.Context, name string) error {
	urlValues := url.Values{}

	resp, err := cli.delete(ctx,
		fmt.Sprintf(gatewayServerNodeDeleteControlPlanePath, name), urlValues, nil, nil)
	defer ensureReaderClosed(resp)

	if err != nil {
		return wrapResponseError(err, resp, "server-delete", name)
	}

	return wrapResponseError(err, resp, "server-delete", name)
}
