// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"context"
	"encoding/json"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

// InfoGet gets the agent info.
func (cli *Client) InfoGet(ctx context.Context) (types.ProviderInfo, error) {
	resp, err := cli.get(ctx, "/system/info", nil, nil)
	defer ensureReaderClosed(resp)

	if err != nil {
		return types.ProviderInfo{},
			wrapResponseError(err, resp, "info", "system")
	}

	var info types.ProviderInfo
	err = json.NewDecoder(resp.body).Decode(&info)
	if err != nil {
		return types.ProviderInfo{},
			wrapResponseError(err, resp, "info", "system")
	}

	return info, wrapResponseError(err, resp, "info", "system")
}
