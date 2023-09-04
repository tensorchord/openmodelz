package client

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/tensorchord/openmodelz/agent/api/types"
)

const (
	DefaultPrefix = "modelz-"
)

func ParseAgentToken(token string) (types.AgentToken, error) {
	agentToken := types.AgentToken{}
	if token == "" {
		return agentToken, errors.New("agent token is empty")
	}

	strings := strings.Split(token, ":")
	if len(strings) != 3 {
		return agentToken, errors.New("invalid agent token")
	}
	agentToken.Type = strings[0]
	agentToken.UID = strings[1]
	agentToken.Token = strings[2]

	return agentToken, nil
}

func GetNamespaceByUserID(uid string) string {
	return fmt.Sprintf("%s%s", DefaultPrefix, uid)
}

func GetUserIDFromNamespace(ns string) (string, error) {
	if len(ns) < 8 {
		return "", fmt.Errorf("namespace too short")
	}

	if ns[:len(DefaultPrefix)] != DefaultPrefix {
		return "", fmt.Errorf("namespace does not start with ")
	}

	return ns[7:], nil
}
