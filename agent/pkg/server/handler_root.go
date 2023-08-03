package server

import (
	"github.com/gin-gonic/gin"
	"github.com/tensorchord/openmodelz/agent/pkg/server/static"
)

func (s *Server) handleRoot(c *gin.Context) error {
	lp, err := static.RenderLoadingPage()
	if err != nil {
		return err
	}

	c.Data(200, "text/html; charset=utf-8", lp.Bytes())
	return nil
}
