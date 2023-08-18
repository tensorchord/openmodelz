package server

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	_ "github.com/tensorchord/openmodelz/agent/api/types"
)

// @Summary     Attach to the inference instance.
// @Description Attach to the inference instance.
// @Tags        inference
// @Accept      json
// @Produce     json
// @Param       namespace query    string true "Namespace"
// @Param       name      path     string true "Name"
// @Param       instance  path     string true "Instance name"
// @Success     200       {object} []types.InferenceDeployment
// @Router      /system/inference/{name}/instance/{instance} [post]
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
		return NewError(http.StatusBadRequest, errors.New("instance is required"),
			"inference-instance-list")
	}

	tty := c.Query("tty")
	if tty == "" {
		tty = "false"
	}
	ttyBoolean, err := strconv.ParseBool(tty)
	if err != nil {
		return NewError(http.StatusBadRequest, err, "inference-instance-exec")
	}

	command := c.Query("command")
	commandSlice := strings.Split(command, ",")

	if err := s.runtime.InferenceExec(
		c, namespace, instance, commandSlice, ttyBoolean); err != nil {
		return errFromErrDefs(err, "inference-instance-exec")
	}

	return nil
}
