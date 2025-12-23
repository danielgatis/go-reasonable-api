package repositories

import (
	"context"
	"database/sql"
	"time"

	"go-reasonable-api/db/sqlcgen"

	"github.com/google/uuid"
)

// PasswordResetRepository manages password reset token persistence.
//
// Tokens are single-use. MarkUsed sets used_at, preventing reuse.
// InvalidateAllForUser marks all pending resets as used when a new
// reset is requested or password is changed.
type PasswordResetRepository interface {
	WithTx(tx *sql.Tx) PasswordResetRepository

	Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*sqlcgen.PasswordReset, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*sqlcgen.PasswordReset, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
	InvalidateAllForUser(ctx context.Context, userID uuid.UUID) error
	DeleteExpiredOrUsed(ctx context.Context) (int64, error)
}
