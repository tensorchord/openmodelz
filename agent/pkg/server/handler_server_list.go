package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary     List the servers.
// @Description List the servers.
// @Tags        namespace
// @Accept      json
// @Produce     json
// @Success     200 {object} []string
// @Router      /system/servers [get]
func (s *Server) handleServerList(c *gin.Context) error {
	ns, err := s.runtime.ServerList(c.Request.Context())
	if err != nil {
		return errFromErrDefs(err, "namespace-list")
	}
	c.JSON(http.StatusOK, ns)
	return nil
}
