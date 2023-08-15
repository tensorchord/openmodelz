package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tensorchord/openmodelz/agent/api/types"
)

// @Summary     List the servers.
// @Description List the servers.
// @Tags        namespace
// @Param       name    path string           true "Server Name"
// @Param       request body types.ServerSpec true "query params"
// @Accept      json
// @Produce     json
// @Success     200 {object} []string
// @Router      /system/server/{name}/labels [post]
func (s *Server) handleServerLabelCreate(c *gin.Context) error {
	name := c.Param("name")
	if name == "" {
		return NewError(http.StatusBadRequest, errors.New("name is required"),
			"server-label-create")
	}

	var req types.ServerSpec
	if err := c.ShouldBindJSON(&req); err != nil {
		return NewError(http.StatusBadRequest, err, "server-label-create")
	}

	err := s.runtime.ServerLabelCreate(c.Request.Context(), name, req)
	if err != nil {
		return errFromErrDefs(err, "namespace-list")
	}
	c.JSON(http.StatusOK, req)
	return nil
}
