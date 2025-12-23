package repositories

import (
	"context"
	"database/sql"
	"time"

	"go-reasonable-api/db/sqlcgen"

	"github.com/google/uuid"
)

// AuthTokenRepository manages authentication token persistence.
//
// Tokens are stored as SHA-256 hashes. GetByHash accepts the hash,
// not the raw token. The service layer handles hashing.
//
// Revoke marks a token as revoked (soft delete). RevokeAllForUser
// is used when password changes to invalidate all sessions.
// DeleteExpiredOrRevoked permanently removes old records.
type AuthTokenRepository interface {
	WithTx(tx *sql.Tx) AuthTokenRepository

	Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*sqlcgen.AuthToken, error)
	GetByHash(ctx context.Context, tokenHash string) (*sqlcgen.AuthToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeByHash(ctx context.Context, tokenHash string) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	DeleteExpiredOrRevoked(ctx context.Context) (int64, error)
}
