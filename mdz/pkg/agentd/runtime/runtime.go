package runtime

import (
	"context"
	"sync"

	"github.com/docker/docker/client"
	"github.com/tensorchord/openmodelz/agent/api/types"
)

type Runtime struct {
	cli   client.APIClient
	cache map[string]types.InferenceDeployment
	mutex sync.Mutex
}

const (
	Localhost = "127.0.0.1"
)

func New() (*Runtime, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	cli.NegotiateAPIVersion(context.Background())

	return &Runtime{
		cli:   cli,
		cache: map[string]types.InferenceDeployment{},
	}, nil
}
