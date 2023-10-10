package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tensorchord/openmodelz/agent/api/types"
)

// @Summary     List the servers.
// @Description List the servers.
// @Tags        namespace
// @Accept      json
// @Produce     json
// @Success     200 {object} []types.Server
// @Router      /system/servers [get]
func (s *Server) handleServerList(c *gin.Context) error {
	ns := []types.Server{}
	ns, err := s.runtime.ServerList(c.Request.Context())
	if err != nil {
		return errFromErrDefs(err, "namespace-list")
	}
	c.JSON(http.StatusOK, ns)
	return nil
}
