// Package errors provides structured application errors with HTTP status mapping.
//
// AppError is the primary error type, carrying a code (for clients), message,
// optional details, and HTTP status. Sentinel errors are defined in app/errors.
//
// # Usage Pattern
//
// Define sentinel errors for known domain conditions:
//
//	var ErrUserNotFound = errors.NotFoundf("user")
//
// Return them directly from services; the HTTP layer maps StatusCode automatically.
// Use WithDetails to add context without losing error identity:
//
//	return apperrors.ErrValidationFailed.WithDetails(map[string]any{"field": "email"})
//
// # Stack Traces
//
// AppError wraps eris errors internally. Use StackTrace or StackTraceJSON
// for debugging; these are logged but never exposed to clients.
package errors

import (
	"net/http"
	"strings"

	"github.com/rotisserie/eris"
)

// AppError represents an application error with code, message, and HTTP status.
// Code is a stable identifier for client error handling (e.g., "USER_NOT_FOUND").
// Message is human-readable. Details provides structured context.
type AppError struct {
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	Details    map[string]any `json:"details,omitempty"`
	StatusCode int            `json:"-"`
	cause      error
}

func (e *AppError) Error() string {
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.cause
}

// Is implements errors.Is comparison by Code.
// This allows errors.Is(err, ErrUserNotFound) to work even when
// the error was created with WithDetails.
func (e *AppError) Is(target error) bool {
	if t, ok := target.(*AppError); ok {
		return e.Code == t.Code
	}
	return false
}

// WithDetails returns a new AppError with the given details added.
// This creates a copy so the original sentinel error is not modified.
func (e *AppError) WithDetails(details map[string]any) *AppError {
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		Details:    details,
		StatusCode: e.StatusCode,
		cause:      e.cause,
	}
}

// WithDetail returns a new AppError with a single key-value detail added.
func (e *AppError) WithDetail(key string, value any) *AppError {
	details := make(map[string]any)
	for k, v := range e.Details {
		details[k] = v
	}
	details[key] = value
	return &AppError{
		Code:       e.Code,
		Message:    e.Message,
		Details:    details,
		StatusCode: e.StatusCode,
		cause:      e.cause,
	}
}

// -----------------------------------------------------------------------------
// Constructors
// -----------------------------------------------------------------------------

func New(code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: http.StatusUnprocessableEntity,
		cause:      eris.New(message),
	}
}

func NewWithStatus(code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		cause:      eris.New(message),
	}
}

func NewWithDetails(code, message string, statusCode int, details map[string]any) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: statusCode,
		cause:      eris.New(message),
	}
}

func Unauthorized(code, message string) *AppError {
	return NewWithStatus(code, message, http.StatusUnauthorized)
}

func NotFound(code, message string) *AppError {
	return NewWithStatus(code, message, http.StatusNotFound)
}

func NotFoundf(resource string) *AppError {
	return NewWithStatus(
		strings.ToUpper(resource)+"_NOT_FOUND",
		resource+" not found",
		http.StatusNotFound,
	)
}

func BadRequest(code, message string) *AppError {
	return NewWithStatus(code, message, http.StatusBadRequest)
}

func InternalError(code, message string) *AppError {
	return NewWithStatus(code, message, http.StatusInternalServerError)
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func Is(err error) (*AppError, bool) {
	if err == nil {
		return nil, false
	}
	if ae, ok := err.(*AppError); ok {
		return ae, true
	}
	return nil, false
}

func Wrap(err error, message string) error {
	return eris.Wrap(err, message)
}

func WrapWithCode(err error, code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		cause:      eris.Wrap(err, message),
	}
}

func WrapWithCodeAndStatus(err error, code, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		cause:      eris.Wrap(err, message),
	}
}

func Cause(err error) error {
	return eris.Cause(err)
}

func StackTrace(err error) string {
	if err == nil {
		return ""
	}
	if ae, ok := err.(*AppError); ok && ae.cause != nil {
		return eris.ToString(ae.cause, true)
	}
	return eris.ToString(err, true)
}

func StackTraceJSON(err error) map[string]any {
	if err == nil {
		return nil
	}
	if ae, ok := err.(*AppError); ok && ae.cause != nil {
		return eris.ToJSON(ae.cause, true)
	}
	return eris.ToJSON(err, true)
}
