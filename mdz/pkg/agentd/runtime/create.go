package runtime

import (
	"context"
	"fmt"
	"os"
	"time"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/moby/moby/pkg/jsonmessage"
	"github.com/moby/term"
	"github.com/phayes/freeport"
	"github.com/tensorchord/openmodelz/agent/api/types"
)

func (r *Runtime) InferenceCreate(ctx context.Context, req types.InferenceDeployment) error {
	cfg := &container.Config{
		Image:        req.Spec.Image,
		ExposedPorts: nat.PortSet{},
	}

	var port int32 = 8080
	if req.Spec.Port != nil {
		port = *req.Spec.Port
	}

	now := time.Now()
	req.Status = types.InferenceDeploymentStatus{
		Phase:     types.PhaseNotReady,
		Replicas:  1,
		CreatedAt: &now,
	}
	// Lock the mutex and set cache
	r.mutex.Lock()
	r.cache[req.Spec.Name] = req
	r.mutex.Unlock()

	go func() error {
		body, err := r.cli.ImagePull(context.TODO(), req.Spec.Image, dockertypes.ImagePullOptions{})
		if err != nil {
			return err
		}
		defer body.Close()

		termFd, isTerm := term.GetFdInfo(os.Stdout)
		err = jsonmessage.DisplayJSONMessagesStream(body, os.Stdout, termFd, isTerm, nil)
		if err != nil {
			return err
		}

		hostCfg := &container.HostConfig{
			RestartPolicy: container.RestartPolicy{
				Name: "always",
			},
			PortBindings: nat.PortMap{},
		}

		natPort := nat.Port(fmt.Sprintf("%d/tcp", port))

		hostPort, err := freeport.GetFreePort()
		if err != nil {
			return err
		}
		hostCfg.PortBindings[natPort] = []nat.PortBinding{
			{
				HostIP:   Localhost,
				HostPort: fmt.Sprintf("%d", hostPort),
			},
		}
		cfg.ExposedPorts[natPort] = struct{}{}

		cfg.Labels = expectedLabels(req)

		ctr, err := r.cli.ContainerCreate(context.TODO(), cfg, hostCfg, nil, nil, req.Spec.Name)
		if err != nil {
			return err
		}

		if err := r.cli.ContainerStart(context.TODO(), ctr.ID, dockertypes.ContainerStartOptions{}); err != nil {
			return err
		}

		r.mutex.Lock()
		new := r.cache[req.Spec.Name]
		new.Status = types.InferenceDeploymentStatus{
			Phase:             types.PhaseReady,
			Replicas:          1,
			AvailableReplicas: 1,
			CreatedAt:         &now,
		}
		r.mutex.Unlock()
		return nil
	}()

	return nil
}
