package client

import (
	"context"
	"net/url"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

// InferenceScale scales the inference.
func (cli *Client) InferenceScale(ctx context.Context, namespace string,
	name string, replicas int, eventMessage string) error {

	urlValues := url.Values{}
	urlValues.Add("namespace", namespace)

	req := types.ScaleServiceRequest{
		ServiceName:  name,
		Replicas:     uint64(replicas),
		EventMessage: eventMessage,
	}

	resp, err := cli.post(ctx, gatewayInferScaleControlPath, urlValues, req, nil)
	defer ensureReaderClosed(resp)
	return wrapResponseError(err, resp, "inference", name)
}
