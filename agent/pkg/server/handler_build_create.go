package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

// @Summary     Create the build.
// @Description Create the build.
// @Tags        build
// @Accept      json
// @Produce     json
// @Param       body body     types.Build true "build"
// @Success     200  {object} types.Build
// @Router      /system/build [post]
func (s *Server) handleBuildCreate(c *gin.Context) error {
	var req types.Build
	if err := c.ShouldBindJSON(&req); err != nil {
		return NewError(
			http.StatusBadRequest, err, "build-create")
	}

	if err := s.validator.ValidateBuildRequest(&req); err != nil {
		return NewError(
			http.StatusBadRequest, err, "build-create")
	}
	s.validator.DefaultBuildRequest(&req)

	if err := s.runtime.BuildCreate(c.Request.Context(), req,
		s.config.Build.BuilderImage, s.config.Build.BuildkitdAddress,
		s.config.Build.BuildCtlBin, s.config.Build.BuildRegistry,
		s.config.Build.BuildRegistryToken,
	); err != nil {
		return errFromErrDefs(err, "build-create")
	}
	c.JSON(http.StatusOK, req)
	return nil
}
