// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// InstanceExec executes command in the instance.
func (cli *Client) InstanceExec(ctx context.Context,
	namespace, inferenceName, instance string, command []string, tty bool) (string, error) {
	urlValues := url.Values{}
	urlValues.Add("namespace", namespace)
	urlValues.Add("tty", fmt.Sprintf("%v", tty))
	urlValues.Add("command", strings.Join(command, ","))

	urlPath := fmt.Sprintf(gatewayInferInstanceExecControlPlanePath, inferenceName, instance)

	resp, err := cli.get(ctx, urlPath, urlValues, nil)
	defer ensureReaderClosed(resp)

	if err != nil {
		return "",
			wrapResponseError(err, resp, "instances with namespace", namespace)
	}

	res, err := io.ReadAll(resp.body)
	if err != nil {
		return "", wrapResponseError(err, resp, "instances with namespace", namespace)
	}

	return string(res), wrapResponseError(err, resp, "instances with namespace", namespace)
}

// InstanceExec executes command in the instance.
func (cli *Client) InstanceExecTTY(ctx context.Context,
	namespace, inferenceName, instance string, command []string,
) (HijackedResponse, error) {
	urlValues := url.Values{}
	urlValues.Add("namespace", namespace)
	urlValues.Add("tty", "true")
	urlValues.Add("command", strings.Join(command, ","))

	urlPath := fmt.Sprintf(gatewayInferInstanceExecControlPlanePath, inferenceName, instance)

	resp, err := cli.websocket(ctx, urlPath, urlValues, nil)
	if err != nil {
		return HijackedResponse{}, wrapResponseError(err, serverResponse{}, "instances with namespace", namespace)
	}

	return resp, nil
}
