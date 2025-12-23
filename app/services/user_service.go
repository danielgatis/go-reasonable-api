package services

import (
	"context"
	"database/sql"
	"time"

	"go-reasonable-api/app/errors"
	"go-reasonable-api/app/interfaces/repositories"
	"go-reasonable-api/app/interfaces/services"
	"go-reasonable-api/app/interfaces/support"
	"go-reasonable-api/app/tasks"
	"go-reasonable-api/db/sqlcgen"
	"go-reasonable-api/support/config"
	"go-reasonable-api/support/db"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"golang.org/x/crypto/bcrypt"
)

// UserService implements services.UserService.
type UserService struct {
	config        *config.Config
	txManager     *db.TxManager
	userRepo      repositories.UserRepository
	authTokenRepo repositories.AuthTokenRepository
	taskClient    support.TaskClient
}

func NewUserService(cfg *config.Config, txManager *db.TxManager, userRepo repositories.UserRepository, authTokenRepo repositories.AuthTokenRepository, taskClient support.TaskClient) *UserService {
	return &UserService{
		config:        cfg,
		txManager:     txManager,
		userRepo:      userRepo,
		authTokenRepo: authTokenRepo,
		taskClient:    taskClient,
	}
}

func (s *UserService) Create(ctx context.Context, name, email, password string) (*sqlcgen.User, error) {
	exists, err := s.userRepo.EmailExists(ctx, email)
	if err != nil {
		return nil, eris.Wrap(err, "failed to check if email exists")
	}
	if exists {
		return nil, errors.ErrEmailAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), s.config.Auth.BcryptCost)
	if err != nil {
		return nil, eris.Wrap(err, "failed to generate password hash")
	}

	return s.userRepo.Create(ctx, name, email, string(passwordHash))
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*sqlcgen.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrUserNotFound
		}
		return nil, eris.Wrap(err, "failed to get user by ID")
	}

	return user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*sqlcgen.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			return nil, errors.ErrUserNotFound
		}
		return nil, eris.Wrap(err, "failed to get user by email")
	}

	return user, nil
}

func (s *UserService) ScheduleDeletion(ctx context.Context, userID uuid.UUID) error {
	scheduledAt := time.Now().UTC().Add(s.config.Auth.AccountDeletionDelay)

	var userEmail, userName string

	// Run all database operations in a transaction to avoid TOCTOU race condition
	err := s.txManager.RunInTx(ctx, func(tx *sql.Tx) error {
		userRepoTx := s.userRepo.WithTx(tx)
		authTokenRepoTx := s.authTokenRepo.WithTx(tx)

		// Get user inside transaction to ensure consistent read
		user, err := userRepoTx.GetByID(ctx, userID)
		if err != nil {
			if eris.Is(err, sql.ErrNoRows) {
				return errors.ErrUserNotFound
			}
			return eris.Wrap(err, "failed to get user by ID")
		}

		if user.DeletionScheduledAt != nil {
			return errors.ErrDeletionAlreadyScheduled
		}

		// Store user info for email after transaction commits
		userEmail = user.Email
		userName = user.Name

		if err := userRepoTx.ScheduleDeletion(ctx, userID, scheduledAt); err != nil {
			return eris.Wrap(err, "failed to schedule user deletion")
		}

		// Revoke all auth tokens to log the user out
		if err := authTokenRepoTx.RevokeAllForUser(ctx, userID); err != nil {
			return eris.Wrap(err, "failed to revoke all auth tokens for user")
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.taskClient.EnqueueCtx(ctx, tasks.TypeEmail, tasks.EmailPayload{
		To:       userEmail,
		Subject:  "Sua conta será excluída - ZapAgenda",
		Template: "account-deletion-scheduled",
		Data: map[string]any{
			"Name":        userName,
			"ScheduledAt": scheduledAt.Format("02/01/2006"),
			"DaysLeft":    int(s.config.Auth.AccountDeletionDelay.Hours() / 24),
		},
	}, tasks.EmailTaskOptions(s.config)...)

	return nil
}

var _ services.UserService = (*UserService)(nil)
