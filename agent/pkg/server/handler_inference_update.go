package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

// @Summary     Update the inferences.
// @Description Update the inferences.
// @Tags        inference
// @Accept      json
// @Produce     json
// @Param       request   body     types.InferenceDeployment true "query params"
// @Param       namespace query    string                    true "Namespace"
// @Success     202       {object} types.InferenceDeployment
// @Router      /system/inferences [put]
func (s *Server) handleInferenceUpdate(c *gin.Context) error {
	event := types.DeploymentUpdateEvent
	var req types.InferenceDeployment
	if err := c.ShouldBindJSON(&req); err != nil {
		return NewError(http.StatusBadRequest, err, event)
	}

	namespace := c.Query("namespace")
	if namespace == "" {
		return NewError(
			http.StatusBadRequest,
			errors.New("namespace is required"), event)
	}

	if err := s.validator.ValidateDeployRequest(&req); err != nil {
		return NewError(http.StatusBadRequest, err, event)
	}

	if err := s.runtime.InferenceUpdate(c.Request.Context(),
		namespace, req, event); err != nil {
		return errFromErrDefs(err, event)
	}

	c.JSON(http.StatusAccepted, req)
	return nil
}
