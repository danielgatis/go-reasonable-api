package services

import (
	"context"
)

// PasswordResetService handles password reset flows.
//
// Create is idempotent: if email doesn't exist, it silently succeeds
// to prevent email enumeration attacks. Tokens are sent via async email task.
//
// Execute validates the token, updates the password, and revokes all
// existing auth tokens for the user in a single transaction.
type PasswordResetService interface {
	Create(ctx context.Context, email string) error
	Execute(ctx context.Context, token, newPassword string) error
}
