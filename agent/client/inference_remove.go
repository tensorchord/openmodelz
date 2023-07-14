// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"context"
	"net/url"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

// InferenceRemove removes the inference.
func (cli *Client) InferenceRemove(ctx context.Context, namespace string,
	name string) error {

	urlValues := url.Values{}
	urlValues.Add("namespace", namespace)

	req := types.DeleteFunctionRequest{
		FunctionName: name,
	}

	resp, err := cli.delete(ctx, gatewayInferControlPlanePath, urlValues, req, nil)
	defer ensureReaderClosed(resp)
	return wrapResponseError(err, resp, "inference", name)
}
