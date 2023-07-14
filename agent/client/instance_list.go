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

// InstanceList lists the deployment instances.
func (cli *Client) InstanceList(ctx context.Context,
	namespace, inferenceName string) ([]types.InferenceDeploymentInstance, error) {
	urlValues := url.Values{}
	urlValues.Add("namespace", namespace)

	urlPath := fmt.Sprintf(gatewayInferInstanceControlPlanePath, inferenceName)

	resp, err := cli.get(ctx, urlPath, urlValues, nil)
	defer ensureReaderClosed(resp)

	if err != nil {
		return nil,
			wrapResponseError(err, resp, "instances with namespace", namespace)
	}

	var instances []types.InferenceDeploymentInstance
	err = json.NewDecoder(resp.body).Decode(&instances)

	return instances, wrapResponseError(err, resp, "instances with namespace", namespace)
}
