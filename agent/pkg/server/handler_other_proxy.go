package server

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tensorchord/openmodelz/agent/pkg/consts"
)

// @Summary     Reverse proxy to the backend other.
// @Description Reverse proxy to the backend other.
// @Tags        inference
// @Accept      */*
// @Produce     json
// @Param       id path string true "Deployment ID"
// @Router      /other/{id} [get]
// @Router      /other/{id} [post]
// @Success     201
func (s *Server) proxyOther(c *gin.Context) error {
	remote, err := url.Parse(fmt.Sprintf("http://0.0.0.0:%d", s.config.Server.ServerPort))
	if err != nil {
		return err
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   s.config.ModelZCloud.UpstreamTimeout,
			KeepAlive: s.config.ModelZCloud.UpstreamTimeout,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          s.config.ModelZCloud.MaxIdleConnections,
		MaxIdleConnsPerHost:   s.config.ModelZCloud.MaxIdleConnectionsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	uid, deployment, err := s.proxyAuth(c)
	if err != nil {
		return err
	}

	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = path.Join(
			"/", "inference", fmt.Sprintf("%s.%s", deployment, consts.DefaultPrefix+uid), c.Param("proxyPath"))
	}

	proxy.ServeHTTP(c.Writer, c.Request)
	return nil
}
