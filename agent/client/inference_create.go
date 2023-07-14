// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"context"
	"net/url"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

// InferenceCreate creates the inference.
func (cli *Client) InferenceCreate(ctx context.Context, namespace string,
	inference types.InferenceDeployment) (types.InferenceDeployment, error) {
	urlValues := url.Values{}
	urlValues.Add("namespace", namespace)

	resp, err := cli.post(ctx, gatewayInferControlPlanePath, urlValues, inference, nil)
	defer ensureReaderClosed(resp)

	if err != nil {
		return inference, wrapResponseError(err, resp, "inference", inference.Spec.Name)
	}

	return inference, wrapResponseError(err, resp, "inference", inference.Spec.Name)
}
