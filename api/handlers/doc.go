// Package handlers implements HTTP request handlers for the API.
//
// Handlers are thin adapters between HTTP and services. They:
//  1. Bind and validate request payloads (via bind.AndValidate)
//  2. Extract context values (user ID, request ID) from middleware
//  3. Call services with domain parameters
//  4. Transform results to response DTOs
//
// # Error Handling
//
// Handlers return errors directly; the global ErrorHandler transforms them
// to JSON responses. Services return domain errors (from app/errors) which
// carry HTTP status codes.
//
// # Authentication
//
// Routes requiring authentication use AuthMiddleware, which populates
// reqctx with user ID and token. Handlers access these via reqctx.GetUserID.
//
// # Documentation
//
// Handlers are annotated with swaggo comments for OpenAPI generation.
// Run `make generate` to update api/docs after changing annotations.
package handlers
