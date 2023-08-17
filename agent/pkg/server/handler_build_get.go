package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/tensorchord/openmodelz/agent/api/types"
)

// @Summary     Get the build by name.
// @Description Get the build by name.
// @Tags        build
// @Accept      json
// @Produce     json
// @Param       namespace query    string true "Namespace"
// @Param       name      path     string true "inference id"
// @Success     200       {object} types.Build
// @Router      /system/build/{name} [get]
func (s *Server) handleBuildGet(c *gin.Context) error {
	namespace := c.Query("namespace")
	if namespace == "" {
		return NewError(
			http.StatusBadRequest, errors.New("namespace is required"), "inference-list")
	}
	name := c.Param("name")
	if name == "" {
		return NewError(
			http.StatusBadRequest, errors.New("name is required"), "build-get")
	}

	build, err := s.runtime.BuildGet(c.Request.Context(), namespace, name)
	if err != nil {
		return errFromErrDefs(err, "build-get")
	}

	c.JSON(http.StatusOK, build)
	return nil
}
