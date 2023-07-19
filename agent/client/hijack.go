package client // import "docker.io/go-docker"

import (
	"net/url"

	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
)

// HijackedResponse holds connection information for a hijacked request.
type HijackedResponse struct {
	Conn *websocket.Conn
}

// Close closes the hijacked connection and reader.
func (h *HijackedResponse) Close() {
	h.Conn.Close()
}

// postHijacked sends a POST request and hijacks the connection.
func (cli *Client) websocket(ctx context.Context, path string, query url.Values, headers map[string][]string) (HijackedResponse, error) {
	apiPath := cli.getAPIPath(ctx, path, nil)

	scheme := "ws"
	if cli.scheme == "https" {
		scheme = "wss"
	}

	apiURL := url.URL{
		Scheme:   scheme,
		Host:     cli.addr,
		Path:     apiPath,
		RawQuery: query.Encode(),
	}
	c, _, err := websocket.DefaultDialer.DialContext(ctx, apiURL.String(), nil)
	if err != nil {
		return HijackedResponse{}, err
	}

	return HijackedResponse{Conn: c}, err
}
