package services

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"go-reasonable-api/app/errors"
	"go-reasonable-api/app/interfaces/repositories"
	"go-reasonable-api/app/interfaces/services"
	"go-reasonable-api/app/interfaces/support"
	"go-reasonable-api/app/tasks"
	"go-reasonable-api/support/config"
	"go-reasonable-api/support/db"

	"github.com/rotisserie/eris"
	"golang.org/x/crypto/bcrypt"
)

type PasswordResetService struct {
	config            *config.Config
	userRepo          repositories.UserRepository
	passwordResetRepo repositories.PasswordResetRepository
	authTokenRepo     repositories.AuthTokenRepository
	txManager         *db.TxManager
	taskClient        support.TaskClient
}

func NewPasswordResetService(
	cfg *config.Config,
	userRepo repositories.UserRepository,
	passwordResetRepo repositories.PasswordResetRepository,
	authTokenRepo repositories.AuthTokenRepository,
	txManager *db.TxManager,
	taskClient support.TaskClient,
) *PasswordResetService {
	return &PasswordResetService{
		config:            cfg,
		userRepo:          userRepo,
		passwordResetRepo: passwordResetRepo,
		authTokenRepo:     authTokenRepo,
		txManager:         txManager,
		taskClient:        taskClient,
	}
}

func (s *PasswordResetService) Create(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			return nil
		}
		return eris.Wrap(err, "failed to get user by email")
	}

	resetToken, err := GenerateSecureToken(32)
	if err != nil {
		return eris.Wrap(err, "failed to generate reset token")
	}

	tokenHash := HashToken(resetToken)
	expiresAt := time.Now().UTC().Add(s.config.Auth.PasswordResetTokenTTL)

	_, err = s.passwordResetRepo.Create(ctx, user.ID, tokenHash, expiresAt)
	if err != nil {
		return eris.Wrap(err, "failed to create password reset")
	}

	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.config.App.BaseURL, url.QueryEscape(resetToken))

	s.taskClient.EnqueueCtx(ctx, tasks.TypeEmail, tasks.EmailPayload{
		To:       user.Email,
		Subject:  "Redefinir Senha - ZapAgenda",
		Template: "password-reset",
		Data: map[string]any{
			"Name":      user.Name,
			"ResetLink": resetLink,
		},
	}, tasks.EmailTaskOptions(s.config)...)

	return nil
}

func (s *PasswordResetService) Execute(ctx context.Context, token, newPassword string) error {
	tokenHash := HashToken(token)

	reset, err := s.passwordResetRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			return errors.ErrInvalidResetToken
		}
		return eris.Wrap(err, "failed to get password reset by token hash")
	}

	if reset.UsedAt != nil {
		return errors.ErrTokenAlreadyUsed
	}

	if time.Now().UTC().After(reset.ExpiresAt) {
		return errors.ErrTokenExpired
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), s.config.Auth.BcryptCost)
	if err != nil {
		return eris.Wrap(err, "failed to hash password")
	}

	return s.txManager.RunInTx(ctx, func(tx *sql.Tx) error {
		txUserRepo := s.userRepo.WithTx(tx)
		txPasswordResetRepo := s.passwordResetRepo.WithTx(tx)
		txAuthTokenRepo := s.authTokenRepo.WithTx(tx)

		if err := txUserRepo.UpdatePassword(ctx, reset.UserID, string(passwordHash)); err != nil {
			return eris.Wrap(err, "failed to update password")
		}

		if err := txPasswordResetRepo.MarkUsed(ctx, reset.ID); err != nil {
			return eris.Wrap(err, "failed to mark password reset as used")
		}

		if err := txPasswordResetRepo.InvalidateAllForUser(ctx, reset.UserID); err != nil {
			return eris.Wrap(err, "failed to invalidate password resets for user")
		}

		if err := txAuthTokenRepo.RevokeAllForUser(ctx, reset.UserID); err != nil {
			return eris.Wrap(err, "failed to revoke auth tokens for user")
		}
		return nil
	})
}

var _ services.PasswordResetService = (*PasswordResetService)(nil)
