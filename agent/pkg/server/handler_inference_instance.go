package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/tensorchord/openmodelz/agent/api/types"
)

// @Summary     List the inference instances.
// @Description List the inference instances.
// @Tags        inference
// @Accept      json
// @Produce     json
// @Param       namespace query    string true "Namespace"
// @Param       name      path     string true "Name"
// @Success     200       {object} []types.InferenceDeployment
// @Router      /system/inference/{name}/instances [get]
func (s *Server) handleInferenceInstance(c *gin.Context) error {
	namespace := c.Query("namespace")
	if namespace == "" {
		return NewError(http.StatusBadRequest, errors.New("namespace is required"), "inference-instance-list")
	}
	name := c.Param("name")
	if name == "" {
		return NewError(http.StatusBadRequest, errors.New("name is required"),
			"inference-instance-list")
	}

	instances, err := s.runtime.InferenceInstanceList(namespace, name)
	if err != nil {
		return errFromErrDefs(err, "inference-instance-list")
	}
	c.JSON(http.StatusOK, instances)
	return nil
}
