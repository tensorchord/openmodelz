package server

import (
	"fmt"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/tensorchord/openmodelz/agent/pkg/consts"
)

// @Summary     Proxy to the backend mosec.
// @Description Proxy to the backend mosec.
// @Tags        inference
// @Accept      */*
// @Produce     json
// @Param       id path string true "Deployment ID"
// @Router      /mosec/{id} [get]
// @Router      /mosec/{id}/metrics [get]
// @Router      /mosec/{id}/inference [post]
// @Success     201
func (s *Server) proxyMosec(c *gin.Context) error {
	uid, deployment, err := s.proxyAuth(c)
	if err != nil {
		return err
	}

	c.Request.URL.Path = path.Join(
		"/", "inference", fmt.Sprintf("%s.%s", deployment, consts.DefaultPrefix+uid), c.Param("proxyPath"))
	return s.handleInferenceProxy(c)
}
