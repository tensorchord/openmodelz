// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

// InferenceList lists the inferences.
func (cli *Client) InferenceList(ctx context.Context, namespace string) ([]types.InferenceDeployment, error) {
	urlValues := url.Values{}
	urlValues.Add("namespace", namespace)

	resp, err := cli.get(ctx, gatewayInferControlPlanePath, urlValues, nil)
	defer ensureReaderClosed(resp)

	if err != nil {
		return nil,
			wrapResponseError(err, resp, "inferences with namespace", namespace)
	}

	var inferences []types.InferenceDeployment
	err = json.NewDecoder(resp.body).Decode(&inferences)

	return inferences, err
}
