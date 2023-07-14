package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/tensorchord/openmodelz/agent/api/types"
)

// @Summary     Get the inference by name.
// @Description Get the inference by name.
// @Tags        inference
// @Accept      json
// @Produce     json
// @Param       namespace query    string true "Namespace"  example("modelz-d3524a71-c17c-4c92-8faf-8603f02f4713")
// @Param       name      path     string true "inference id" example("e50886f3-caa6-449f-9fa8-7849c6ba2e08")
// @Success     200       {object} types.InferenceDeployment
// @Router      /system/inference/{name} [get]
func (s *Server) handleInferenceGet(c *gin.Context) error {
	namespace := c.Query("namespace")
	if namespace == "" {
		return NewError(
			http.StatusBadRequest, errors.New("namespace is required"), "inference-get")
	}
	name := c.Param("name")
	if name == "" {
		return NewError(
			http.StatusBadRequest, errors.New("name is required"), "inference-get")
	}

	// function, err := s.runtime.InferenceGet(namespace, name)
	// if err != nil {
	// 	return errFromErrDefs(err, "inference-get")
	// }

	// c.JSON(http.StatusOK, function)
	return nil
}
