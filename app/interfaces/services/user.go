package services

import (
	"context"

	"go-reasonable-api/db/sqlcgen"

	"github.com/google/uuid"
)

// UserService manages user lifecycle operations.
//
// Create hashes passwords using bcrypt before storage.
// ScheduleDeletion implements soft-delete with a configurable delay period,
// allowing users to cancel deletion by logging in before the deadline.
type UserService interface {
	Create(ctx context.Context, name, email, password string) (*sqlcgen.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*sqlcgen.User, error)
	GetByEmail(ctx context.Context, email string) (*sqlcgen.User, error)
	ScheduleDeletion(ctx context.Context, userID uuid.UUID) error
}
