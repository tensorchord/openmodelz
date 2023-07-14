package log

import (
	"context"

	"github.com/tensorchord/openmodelz/agent/api/types"
)

// Requester submits queries the logging system.
type Requester interface {
	// Query submits a log request to the actual logging system.
	Query(ctx context.Context, req types.LogRequest) (<-chan types.Message, error)
}
