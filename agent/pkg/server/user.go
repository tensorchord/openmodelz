package server

import (
	"context"
)

func (s *Server) getUIDFromDeploymentID(ctx context.Context, id string) (string, bool) {

	uid, exit := s.cache.Get(id)
	if exit {
		return uid.(string), true
	}

	uid, err := s.modelzCloudClient.GetUIDFromDeploymentID(ctx, s.config.ModelZCloud.TokenID, s.config.ModelZCloud.ID, id)
	if err != nil {
		return "", false
	}

	// no expiration
	s.cache.SetWithTTL(id, uid, 1, 0)
	return uid.(string), true
}
