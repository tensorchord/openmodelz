// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package client

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/docker/go-connections/sockets"
)

// Refer to github.com/docker/docker/client

// ErrRedirect is the error returned by checkRedirect when the request is non-GET.
var ErrRedirect = errors.New("unexpected redirect in response")

// Client is the API client that performs all operations
// against a docker server.
type Client struct {
	// scheme sets the scheme for the client
	scheme string
	// host holds the server address to connect to
	host string
	// proto holds the client protocol i.e. unix.
	proto string
	// addr holds the client address.
	addr string
	// basePath holds the path to prepend to the requests.
	basePath string
	// client used to send and receive http requests.
	client *http.Client
	// version of the server to talk to.
	version string
	// custom http headers configured by users.
	customHTTPHeaders map[string]string
	// manualOverride is set to true when the version was set by users.
	manualOverride bool

	// negotiateVersion indicates if the client should automatically negotiate
	// the API version to use when making requests. API version negotiation is
	// performed on the first request, after which negotiated is set to "true"
	// so that subsequent requests do not re-negotiate.
	negotiateVersion bool

	// negotiated indicates that API version negotiation took place
	negotiated bool
}

// NewClientWithOpts initializes a new API client with a default HTTPClient, and
// default API host and version. It also initializes the custom HTTP headers to
// add to each request.
//
// It takes an optional list of Opt functional arguments, which are applied in
// the order they're provided, which allows modifying the defaults when creating
// the client. For example, the following initializes a client that configures
// itself with values from environment variables (client.FromEnv), and has
// automatic API version negotiation enabled (client.WithAPIVersionNegotiation()).
//
//	cli, err := client.NewClientWithOpts(
//		client.FromEnv,
//		client.WithAPIVersionNegotiation(),
//	)
func NewClientWithOpts(ops ...Opt) (*Client, error) {
	client, err := defaultHTTPClient(DefaultModelzGatewayHost)
	if err != nil {
		return nil, err
	}
	c := &Client{
		host:     DefaultModelzGatewayHost,
		version:  "",
		client:   client,
		proto:    defaultProto,
		addr:     defaultAddr,
		basePath: apiBasePath,
	}

	for _, op := range ops {
		if err := op(c); err != nil {
			return nil, err
		}
	}

	if c.scheme == "" {
		c.scheme = "http"

		tlsConfig := resolveTLSConfig(c.client.Transport)
		if tlsConfig != nil {
			// TODO(stevvooe): This isn't really the right way to write clients in Go.
			// `NewClient` should probably only take an `*http.Client` and work from there.
			// Unfortunately, the model of having a host-ish/url-thingy as the connection
			// string has us confusing protocol and transport layers. We continue doing
			// this to avoid breaking existing clients but this should be addressed.
			c.scheme = "https"
		}
	}

	return c, nil
}

func defaultHTTPClient(host string) (*http.Client, error) {
	hostURL, err := ParseHostURL(host)
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{}
	_ = sockets.ConfigureTransport(transport, hostURL.Scheme, hostURL.Host)
	return &http.Client{
		Transport:     transport,
		CheckRedirect: CheckRedirectKeepHeader,
	}, nil
}

// CheckRedirect specifies the policy for dealing with redirect responses:
// If the request is non-GET return ErrRedirect, otherwise use the last response.
//
// Go 1.8 changes behavior for HTTP redirects (specifically 301, 307, and 308)
// in the client. The envd client (and by extension envd API client) can be
// made to send a request like POST /containers//start where what would normally
// be in the name section of the URL is empty. This triggers an HTTP 301 from
// the daemon.
//
// In go 1.8 this 301 will be converted to a GET request, and ends up getting
// a 404 from the daemon. This behavior change manifests in the client in that
// before, the 301 was not followed and the client did not generate an error,
// but now results in a message like Error response from daemon: page not found.
func CheckRedirect(req *http.Request, via []*http.Request) error {
	if via[0].Method == http.MethodGet {
		return http.ErrUseLastResponse
	}
	return ErrRedirect
}

func CheckRedirectKeepHeader(req *http.Request, via []*http.Request) error {
	req.Header = via[0].Header.Clone()
	return nil
}

// DaemonHost returns the host address used by the client
func (cli *Client) DaemonHost() string {
	return cli.host
}

// HTTPClient returns a copy of the HTTP client bound to the server
func (cli *Client) HTTPClient() *http.Client {
	c := *cli.client
	return &c
}

// ParseHostURL parses a url string, validates the string is a host url, and
// returns the parsed URL
func ParseHostURL(host string) (*url.URL, error) {
	protoAddrParts := strings.SplitN(host, "://", 2)
	if len(protoAddrParts) == 1 {
		return nil, errors.Errorf("unable to parse docker host `%s`", host)
	}

	var basePath string
	proto, addr := protoAddrParts[0], protoAddrParts[1]
	if proto == "tcp" {
		parsed, err := url.Parse("tcp://" + addr)
		if err != nil {
			return nil, err
		}
		addr = parsed.Host
		basePath = parsed.Path
	}
	return &url.URL{
		Scheme: proto,
		Host:   addr,
		Path:   basePath,
	}, nil
}

// Close the transport used by the client
func (cli *Client) Close() error {
	if t, ok := cli.client.Transport.(*http.Transport); ok {
		t.CloseIdleConnections()
	}
	return nil
}

// getAPIPath returns the versioned request path to call the api.
// It appends the query parameters to the path if they are not empty.
func (cli *Client) getAPIPath(ctx context.Context, p string, query url.Values) string {
	var apiPath string
	if cli.version != "" {
		v := strings.TrimPrefix(cli.version, "v")
		apiPath = path.Join(cli.basePath, "/v"+v, p)
	} else {
		apiPath = path.Join(cli.basePath, p)
	}
	return (&url.URL{Path: apiPath, RawQuery: query.Encode()}).String()
}
