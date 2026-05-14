// Package tasks defines background job handlers for Asynq.
//
// Tasks are enqueued by services via TaskClient and processed by the worker.
// Each task type has a payload struct and a handler that implements
// the processing logic.
//
// # Task Types
//
// Task type names follow the "resource:action" convention (lowercase).
//
//   - TypeEmail ("email:send"): Generic email sending with template rendering
//   - TypeMaintenance ("maintenance:cleanup"): Periodic cleanup of expired
//     tokens and finalisation of scheduled account deletions
//
// # Lifecycle
//
// Tasks are registered in Registry and wired into the worker's ServeMux.
// The worker runs tasks with configurable concurrency, retries, and timeouts.
//
// # Metadata
//
// Tasks carry request metadata (request ID, user ID) for distributed tracing.
// Use taskqueue.UnwrapPayload to extract metadata and the typed payload.
package tasks
