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

type UserRepository struct {
	queries *sqlcgen.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		queries: sqlcgen.New(pool),
	}
}

func (r *UserRepository) WithTx(tx pgx.Tx) repositories.UserRepository {
	return &UserRepository{
		queries: sqlcgen.New(tx),
	}
}

func (r *UserRepository) Create(ctx context.Context, name, email, passwordHash string) (*sqlcgen.User, error) {
	now := time.Now().UTC()

	user, err := r.queries.CreateUser(ctx, sqlcgen.CreateUserParams{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	})
	if err != nil {
		return nil, eris.Wrap(err, "failed to create user")
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*sqlcgen.User, error) {
	user, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get user by id")
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*sqlcgen.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, eris.Wrap(err, "failed to get user by email")
	}

	return &user, nil
}

func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	if err := r.queries.UpdateUserPassword(ctx, sqlcgen.UpdateUserPasswordParams{
		PasswordHash: passwordHash,
		UpdatedAt:    time.Now().UTC(),
		ID:           userID,
	}); err != nil {
		return eris.Wrap(err, "failed to update user password")
	}
	return nil
}

func (r *UserRepository) MarkEmailVerified(ctx context.Context, userID uuid.UUID) error {
	now := time.Now().UTC()
	if err := r.queries.MarkUserEmailVerified(ctx, sqlcgen.MarkUserEmailVerifiedParams{
		EmailVerifiedAt: &now,
		UpdatedAt:       now,
		ID:              userID,
	}); err != nil {
		return eris.Wrap(err, "failed to mark email as verified")
	}
	return nil
}

func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	exists, err := r.queries.EmailExists(ctx, email)
	if err != nil {
		return false, eris.Wrap(err, "failed to check if email exists")
	}
	return exists, nil
}

func (r *UserRepository) ScheduleDeletion(ctx context.Context, userID uuid.UUID, scheduledAt time.Time) error {
	if err := r.queries.ScheduleUserDeletion(ctx, sqlcgen.ScheduleUserDeletionParams{
		DeletionScheduledAt: &scheduledAt,
		UpdatedAt:           time.Now().UTC(),
		ID:                  userID,
	}); err != nil {
		return eris.Wrap(err, "failed to schedule user deletion")
	}
	return nil
}

func (r *UserRepository) CancelDeletion(ctx context.Context, userID uuid.UUID) error {
	if err := r.queries.CancelUserDeletion(ctx, sqlcgen.CancelUserDeletionParams{
		UpdatedAt: time.Now().UTC(),
		ID:        userID,
	}); err != nil {
		return eris.Wrap(err, "failed to cancel user deletion")
	}
	return nil
}

func (r *UserRepository) DeleteScheduledUsers(ctx context.Context) (int64, error) {
	now := time.Now().UTC()
	deleted, err := r.queries.DeleteScheduledUsers(ctx, &now)
	if err != nil {
		return 0, eris.Wrap(err, "failed to delete scheduled users")
	}
	return deleted, nil
}

var _ repositories.UserRepository = (*UserRepository)(nil)
