package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tensorchord/openmodelz/agent/api/types"
	_ "github.com/tensorchord/openmodelz/agent/api/types"
)

// @Summary     Attach to the inference instance.
// @Description Attach to the inference instance.
// @Tags        inference
// @Accept      json
// @Produce     json
// @Param       namespace query    string true "Namespace"  example("modelz-d3524a71-c17c-4c92-8faf-8603f02f4713")
// @Param       name      path    string true "Name"
// @Param       instance  path    string true "Instance name"
// @Success     200       {object} []types.InferenceDeployment
// @Router      /system/inference/{name}/instances/{instance} [post]
func (s *Server) handleInferenceInstanceExec(c *gin.Context) error {
	namespace := c.Query("namespace")
	if namespace == "" {
		return NewError(http.StatusBadRequest, errors.New("namespace is required"), "inference-instance-list")
	}
	name := c.Param("name")
	if name == "" {
		return NewError(http.StatusBadRequest, errors.New("name is required"),
			"inference-instance-list")
	}
	instance := c.Param("instance")
	if name == "" {
		return NewError(http.StatusBadRequest, errors.New("name is required"),
			"inference-instance-list")
	}

	var req types.InferenceDeploymentExecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return NewError(http.StatusBadRequest, err, "inference-instance-exec")
	}

	err := s.runtime.InferenceExec(c, namespace, instance, req.Commands)
	if err != nil {
		return errFromErrDefs(err, "inference-instance-exec")
	}

	return nil
}
