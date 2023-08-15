package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary     Delete a node from the cluster.
// @Description Delete a node.
// @Tags        namespace
// @Param       name path string true "Server Name"
// @Accept      json
// @Produce     json
// @Success     200
// @Router      /system/server/{name}/delete [delete]
func (s *Server) handleServerDelete(c *gin.Context) error {
	name := c.Param("name")
	if name == "" {
		return NewError(http.StatusBadRequest, errors.New("name is required"), "server-delete-node")
	}
	err := s.runtime.ServerDeleteNode(c.Request.Context(), name)
	if err != nil {
		return errFromErrDefs(err, "server-delete-node")
	}
	return nil
}
