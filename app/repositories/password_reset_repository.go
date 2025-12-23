package repositories

import (
	"context"
	"database/sql"
	"time"

	"go-reasonable-api/app/interfaces/repositories"
	"go-reasonable-api/db/sqlcgen"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

type PasswordResetRepository struct {
	db      *sql.DB
	queries *sqlcgen.Queries
}

func NewPasswordResetRepository(db *sql.DB) *PasswordResetRepository {
	return &PasswordResetRepository{
		db:      db,
		queries: sqlcgen.New(db),
	}
}

func (r *PasswordResetRepository) WithTx(tx *sql.Tx) repositories.PasswordResetRepository {
	return &PasswordResetRepository{
		db:      r.db,
		queries: sqlcgen.New(tx),
	}
}

func (r *PasswordResetRepository) Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*sqlcgen.PasswordReset, error) {
	reset := sqlcgen.PasswordReset{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}

	err := r.queries.CreatePasswordReset(ctx, sqlcgen.CreatePasswordResetParams{
		ID:        reset.ID,
		UserID:    reset.UserID,
		TokenHash: reset.TokenHash,
		ExpiresAt: reset.ExpiresAt,
		CreatedAt: reset.CreatedAt,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to create password reset")
	}

	return &reset, nil
}

func (r *PasswordResetRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*sqlcgen.PasswordReset, error) {
	reset, err := r.queries.GetPasswordResetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get password reset by token hash")
	}

	return &reset, nil
}

func (r *PasswordResetRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	err := r.queries.MarkPasswordResetUsed(ctx, sqlcgen.MarkPasswordResetUsedParams{
		UsedAt: &now,
		ID:     id,
	})
	if err != nil {
		return eris.Wrap(err, "failed to mark password reset as used")
	}
	return nil
}

func (r *PasswordResetRepository) InvalidateAllForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	err := r.queries.InvalidateAllPasswordResetsForUser(ctx, sqlcgen.InvalidateAllPasswordResetsForUserParams{
		UsedAt: &now,
		UserID: userID,
	})
	if err != nil {
		return eris.Wrap(err, "failed to invalidate all password resets for user")
	}
	return nil
}

func (r *PasswordResetRepository) DeleteExpiredOrUsed(ctx context.Context) (int64, error) {
	deleted, err := r.queries.DeleteExpiredOrUsedPasswordResets(ctx, time.Now().UTC())
	if err != nil {
		return 0, eris.Wrap(err, "failed to delete expired or used password resets")
	}
	return deleted, nil
}

var _ repositories.PasswordResetRepository = (*PasswordResetRepository)(nil)
