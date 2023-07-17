// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"context"
	"encoding/json"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

// ServerList lists the servers.
func (cli *Client) ServerList(ctx context.Context) ([]types.Server, error) {
	resp, err := cli.get(ctx, gatewayServerControlPlanePath, nil, nil)
	defer ensureReaderClosed(resp)

	if err != nil {
		return nil,
			wrapResponseError(err, resp, "servers", "")
	}

	var servers []types.Server
	err = json.NewDecoder(resp.body).Decode(&servers)

	return servers, wrapResponseError(err, resp, "servers", "")
}
