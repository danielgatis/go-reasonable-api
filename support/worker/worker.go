package worker

import (
	"go-reasonable-api/app/tasks"

	"github.com/hibiken/asynq"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
)

// Worker encapsulates the asynq server, task handlers, and scheduler
type Worker struct {
	server    *asynq.Server
	mux       *asynq.ServeMux
	scheduler *asynq.Scheduler
	registry  *tasks.Registry
	logger    *zerolog.Logger
}

// NewWorker creates a new Worker instance
func NewWorker(server *asynq.Server, mux *asynq.ServeMux, scheduler *asynq.Scheduler, registry *tasks.Registry, logger *zerolog.Logger) *Worker {
	return &Worker{
		server:    server,
		mux:       mux,
		scheduler: scheduler,
		registry:  registry,
		logger:    logger,
	}
}

// Run starts the worker and scheduler, blocking until shutdown
func (w *Worker) Run() error {
	w.logger.Info().Msg("starting worker")

	// Register scheduled tasks from the registry
	if err := w.registry.RegisterScheduledTasks(w.scheduler); err != nil {
		w.logger.Error().Err(err).Msg("failed to register scheduled tasks")
		return eris.Wrap(err, "failed to register scheduled tasks")
	}
	w.logger.Info().Msg("registered scheduled tasks")

	// Start the scheduler in a goroutine
	go func() {
		if err := w.scheduler.Run(); err != nil {
			w.logger.Error().Err(err).Msg("scheduler error")
			panic(eris.Wrap(err, "scheduler failed"))
		}
	}()

	// Run the server (blocks)
	if err := w.server.Run(w.mux); err != nil {
		return eris.Wrap(err, "failed to run worker server")
	}
	return nil
}

// Shutdown gracefully shuts down the worker server and scheduler
func (w *Worker) Shutdown() {
	w.logger.Info().Msg("shutting down scheduler...")
	w.scheduler.Shutdown()

	w.logger.Info().Msg("shutting down worker server...")
	w.server.Shutdown()

	w.logger.Info().Msg("worker shutdown complete")
}
