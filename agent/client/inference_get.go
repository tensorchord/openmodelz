// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

// InferenceGet gets the inference.
func (cli *Client) InferenceGet(ctx context.Context, namespace, name string) (types.InferenceDeployment, error) {
	urlValues := url.Values{}
	urlValues.Add("namespace", namespace)

	url := fmt.Sprintf("/system/inference/%s", name)
	resp, err := cli.get(ctx, url, urlValues, nil)
	defer ensureReaderClosed(resp)

	if err != nil {
		return types.InferenceDeployment{},
			wrapResponseError(err, resp, "inference", name)
	}

	var inference types.InferenceDeployment
	err = json.NewDecoder(resp.body).Decode(&inference)
	if err != nil {
		return types.InferenceDeployment{},
			wrapResponseError(err, resp, "inference", name)
	}

	return inference, wrapResponseError(err, resp, "inference", name)
}
