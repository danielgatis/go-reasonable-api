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

type EmailVerificationRepository struct {
	queries *sqlcgen.Queries
}

func NewEmailVerificationRepository(pool *pgxpool.Pool) *EmailVerificationRepository {
	return &EmailVerificationRepository{
		queries: sqlcgen.New(pool),
	}
}

func (r *EmailVerificationRepository) WithTx(tx pgx.Tx) repositories.EmailVerificationRepository {
	return &EmailVerificationRepository{
		queries: sqlcgen.New(tx),
	}
}

func (r *EmailVerificationRepository) Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*sqlcgen.EmailVerification, error) {
	verification := sqlcgen.EmailVerification{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().UTC(),
	}

	if err := r.queries.CreateEmailVerification(ctx, sqlcgen.CreateEmailVerificationParams{
		ID:        verification.ID,
		UserID:    verification.UserID,
		TokenHash: verification.TokenHash,
		ExpiresAt: verification.ExpiresAt,
		CreatedAt: verification.CreatedAt,
	}); err != nil {
		return nil, eris.Wrap(err, "failed to create email verification")
	}

	return &verification, nil
}

func (r *EmailVerificationRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*sqlcgen.EmailVerification, error) {
	verification, err := r.queries.GetEmailVerificationByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get email verification by token hash")
	}

	return &verification, nil
}

func (r *EmailVerificationRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	if err := r.queries.MarkEmailVerificationUsed(ctx, sqlcgen.MarkEmailVerificationUsedParams{
		UsedAt: &now,
		ID:     id,
	}); err != nil {
		return eris.Wrap(err, "failed to mark email verification as used")
	}
	return nil
}

func (r *EmailVerificationRepository) InvalidateAllForUser(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	if err := r.queries.InvalidateAllEmailVerificationsForUser(ctx, sqlcgen.InvalidateAllEmailVerificationsForUserParams{
		UsedAt: &now,
		UserID: userID,
	}); err != nil {
		return eris.Wrap(err, "failed to invalidate all email verifications for user")
	}
	return nil
}

func (r *EmailVerificationRepository) DeleteExpiredOrUsed(ctx context.Context) (int64, error) {
	deleted, err := r.queries.DeleteExpiredOrUsedEmailVerifications(ctx, time.Now().UTC())
	if err != nil {
		return 0, eris.Wrap(err, "failed to delete expired or used email verifications")
	}
	return deleted, nil
}

var _ repositories.EmailVerificationRepository = (*EmailVerificationRepository)(nil)
