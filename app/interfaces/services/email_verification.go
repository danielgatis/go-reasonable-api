package services

import (
	"context"

	"github.com/google/uuid"
)

// EmailVerificationService manages email address verification.
//
// Send queues a verification email asynchronously. It's safe to call
// multiple times; previous tokens are invalidated when a new one is created.
//
// Resend is for unauthenticated users who need a new verification email.
// It silently succeeds for non-existent emails to prevent enumeration.
type EmailVerificationService interface {
	Send(ctx context.Context, userID uuid.UUID) error
	Verify(ctx context.Context, token string) error
	Resend(ctx context.Context, email string) error
}
