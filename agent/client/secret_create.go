package client

import (
	"context"
	"net/url"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

func (cli *Client) SecretCreate(ctx context.Context, secret types.Secret) error {
	urlValues := url.Values{}

	resp, err := cli.post(ctx, gatewaySecretControlPlanePath, urlValues, secret, nil)
	defer ensureReaderClosed(resp)
	return wrapResponseError(err, resp, "secret", secret.Namespace+"/"+secret.Name)
}
