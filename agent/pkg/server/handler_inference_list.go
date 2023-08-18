package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/tensorchord/openmodelz/agent/api/types"
)

// @Summary     List the inferences.
// @Description List the inferences.
// @Tags        inference
// @Accept      json
// @Produce     json
// @Param       namespace query    string true "Namespace"
// @Success     200       {object} []types.InferenceDeployment
// @Router      /system/inferences [get]
func (s *Server) handleInferenceList(c *gin.Context) error {
	namespace := c.Query("namespace")
	if namespace == "" {
		return NewError(
			http.StatusBadRequest, errors.New("namespace is required"), "inference-list")
	}

	inferenes, err := s.runtime.InferenceList(namespace)
	if err != nil {
		return errFromErrDefs(err, "inference-list")
	}
	// Add invocation count metrics into the body.
	// TODO: https://github.com/tensorchord/openmodelz/issues/203
	s.prometheusClient.AddMetrics(inferenes)

	c.JSON(http.StatusOK, inferenes)
	return nil
}
