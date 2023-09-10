package event

import "fmt"

const (
	DefaultPrefix = "modelz-"
)

func getUserIDFromNamespace(ns string) (string, error) {
	if len(ns) < 8 {
		return "", fmt.Errorf("namespace too short")
	}

	if ns[:len(DefaultPrefix)] != DefaultPrefix {
		return "", fmt.Errorf("namespace does not start with %s", DefaultPrefix)
	}

	return ns[len(DefaultPrefix):], nil
}
