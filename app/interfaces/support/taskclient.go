package support

import (
	"context"

	"github.com/hibiken/asynq"
)

// TaskClient abstracts async task enqueueing.
//
// EnqueueCtx is fire-and-forget: it logs errors but doesn't return them.
// This design ensures email sending failures don't break user flows.
// The implementation extracts request metadata from context for tracing.
type TaskClient interface {
	EnqueueCtx(ctx context.Context, taskType string, payload any, opts ...asynq.Option)
}
