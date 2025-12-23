// Package errors defines domain error sentinels for the application.
//
// These errors represent known business rule violations and are returned
// by services. The HTTP layer maps them to appropriate status codes via
// the StatusCode field inherited from support/errors.AppError.
//
// Errors are compared by Code using errors.Is, so they can be wrapped
// with additional context while maintaining identity.
package errors

import (
	"go-reasonable-api/support/errors"
)

var (
	ErrInvalidCredentials = errors.Unauthorized("INVALID_CREDENTIALS", "invalid credentials")
	ErrMissingAuthHeader  = errors.Unauthorized("MISSING_AUTH_HEADER", "missing authorization header")
	ErrInvalidAuthFormat  = errors.Unauthorized("INVALID_AUTH_FORMAT", "invalid authorization header format")
)

var (
	ErrInvalidToken             = errors.Unauthorized("INVALID_TOKEN", "invalid token")
	ErrTokenExpired             = errors.Unauthorized("TOKEN_EXPIRED", "token expired")
	ErrTokenRevoked             = errors.Unauthorized("TOKEN_REVOKED", "token revoked")
	ErrTokenAlreadyUsed         = errors.New("TOKEN_ALREADY_USED", "token already used")
	ErrInvalidResetToken        = errors.New("INVALID_RESET_TOKEN", "invalid or expired reset token")
	ErrInvalidVerificationToken = errors.New("INVALID_VERIFICATION_TOKEN", "invalid or expired verification token")
)

var (
	ErrUserNotFound             = errors.NotFoundf("user")
	ErrEmailAlreadyExists       = errors.New("EMAIL_ALREADY_EXISTS", "email already exists")
	ErrEmailAlreadyVerified     = errors.New("EMAIL_ALREADY_VERIFIED", "email already verified")
	ErrDeletionAlreadyScheduled = errors.New("DELETION_ALREADY_SCHEDULED", "account deletion is already scheduled")
)
