package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/tensorchord/openmodelz/agent/errdefs"
	"github.com/tensorchord/openmodelz/agent/pkg/consts"
)

func (s *Server) proxyAuth(c *gin.Context) (string, string, error) {
	var uid string
	var valid bool

	deployment := c.Param("id")
	if len(deployment) == 0 {
		return "", "", errFromErrDefs(
			fmt.Errorf("cannot find the deployment name in %s", c.Request.RequestURI), "get-deployment")
	}

	key := c.GetHeader("X-API-Key")
	// Be compatible with the OpenAI API.
	rawKeyStr := c.GetHeader("Authorization")
	logrus.Debug("proxyOther: key: ", key, ", rawKeyStr: ", rawKeyStr)

	if s.validateUnifiedKey(key) {
		// uid 0 means to use unified api key
		uid = "00000000-0000-0000-0000-000000000000"
	} else if len(key) > 0 {
		uid, valid = s.validateAPIKey(key)
		if !valid {
			return "", "", errdefs.Unauthorized(fmt.Errorf("invalid API key"))
		}
	} else if len(rawKeyStr) > 0 {
		strs := strings.Split(rawKeyStr, " ")
		if len(strs) != 2 {
			return "", "", errdefs.Unauthorized(fmt.Errorf("invalid Authorization API key"))
		}

		if strs[0] != "Bearer" {
			return "", "", errdefs.Unauthorized(fmt.Errorf("invalid Authorization API key"))
		}

		uid, valid = s.validateAPIKey(strs[1])
		if !valid {
			return "", "", errdefs.Unauthorized(fmt.Errorf("invalid Authorization API key"))
		}
	}

	if len(uid) == 0 {
		return "", "", errdefs.Unauthorized(fmt.Errorf("invalid API key"))
	}
	return uid, deployment, nil
}

func (s *Server) proxyNoAuth(c *gin.Context) (string, string, error) {
	deployment := c.Param("id")
	if len(deployment) == 0 {
		return "", "", errdefs.InvalidParameter(
			fmt.Errorf("cannot find the deployment name in %s", c.Request.RequestURI))
	}

	uid, found := s.getUIDFromDeploymentID(c.Request.Context(), deployment)
	if !found {
		return "", "", errdefs.InvalidParameter(
			fmt.Errorf("cannot find the user id from the deployment id"))
	}
	return uid, deployment, nil
}

func (s *Server) validateAPIKey(key string) (string, bool) {
	if !strings.HasPrefix(key, consts.APIKEY_PREFIX) {
		return "", false
	}

	apikeys := s.config.ModelZCloud.APIKeys
	uid, exit := apikeys[key]
	if exit {
		return uid, true
	}

	apiServerReady := make(chan struct{})
	go func() {
		if err := s.modelzCloudClient.WaitForAPIServerReady(); err != nil {
			logrus.Fatalf("failed to wait for apiserver ready: %v", err)
		}
		close(apiServerReady)
	}()
	// Get from apiserver
	apikeys, err := s.modelzCloudClient.GetAPIKeys(context.Background(), apiServerReady, s.config.ModelZCloud.AgentToken, s.config.ModelZCloud.ID)
	if err != nil {
		logrus.Errorf("failed to get apikeys: %v", err)
		return "", false
	}
	uid, exit = apikeys[key]
	if exit {
		return uid, true
	}

	return "", false
}

func (s *Server) validateUnifiedKey(key string) bool {
	if !strings.HasPrefix(key, consts.APIKEY_PREFIX) {
		return false
	}
	if len(s.config.ModelZCloud.UnifiedAPIKey) != 0 && s.config.ModelZCloud.UnifiedAPIKey == key {
		return true
	}
	return false
}
