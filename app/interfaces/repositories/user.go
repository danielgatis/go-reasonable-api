package repositories

import (
	"context"
	"time"

	"go-reasonable-api/db/sqlcgen"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// UserRepository provides user persistence operations.
//
// GetByID and GetByEmail return a wrapped pgx.ErrNoRows when the user is not
// found. Callers should check via eris.Is(err, pgx.ErrNoRows).
//
// DeleteScheduledUsers removes users whose deletion_scheduled_at has passed.
// Returns the count of deleted users for logging purposes.
type UserRepository interface {
	WithTx(tx pgx.Tx) UserRepository

	Create(ctx context.Context, name, email, passwordHash string) (*sqlcgen.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*sqlcgen.User, error)
	GetByEmail(ctx context.Context, email string) (*sqlcgen.User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	MarkEmailVerified(ctx context.Context, userID uuid.UUID) error
	EmailExists(ctx context.Context, email string) (bool, error)
	ScheduleDeletion(ctx context.Context, userID uuid.UUID, scheduledAt time.Time) error
	CancelDeletion(ctx context.Context, userID uuid.UUID) error
	DeleteScheduledUsers(ctx context.Context) (int64, error)
}
