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

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
)

type EmailVerificationService struct {
	config                *config.Config
	userRepo              repositories.UserRepository
	emailVerificationRepo repositories.EmailVerificationRepository
	txManager             *db.TxManager
	taskClient            support.TaskClient
}

func NewEmailVerificationService(
	cfg *config.Config,
	userRepo repositories.UserRepository,
	emailVerificationRepo repositories.EmailVerificationRepository,
	txManager *db.TxManager,
	taskClient support.TaskClient,
) *EmailVerificationService {
	return &EmailVerificationService{
		config:                cfg,
		userRepo:              userRepo,
		emailVerificationRepo: emailVerificationRepo,
		txManager:             txManager,
		taskClient:            taskClient,
	}
}

func (s *EmailVerificationService) Send(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return eris.Wrap(err, "failed to get user by id")
	}

	if user.EmailVerifiedAt != nil {
		return errors.ErrEmailAlreadyVerified
	}

	verificationToken, err := GenerateSecureToken(32)
	if err != nil {
		return eris.Wrap(err, "failed to generate verification token")
	}

	tokenHash := HashToken(verificationToken)
	expiresAt := time.Now().UTC().Add(s.config.Auth.EmailConfirmationTokenTTL)

	if err := s.emailVerificationRepo.InvalidateAllForUser(ctx, userID); err != nil {
		return eris.Wrap(err, "failed to invalidate email verifications for user")
	}

	_, err = s.emailVerificationRepo.Create(ctx, userID, tokenHash, expiresAt)
	if err != nil {
		return eris.Wrap(err, "failed to create email verification")
	}

	verificationLink := fmt.Sprintf("%s/verify-email?token=%s", s.config.App.BaseURL, url.QueryEscape(verificationToken))

	s.taskClient.EnqueueCtx(ctx, tasks.TypeEmail, tasks.EmailPayload{
		To:       user.Email,
		Subject:  "Confirme seu Email - ZapAgenda",
		Template: "email-verification",
		Data: map[string]any{
			"Name":             user.Name,
			"VerificationLink": verificationLink,
		},
	}, tasks.EmailTaskOptions(s.config)...)

	return nil
}

func (s *EmailVerificationService) Verify(ctx context.Context, token string) error {
	tokenHash := HashToken(token)

	verification, err := s.emailVerificationRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			return errors.ErrInvalidVerificationToken
		}
		return eris.Wrap(err, "failed to get email verification by token hash")
	}

	if verification.UsedAt != nil {
		return errors.ErrTokenAlreadyUsed
	}

	if time.Now().UTC().After(verification.ExpiresAt) {
		return errors.ErrTokenExpired
	}

	return s.txManager.RunInTx(ctx, func(tx *sql.Tx) error {
		txUserRepo := s.userRepo.WithTx(tx)
		txEmailVerificationRepo := s.emailVerificationRepo.WithTx(tx)

		if err := txUserRepo.MarkEmailVerified(ctx, verification.UserID); err != nil {
			return eris.Wrap(err, "failed to mark email as verified")
		}

		if err := txEmailVerificationRepo.MarkUsed(ctx, verification.ID); err != nil {
			return eris.Wrap(err, "failed to mark email verification as used")
		}

		if err := txEmailVerificationRepo.InvalidateAllForUser(ctx, verification.UserID); err != nil {
			return eris.Wrap(err, "failed to invalidate email verifications for user")
		}
		return nil
	})
}

func (s *EmailVerificationService) Resend(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			return nil
		}
		return eris.Wrap(err, "failed to get user by email")
	}

	if user.EmailVerifiedAt != nil {
		return nil
	}

	return s.Send(ctx, user.ID)
}

var _ services.EmailVerificationService = (*EmailVerificationService)(nil)
