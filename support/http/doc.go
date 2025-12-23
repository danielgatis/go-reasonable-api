// Package http configures the Echo web server and provides HTTP utilities.
//
// # Router
//
// Router is the main entry point, created via Wire. It configures:
//   - Middleware stack (request ID, logging, rate limiting, recovery, CORS)
//   - Custom error handler that maps AppError to JSON responses
//   - Request validator using go-playground/validator
//   - All API routes via api.SetupRoutes
//
// # Error Handling
//
// ErrorHandler converts errors to consistent JSON responses:
//   - AppError: uses Code, Message, Details, and StatusCode directly
//   - echo.HTTPError: extracts status and message
//   - Other errors: returns 500 with generic message (details logged, not exposed)
//
// Server errors (5xx) are automatically reported to Sentry with request context.
//
// # Request Context
//
// The reqctx subpackage provides typed accessors for request-scoped values
// (user ID, request ID, auth token). Middleware populates these; handlers
// and services read them via context.Context.
package http
