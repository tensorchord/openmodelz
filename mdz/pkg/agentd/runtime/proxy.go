package runtime

import (
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tensorchord/openmodelz/agent/errdefs"
)

func (r *Runtime) InfereceProxy(c *gin.Context, name string) error {
	ctr, err := r.cli.ContainerInspect(c.Request.Context(), name)
	if err != nil {
		return errdefs.System(err)
	}

	if ctr.Config.Labels[labelVendor] != valueVendor {
		return errdefs.NotFound(errors.New("container not found"))
	}

	port := ""
	for _, c := range ctr.HostConfig.PortBindings {
		if len(c) > 0 {
			port = c[0].HostPort
			break
		}
	}

	if port == "" {
		return errdefs.NotFound(errors.New("port not found"))
	}

	url, err := url.Parse("http://" + Localhost + ":" + port)
	if err != nil {
		return errdefs.System(err)
	}
	proxyServer := httputil.NewSingleHostReverseProxy(url)

	proxyServer.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   time.Minute * 5,
			KeepAlive: time.Minute * 5,
			DualStack: true,
		}).DialContext,
	}
	proxyServer.Director = func(req *http.Request) {
		targetQuery := url.RawQuery
		req.URL.Scheme = url.Scheme
		req.URL.Host = url.Host
		// req.URL.Path, req.URL.RawPath = joinURLPath(target, req.URL)
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		req.URL.Path = c.Param("proxyPath")
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}
	}

	proxyServer.ServeHTTP(c.Writer, c.Request)
	return nil
}
