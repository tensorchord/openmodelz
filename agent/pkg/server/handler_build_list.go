package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/tensorchord/openmodelz/agent/api/types"
)

// @Summary     List the builds.
// @Description List the builds.
// @Tags        build
// @Accept      json
// @Produce     json
// @Param       namespace query    string true "Namespace"
// @Success     200       {object} []types.Build
// @Router      /system/build [get]
func (s *Server) handleBuildList(c *gin.Context) error {
	namespace := c.Query("namespace")
	if namespace == "" {
		return NewError(
			http.StatusBadRequest, errors.New("namespace is required"), "inference-list")
	}

	builds, err := s.runtime.BuildList(c.Request.Context(), namespace)
	if err != nil {
		return errFromErrDefs(err, "build-list")
	}
	c.JSON(http.StatusOK, builds)
	return nil
}
