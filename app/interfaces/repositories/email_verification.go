package repositories

import (
	"context"
	"database/sql"
	"time"

	"go-reasonable-api/db/sqlcgen"

	"github.com/google/uuid"
)

// EmailVerificationRepository manages email verification token persistence.
//
// Similar lifecycle to PasswordResetRepository: tokens are single-use
// and previous tokens are invalidated when a new one is created.
type EmailVerificationRepository interface {
	WithTx(tx *sql.Tx) EmailVerificationRepository

	Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*sqlcgen.EmailVerification, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*sqlcgen.EmailVerification, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
	InvalidateAllForUser(ctx context.Context, userID uuid.UUID) error
	DeleteExpiredOrUsed(ctx context.Context) (int64, error)
}
