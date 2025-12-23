package tasks

import (
	"context"

	"github.com/hibiken/asynq"
)

type TaskHandler interface {
	HandlePasswordResetEmail(ctx context.Context, t *asynq.Task) error
}
