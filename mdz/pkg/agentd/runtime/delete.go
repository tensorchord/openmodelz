package runtime

import (
	"context"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/tensorchord/openmodelz/agent/client"
)

func (r *Runtime) InferenceDelete(ctx context.Context, name string) error {
	defer func() {
		r.mutex.Lock()
		delete(r.cache, name)
		r.mutex.Unlock()
	}()

	ctr, err := r.cli.ContainerInspect(ctx, name)
	if err != nil {
		if !client.IsErrNotFound(err) {
			return nil
		}
	}

	if ctr.Config.Labels[labelVendor] != valueVendor {
		return nil
	}

	if err := r.cli.ContainerRemove(ctx, name, dockertypes.ContainerRemoveOptions{
		Force: true,
	}); err != nil {
		return err
	}

	return nil
}
