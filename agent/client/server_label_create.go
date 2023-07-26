// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

// ServerLabelCreate create the labels for the servers.
func (cli *Client) ServerLabelCreate(ctx context.Context, name string,
	labels map[string]string) error {
	req := types.ServerSpec{
		Name:   name,
		Labels: labels,
	}

	urlValues := url.Values{}

	resp, err := cli.post(ctx,
		fmt.Sprintf(gatewayServerLabelCreateControlPlanePath, name), urlValues, req, nil)
	defer ensureReaderClosed(resp)

	if err != nil {
		return wrapResponseError(err, resp, "server", name)
	}

	return wrapResponseError(err, resp, "server", name)
}
