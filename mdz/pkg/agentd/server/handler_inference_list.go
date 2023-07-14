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
// @Param       namespace query    string true "Namespace"  example("modelz-d3524a71-c17c-4c92-8faf-8603f02f4713")
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

	c.JSON(http.StatusOK, inferenes)
	return nil
}
