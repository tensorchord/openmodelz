package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// @Summary     Inference.
// @Description Inference proxy.
// @Tags        inference-proxy
// @Accept      json
// @Produce     json
// @Router      /inference/{name} [post]
// @Router      /inference/{name} [get]
// @Router      /inference/{name} [put]
// @Router      /inference/{name} [delete]
func (s *Server) handleInferenceProxy(c *gin.Context) error {
	namespacedName := c.Param("name")
	if namespacedName == "" {
		return NewError(
			http.StatusBadRequest, errors.New("name is required"), "inference-proxy")
	}

	_, name, err := getNamespaceAndName(namespacedName)
	if err != nil {
		return NewError(
			http.StatusBadRequest, err, "inference-proxy")
	}

	return s.runtime.InfereceProxy(c, name)
}

func getNamespaceAndName(name string) (string, string, error) {
	if !strings.Contains(name, ".") {
		return "", "", fmt.Errorf("name is not namespaced")
	}
	namespace := name[strings.LastIndexAny(name, ".")+1:]
	infName := strings.TrimSuffix(name, "."+namespace)

	if namespace == "" {
		return "", "", fmt.Errorf("namespace is empty")
	}

	if infName == "" {
		return "", "", fmt.Errorf("inference name is empty")
	}
	return namespace, infName, nil
}
