package server

import (
	"github.com/gin-gonic/gin"

	_ "github.com/tensorchord/openmodelz/agent/api/types"
)

// @Summary     Get the build logs.
// @Description Get the build logs.
// @Tags        log
// @Accept      json
// @Produce     json
// @Param       namespace query    string true  "Namespace"
// @Param       name      query    string true  "Build Name"
// @Param       instance  query    string false "Instance"
// @Param       tail      query    int    false "Tail"
// @Param       follow    query    bool   false "Follow"
// @Param       since     query    string false "Since"
// @Success     200       {object} []types.Message
// @Router      /system/logs/build [get]
func (s *Server) handleBuildLogs(c *gin.Context) error {
	return s.getLogsFromRequester(c, s.buildLogRequester)
}
