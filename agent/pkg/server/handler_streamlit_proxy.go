package server

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/agent/pkg/consts"
	"github.com/tensorchord/openmodelz/agent/pkg/server/static"
)

// @Summary     Reverse proxy to streamlit.
// @Description Reverse proxy to streamlit.
// @Tags        inference
// @Accept      */*
// @Produce     json
// @Param       id path string true "Deployment ID"
// @Router      /streamlit/{id} [get]
// @Router      /streamlit/{id} [post]
// @Success     201
func (s *Server) proxyStreamlit(c *gin.Context) error {
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

	uid, deployment, err := s.proxyNoAuth(c)
	if err != nil {
		return err
	}

	ns := consts.DefaultPrefix + uid
	proxy.Director = func(req *http.Request) {
		req.Header = c.Request.Header
		req.Host = remote.Host
		req.URL.Scheme = remote.Scheme
		req.URL.Host = remote.Host
		req.URL.Path = path.Join(
			"/", "inference", fmt.Sprintf("%s.%s", deployment, ns), c.Param("proxyPath"))

		logrus.WithFields(logrus.Fields{
			"deployment": deployment,
			"uid":        uid,
			"ns":         ns,
			"path":       req.URL.Path,
			"remote":     remote.String(),
		}).Debug("proxying to streamlit")
	}
	proxy.ModifyResponse = func(resp *http.Response) error {
		if resp.StatusCode == http.StatusSeeOther {
			resp.StatusCode = http.StatusOK
			instances, err := s.runtime.InferenceInstanceList(ns, deployment)
			if err != nil {
				return NewError(http.StatusInternalServerError, err, "instance-list")
			}

			buf, err := static.RenderDeploymentLoadingPage("streamlit", resp.Header.Get("X-Call-Id"),
				"We are currently processing your request.", deployment, instances)
			if err != nil {
				return NewError(http.StatusInternalServerError, err, "render-loading-page")
			}
			resp.Body = io.NopCloser(buf)
			resp.ContentLength = int64(buf.Len())
			resp.Header.Set("Content-Length", strconv.Itoa(buf.Len()))
			resp.Header.Set("Content-Type", "text/html")
			resp.StatusCode = http.StatusServiceUnavailable
		}
		return nil
	}

	proxy.ServeHTTP(c.Writer, c.Request)
	return nil
}
