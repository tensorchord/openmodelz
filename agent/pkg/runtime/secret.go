package runtime

import (
	"context"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

func (r generalRuntime) CreateSecret(ctx context.Context, secret *types.Secret) error {
	err := r.secretClient.Create(*secret)
	if err != nil {
		return err
	}
	return nil
}
