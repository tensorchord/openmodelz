package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tensorchord/openmodelz/agent/api/types"
)

// @Summary     Delete the namespace.
// @Description Delete the namespace.
// @Tags        namespace
// @Accept      json
// @Produce     json
// @Param       body body     types.NamespaceRequest true "Namespace name"
// @Success     200  {object} types.NamespaceRequest
// @Router      /system/namespaces [delete]
func (s *Server) handleNamespaceDelete(c *gin.Context) error {
	var req types.NamespaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return NewError(http.StatusBadRequest, err, "namespace-delete")
	}

	if err := s.runtime.NamespaceDelete(c.Request.Context(), req.Name); err != nil {
		return errFromErrDefs(err, "namespace-create")
	}

	c.JSON(http.StatusOK, req)
	return nil
}
