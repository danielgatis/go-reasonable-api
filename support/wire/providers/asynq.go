package providers

import (
	"context"
	"encoding/json"

	"go-reasonable-api/app/interfaces/support"
	"go-reasonable-api/support/config"
	"go-reasonable-api/support/sentry"
	"go-reasonable-api/support/taskqueue"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
)

func ProvideAsynqClient(cfg *config.Config) (*asynq.Client, func(), error) {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: cfg.Redis.Addr})

	cleanup := func() {
		_ = client.Close()
	}

	return client, cleanup, nil
}

func ProvideTaskClient(asynqClient *asynq.Client) support.TaskClient {
	return taskqueue.NewClient(asynqClient)
}

func ProvideAsynqServer(cfg *config.Config, logger *zerolog.Logger) *asynq.Server {
	return asynq.NewServer(
		asynq.RedisClientOpt{Addr: cfg.Redis.Addr},
		asynq.Config{
			Concurrency: cfg.Worker.Concurrency,
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				var payload any
				if jsonErr := json.Unmarshal(task.Payload(), &payload); jsonErr != nil {
					payload = string(task.Payload())
				}
				sentry.CaptureError(err, map[string]any{
					"task_type":    task.Type(),
					"task_payload": payload,
				})
			}),
		},
	)
}
