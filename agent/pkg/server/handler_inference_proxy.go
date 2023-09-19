package server

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/tensorchord/openmodelz/agent/errdefs"
)

// @Summary     Inference.
// @Description Inference proxy.
// @Tags        inference-proxy
// @Accept      json
// @Produce     json
// @Param       name path string true "inference id"
// @Router      /inference/{name} [post]
// @Router      /inference/{name} [get]
// @Router      /inference/{name} [put]
// @Router      /inference/{name} [delete]
// @Success     200
// @Failure     303
// @Failure     400
// @Failure     404
// @Failure     500
func (s *Server) handleInferenceProxy(c *gin.Context) error {
	namespacedName := c.Param("name")
	if namespacedName == "" {
		return NewError(
			http.StatusBadRequest, errors.New("name is required"), "inference-proxy")
	}

	namespace, name, err := getNamespaceAndName(namespacedName)
	if err != nil {
		return NewError(
			http.StatusBadRequest, err, "inference-proxy")
	}

	// Update metrics.
	s.metricsOptions.GatewayInferenceInvocationStarted.
		WithLabelValues(namespacedName).Inc()
	s.metricsOptions.GatewayInferenceInvocationInflight.
		WithLabelValues(namespacedName).Inc()
	start := time.Now()
	label := prometheus.Labels{"inference_name": namespacedName, "code": strconv.Itoa(http.StatusProcessing)}
	defer func() {
		s.metricsOptions.GatewayInferenceInvocationInflight.
			WithLabelValues(namespacedName).Dec()
		s.metricsOptions.GatewayInferencesHistogram.With(label).
			Observe(time.Since(start).Seconds())
		s.metricsOptions.GatewayInferenceInvocation.With(label).Inc()
	}()

	res := s.scaler.Scale(c.Request.Context(), namespace, name)
	if !res.Found {
		label["code"] = strconv.Itoa(http.StatusNotFound)
		return NewError(
			http.StatusNotFound, errors.New("inference not found"), "inference-proxy")
	} else if res.Error != nil {
		label["code"] = strconv.Itoa(http.StatusInternalServerError)
		return NewError(
			http.StatusInternalServerError, res.Error, "inference-proxy")
	}

	if res.Available {
		statusCode, err := s.forward(c, namespace, name)
		if err != nil {
			label["code"] = strconv.Itoa(statusCode)
			return NewError(statusCode, err, "inference-proxy")
		}
		label["code"] = strconv.Itoa(statusCode)
		return nil
	} else {
		// The inference is still being created.
		label["code"] = strconv.Itoa(http.StatusSeeOther)
		return NewError(http.StatusSeeOther,
			fmt.Errorf("inference %s is not available", name), "inference-proxy")
	}
}

func (s *Server) forward(c *gin.Context, namespace, name string) (int, error) {
	backendURL, err := s.endpointResolver.Resolve(namespace, name)
	if err != nil {
		return 0, errdefs.InvalidParameter(err)
	}
	defer s.endpointResolver.Close(backendURL)

	proxyServer := httputil.ReverseProxy{}
	proxyServer.Transport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   s.config.Server.ReadTimeout,
			KeepAlive: s.config.Server.ReadTimeout,
			DualStack: true,
		}).DialContext,
	}
	proxyServer.Director = func(req *http.Request) {
		targetQuery := backendURL.RawQuery
		req.URL.Scheme = backendURL.Scheme
		req.URL.Host = backendURL.Host
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		req.URL.Path = c.Param("proxyPath")
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}

		s.logger.WithField("url", backendURL.String()).
			WithField("path", req.URL.Path).
			WithField("header", req.Header).
			WithField("raw-query", req.URL.RawQuery).Debug("reverse proxy")
	}

	var statusCode int
	proxyServer.ModifyResponse = func(resp *http.Response) error {
		statusCode = resp.StatusCode
		return nil
	}

	proxyServer.ServeHTTP(c.Writer, c.Request)
	return statusCode, nil
}

func getNamespaceAndName(name string) (string, string, error) {
	if !strings.Contains(name, ".") {
		return "", "", fmt.Errorf("name is not namespaced")
	}
	namespace := name[strings.LastIndexAny(name, ".")+1:]
	infName := strings.TrimSuffix(name, "."+namespace)

	if namespace == "" {
		return "", "", fmt.Errorf("namespace is empty")
	}

	if infName == "" {
		return "", "", fmt.Errorf("inference name is empty")
	}
	return namespace, infName, nil
}
