package services

import (
	"context"

	"go-reasonable-api/db/sqlcgen"

	"github.com/google/uuid"
)

// SessionService manages authentication tokens.
//
// Tokens are opaque strings returned to clients. The service stores only
// SHA-256 hashes, making token theft from the database ineffective.
//
// Create validates credentials and returns both user and token on success.
// CreateForUser generates a token without credential validation (for post-registration).
// ValidateToken returns the token record if valid; callers must check expiration
// and revocation status on the returned AuthToken.
type SessionService interface {
	Create(ctx context.Context, email, password string) (*sqlcgen.User, string, error)
	CreateForUser(ctx context.Context, userID uuid.UUID) (string, error)
	Delete(ctx context.Context, token string) error
	ValidateToken(ctx context.Context, token string) (*sqlcgen.AuthToken, error)
}
