package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

// @Summary     Create the inferences.
// @Description Create the inferences.
// @Tags        inference
// @Accept      json
// @Produce     json
// @Param       request body     types.InferenceDeployment true "query params"
// @Success     201     {object} types.InferenceDeployment
// @Router      /system/inferences [post]
func (s *Server) handleInferenceCreate(c *gin.Context) error {
	event := types.DeploymentCreateEvent

	var req types.InferenceDeployment
	if err := c.ShouldBindJSON(&req); err != nil {
		return NewError(http.StatusBadRequest, err, event)
	}

	// Set the default values.
	s.validator.DefaultDeployRequest(&req)

	// Validate the request.
	if err := s.validator.ValidateDeployRequest(&req); err != nil {
		return NewError(http.StatusBadRequest, err, event)
	}

	// Create the inference.
	if err := s.runtime.InferenceCreate(c.Request.Context(), req); err != nil {
		return errFromErrDefs(err, event)
	}
	c.JSON(http.StatusCreated, req)
	return nil
}
