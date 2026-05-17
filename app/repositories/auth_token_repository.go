package repositories

import (
	"context"
	"time"

	"go-reasonable-api/app/interfaces/repositories"
	"go-reasonable-api/db/sqlcgen"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotisserie/eris"
)

type AuthTokenRepository struct {
	queries *sqlcgen.Queries
}

func NewAuthTokenRepository(pool *pgxpool.Pool) *AuthTokenRepository {
	return &AuthTokenRepository{
		queries: sqlcgen.New(pool),
	}
}

func (r *AuthTokenRepository) WithTx(tx pgx.Tx) repositories.AuthTokenRepository {
	return &AuthTokenRepository{
		queries: sqlcgen.New(tx),
	}
}

func (r *AuthTokenRepository) Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*sqlcgen.AuthToken, error) {
	token := sqlcgen.AuthToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}

	if err := r.queries.CreateAuthToken(ctx, sqlcgen.CreateAuthTokenParams{
		ID:        token.ID,
		UserID:    token.UserID,
		TokenHash: token.TokenHash,
		ExpiresAt: token.ExpiresAt,
		CreatedAt: token.CreatedAt,
	}); err != nil {
		return nil, eris.Wrap(err, "failed to create auth token")
	}

	return &token, nil
}

func (r *AuthTokenRepository) GetByHash(ctx context.Context, tokenHash string) (*sqlcgen.AuthToken, error) {
	token, err := r.queries.GetAuthTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get auth token by hash")
	}

	return &token, nil
}

func (r *AuthTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	if err := r.queries.RevokeAuthToken(ctx, sqlcgen.RevokeAuthTokenParams{
		RevokedAt: &now,
		ID:        id,
	}); err != nil {
		return eris.Wrap(err, "failed to revoke auth token")
	}
	return nil
}

func (r *AuthTokenRepository) RevokeByHash(ctx context.Context, tokenHash string) error {
	now := time.Now().UTC()
	if err := r.queries.RevokeAuthTokenByHash(ctx, sqlcgen.RevokeAuthTokenByHashParams{
		RevokedAt: &now,
		TokenHash: tokenHash,
	}); err != nil {
		return eris.Wrap(err, "failed to revoke auth token by hash")
	}
	return nil
}

func (r *AuthTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	if err := r.queries.RevokeAllAuthTokensForUser(ctx, sqlcgen.RevokeAllAuthTokensForUserParams{
		RevokedAt: &now,
		UserID:    userID,
	}); err != nil {
		return eris.Wrap(err, "failed to revoke all auth tokens for user")
	}
	return nil
}

func (r *AuthTokenRepository) DeleteExpiredOrRevoked(ctx context.Context) (int64, error) {
	deleted, err := r.queries.DeleteExpiredOrRevokedAuthTokens(ctx, time.Now().UTC())
	if err != nil {
		return 0, eris.Wrap(err, "failed to delete expired or revoked auth tokens")
	}
	return deleted, nil
}

var _ repositories.AuthTokenRepository = (*AuthTokenRepository)(nil)
