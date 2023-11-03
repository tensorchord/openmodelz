package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tensorchord/openmodelz/agent/api/types"
)

// @Summary     Create the secret.
// @Description Create the secret.
// @Tags        secret
// @Accept      json
// @Produce     json
// @Param       body body     types.Secret true "Secret"
// @Success     200  {object} types.Secret
// @Router      /system/secrets [post]
func (s *Server) handleSecretCreate(c *gin.Context) error {
	var req types.Secret
	if err := c.ShouldBindJSON(&req); err != nil {
		return NewError(http.StatusBadRequest, err, "failed to parse request body")
	}

	if err := s.runtime.CreateSecret(c.Request.Context(), &req); err != nil {
		return NewError(http.StatusInternalServerError, err, "failed to create secret")
	}

	c.JSON(http.StatusOK, req)
	return nil
}
