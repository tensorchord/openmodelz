package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

// @Summary     Delete the inferences.
// @Description Delete the inferences.
// @Tags        inference
// @Accept      json
// @Produce     json
// @Param       request   body     types.DeleteFunctionRequest true "query params"
// @Param       namespace query    string                      true "Namespace"  example("modelz-d3524a71-c17c-4c92-8faf-8603f02f4713")
// @Success     202       {object} types.DeleteFunctionRequest
// @Router      /system/inferences [delete]
func (s *Server) handleInferenceDelete(c *gin.Context) error {
	event := types.DeploymentDeleteEvent
	var req types.DeleteFunctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return NewError(http.StatusBadRequest, err, event)
	}

	if req.FunctionName == "" {
		return NewError(
			http.StatusBadRequest,
			errors.New("function name is required"), event)
	}

	if err := s.runtime.InferenceDelete(c.Request.Context(),
		req.FunctionName); err != nil {
		return errFromErrDefs(err, event)
	}

	c.JSON(http.StatusAccepted, req)
	return nil
}
