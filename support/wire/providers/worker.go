package providers

import (
	"go-reasonable-api/app/interfaces/repositories"
	"go-reasonable-api/app/interfaces/support"
	"go-reasonable-api/app/tasks"
	"go-reasonable-api/support/config"
	"go-reasonable-api/support/worker"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
)

func ProvideEmailTask(logger *zerolog.Logger, emailSender support.EmailSender) (*tasks.EmailTask, error) {
	return tasks.NewEmailTask(logger, emailSender)
}

func ProvideCleanupTask(
	logger *zerolog.Logger,
	authTokenRepo repositories.AuthTokenRepository,
	passwordResetRepo repositories.PasswordResetRepository,
	emailVerificationRepo repositories.EmailVerificationRepository,
	userRepo repositories.UserRepository,
) *tasks.CleanupTask {
	return tasks.NewCleanupTask(logger, authTokenRepo, passwordResetRepo, emailVerificationRepo, userRepo)
}

func ProvideTaskRegistry(emailTask *tasks.EmailTask, cleanupTask *tasks.CleanupTask) *tasks.Registry {
	return tasks.NewRegistry(emailTask, cleanupTask)
}

func ProvideServeMux(registry *tasks.Registry) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	registry.RegisterHandlers(mux)
	return mux
}

func ProvideScheduler(cfg *config.Config) *asynq.Scheduler {
	return asynq.NewScheduler(
		asynq.RedisClientOpt{Addr: cfg.Redis.Addr},
		&asynq.SchedulerOpts{},
	)
}

func ProvideWorker(server *asynq.Server, mux *asynq.ServeMux, scheduler *asynq.Scheduler, registry *tasks.Registry, logger *zerolog.Logger) *worker.Worker {
	return worker.NewWorker(server, mux, scheduler, registry, logger)
}
