package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/sirupsen/logrus"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

func (cli *Client) BuildCreate(ctx context.Context, namespace string, build types.Build) error {
	build.Spec.Namespace = namespace
	logrus.Debugf("create new build: %s", build)

	val := url.Values{}
	resp, err := cli.post(ctx, gatewayBuildControlPlanePath, val, build, nil)
	defer ensureReaderClosed(resp)

	if err != nil {
		return wrapResponseError(err, resp, "build", build.Spec.Name)
	}

	return nil
}

func (cli *Client) BuildGet(ctx context.Context, namespace, name string) (types.Build, error) {
	val := url.Values{}
	val.Add("namespace", namespace)
	build := types.Build{}
	resp, err := cli.get(
		ctx, fmt.Sprintf(gatewayBuildInstanceControlPlanePath, name), val, nil)
	defer ensureReaderClosed(resp)
	if err != nil {
		logrus.Infof("failed to query build.get: %s", err)
		return build, wrapResponseError(err, resp, "build", name)
	}

	err = json.NewDecoder(resp.body).Decode(&build)
	if err != nil {
		logrus.Infof("failed to decode build: %s", err)
		return build, wrapResponseError(err, resp, "build", name)
	}
	return build, nil
}

func (cli *Client) BuildList(ctx context.Context, namespace string) ([]types.Build, error) {
	val := url.Values{}
	val.Add("namespace", namespace)
	resp, err := cli.get(ctx, gatewayBuildControlPlanePath, val, nil)
	defer ensureReaderClosed(resp)
	if err != nil {
		logrus.Infof("failed to query build.list: %s", err)
		return nil, wrapResponseError(err, resp, "build", namespace)
	}

	var builds []types.Build
	err = json.NewDecoder(resp.body).Decode(&builds)
	if err != nil {
		logrus.Infof("failed to decode builds: %s", err)
		return nil, wrapResponseError(err, resp, "build", namespace)
	}
	return builds, nil
}
