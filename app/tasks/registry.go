package tasks

import (
	"github.com/hibiken/asynq"
	"github.com/rotisserie/eris"
)

// Registry centralizes all task handler and scheduled task registration.
// Add new tasks here to register them with the worker.
type Registry struct {
	emailTask   *EmailTask
	cleanupTask *CleanupTask
}

func NewRegistry(emailTask *EmailTask, cleanupTask *CleanupTask) *Registry {
	return &Registry{
		emailTask:   emailTask,
		cleanupTask: cleanupTask,
	}
}

// RegisterHandlers registers all task handlers with the mux.
// Add new task handlers here.
func (r *Registry) RegisterHandlers(mux *asynq.ServeMux) {
	mux.HandleFunc(TypeEmail, r.emailTask.Handle)
	mux.HandleFunc(TypeCleanup, r.cleanupTask.Handle)
}

// RegisterScheduledTasks registers all periodic tasks with the scheduler.
// Add new scheduled tasks here.
func (r *Registry) RegisterScheduledTasks(scheduler *asynq.Scheduler) error {
	// Cleanup expired tokens every hour
	if _, err := scheduler.Register("@every 1h", asynq.NewTask(TypeCleanup, nil)); err != nil {
		return eris.Wrap(err, "failed to register cleanup task")
	}

	return nil
}
