// Package services defines the business logic contracts for the application.
//
// Services encapsulate domain operations and orchestrate repositories.
// All methods accept context.Context as the first parameter for cancellation,
// timeouts, and request-scoped values (user ID, request ID, logger).
//
// # Error Handling
//
// Services return domain errors from app/errors for known business rule
// violations (e.g., ErrUserNotFound, ErrInvalidCredentials). Infrastructure
// errors are wrapped with eris to preserve stack traces.
//
// # Transaction Boundaries
//
// Services own transaction boundaries. Operations requiring atomicity
// use TxManager.RunInTx internally. Callers should not wrap service
// calls in transactions.
package services
