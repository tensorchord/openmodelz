package server

import (
	"github.com/tensorchord/openmodelz/agent/pkg/log"
)

func (s *Server) initLogs() {
	if len(s.config.Logs.LokiURL) > 0 {
		s.logger.Info("enable Loki logs requester")
		s.buildLogRequester = log.NewLokiAPIRequestor(
			s.config.Logs.LokiURL, s.config.Logs.LokiUser, s.config.Logs.LokiToken)
	}

}
