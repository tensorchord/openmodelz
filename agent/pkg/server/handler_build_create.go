package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

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

	inference, err := s.runtime.InferenceGetCRD(req.Spec.Namespace, req.Spec.Name)
	if err != nil {
		return errFromErrDefs(err, "inference-instance-list")
	}

	if err := s.runtime.BuildCreate(c.Request.Context(), req, inference,
		s.config.Build.BuilderImage, s.config.Build.BuildkitdAddress,
		s.config.Build.BuildCtlBin, s.config.Build.BuildImagePullSecret); err != nil {
		logrus.Errorf("failed to create build: %v", err)
		return errFromErrDefs(err, "build-create")
	}
	c.JSON(http.StatusOK, req)
	return nil
}
