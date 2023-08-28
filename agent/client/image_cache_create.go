package client

import (
	"context"
	"net/url"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

func (cli *Client) ImageCacheCreate(ctx context.Context, namespace string,
	imageCache *types.ImageCache) error {
	urlValues := url.Values{}
	urlValues.Add("namespace", namespace)

	resp, err := cli.post(ctx, gatewayImageCacheControlPlanePath, urlValues, imageCache, nil)
	defer ensureReaderClosed(resp)
	if err != nil {
		return wrapResponseError(err, resp, "imagecache", imageCache.Name)
	}

	return wrapResponseError(err, resp, "imagecache", imageCache.Name)
}
