package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary     Healthz
// @Description Healthz
// @Tags        system
// @Accept      json
// @Produce     json
// @Success     200
// @Router      /healthz [get]
func (s *Server) handleHealthz(c *gin.Context) error {
	c.Status(http.StatusOK)
	return nil
}
