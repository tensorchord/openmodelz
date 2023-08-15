// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/agent/api/types"
)

const LogBufferSize = 128

// DeploymentLogGet gets the deployment logs.
func (cli *Client) DeploymentLogGet(ctx context.Context, namespace, name string,
	since string, tail int, end string, follow bool) (
	<-chan types.Message, error) {
	urlValues := url.Values{}
	urlValues.Add("namespace", namespace)
	urlValues.Add("name", name)

	if since != "" {
		urlValues.Add("since", since)
	}

	if end != "" {
		urlValues.Add("end", end)
	}

	if tail != 0 {
		urlValues.Add("tail", fmt.Sprintf("%d", tail))
	}

	if follow {
		urlValues.Add("follow", "true")
	}

	resp, err := cli.get(ctx, "/system/logs/inference", urlValues, nil)

	if err != nil {
		return nil, wrapResponseError(err, resp, "deployment logs", name)
	}

	stream := make(chan types.Message, LogBufferSize)
	var log types.Message
	scanner := bufio.NewScanner(resp.body)
	go func() {
		defer ensureReaderClosed(resp)
		defer close(stream)
		for scanner.Scan() {
			err = json.Unmarshal(scanner.Bytes(), &log)
			if err != nil {
				logrus.Warnf("failed to decode %s log: %v | %s | [%s]", name, err, scanner.Text(), scanner.Err())
				return
				// continue
			}
			stream <- log
		}
	}()

	return stream, err
}

func (cli *Client) BuildLogGet(ctx context.Context, namespace, name, since string,
	tail int) ([]types.Message, error) {
	urlValues := url.Values{}
	urlValues.Add("namespace", namespace)
	urlValues.Add("name", name)

	if since != "" {
		urlValues.Add("since", since)
	}
	if tail != 0 {
		urlValues.Add("tail", fmt.Sprintf("%d", tail))
	}

	resp, err := cli.get(ctx, "/system/logs/build", urlValues, nil)
	defer ensureReaderClosed(resp)

	if err != nil {
		return nil,
			wrapResponseError(err, resp, "build logs", name)
	}

	var log types.Message
	logs := []types.Message{}
	scanner := bufio.NewScanner(resp.body)
	for scanner.Scan() {
		err = json.NewDecoder(strings.NewReader(scanner.Text())).Decode(&log)
		if err != nil {
			return nil, wrapResponseError(err, resp, "build logs", name)
		}
		logs = append(logs, log)
	}

	return logs, err
}
