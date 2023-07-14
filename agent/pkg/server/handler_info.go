package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/tensorchord/openmodelz/agent/api/types"
	"github.com/tensorchord/openmodelz/agent/pkg/version"
)

// @Summary     Get system info.
// @Description Get system info.
// @Tags        system
// @Accept      json
// @Produce     json
// @Success     200 {object} types.ProviderInfo
// @Router      /system/info [get]
func (s *Server) handleInfo(c *gin.Context) error {
	v := version.GetVersion()
	c.JSON(http.StatusOK, types.ProviderInfo{
		Name:          "agent",
		Orchestration: "kubernetes",
		Version: &types.VersionInfo{
			Version:      v.Version,
			BuildDate:    v.BuildDate,
			GitCommit:    v.GitCommit,
			GitTag:       v.GitTag,
			GitTreeState: v.GitTreeState,
			GoVersion:    v.GoVersion,
			Compiler:     v.Compiler,
			Platform:     v.Platform,
		},
	})
	return nil
}
