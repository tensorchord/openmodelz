package server

import (
	"context"

	"github.com/sirupsen/logrus"
)

func (s *Server) getUIDFromDeploymentID(ctx context.Context, id string) (string, bool) {

	uid, exit := s.cache.Get(id)
	if exit {
		return uid.(string), true
	}

	uid, err := s.modelzCloudClient.GetUIDFromDeploymentID(ctx, s.config.ModelZCloud.AgentToken, s.config.ModelZCloud.ID, id)
	if err != nil {
		logrus.Errorf("failed to get uid from deployment id: %v", err)
		return "", false
	}

	// no expiration
	s.cache.SetWithTTL(id, uid, 1, 0)
	return uid.(string), true
}
