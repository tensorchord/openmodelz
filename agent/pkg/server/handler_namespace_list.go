package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary     List the namespaces.
// @Description List the namespaces.
// @Tags        namespace
// @Accept      json
// @Produce     json
// @Success     200 {object} []string
// @Router      /system/namespaces [get]
func (s *Server) handleNamespaceList(c *gin.Context) error {
	ns, err := s.runtime.NamespaceList(c.Request.Context())
	if err != nil {
		return errFromErrDefs(err, "namespace-list")
	}
	c.JSON(http.StatusOK, ns)
	return nil
}
