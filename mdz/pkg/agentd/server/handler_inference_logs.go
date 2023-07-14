package server

import (
	"github.com/gin-gonic/gin"
)

// @Summary     Get the inference logs.
// @Description Get the inference logs.
// @Tags        log
// @Accept      json
// @Produce     json
// @Param       namespace query    string true  "Namespace" example("modelz-d3524a71-c17c-4c92-8faf-8603f02f4713")
// @Param       name      query    string true  "Name"
// @Param       instance  query    string false "Instance"
// @Param       tail      query    int    false "Tail"
// @Param       follow    query    bool   false "Follow"
// @Param       since     query    string false "Since" example("2023-04-01T00:06:31+08:00")
// @Param       end       query    string false "End"   example("2023-05-31T00:06:31+08:00")
// @Success     200       {object} []types.Message
// @Router      /system/logs/inference [get]
func (s *Server) handleInferenceLogs(c *gin.Context) error {
	return nil
}
