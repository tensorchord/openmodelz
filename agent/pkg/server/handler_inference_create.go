package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/client"
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

	if s.config.ModelZCloud.Enabled {
		ns := req.Spec.Namespace
		user, err := client.GetUserIDFromNamespace(ns)
		if err != nil {
			return err
		} else if user == "" {
			return fmt.Errorf("user id is empty")
		}
		s.cache.SetWithTTL(req.Spec.Name, user, 1, 0)

		exist := s.runtime.NamespaceGet(c.Request.Context(), ns)
		if !exist {
			if err := s.runtime.NamespaceCreate(c.Request.Context(), ns); err != nil {
				return err
			}
		}
	}

	// Set the default values.
	s.validator.DefaultDeployRequest(&req)

	// Validate the request.
	if err := s.validator.ValidateDeployRequest(&req); err != nil {
		return NewError(http.StatusBadRequest, err, event)
	}

	// Create the inference.
	if err := s.runtime.InferenceCreate(c.Request.Context(), req,
		s.config.Ingress, event, s.config.Server.ServerPort); err != nil {
		return errFromErrDefs(err, event)
	}
	c.JSON(http.StatusCreated, req)
	return nil
}
