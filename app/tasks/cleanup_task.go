package tasks

import (
	"context"

	"go-reasonable-api/app/interfaces/repositories"
	"go-reasonable-api/support/logger"
	"go-reasonable-api/support/taskqueue"

	"github.com/hibiken/asynq"
	"github.com/rotisserie/eris"
	"github.com/rs/zerolog"
)

const TypeCleanup = "cleanup:tokens"

// CleanupTask handles periodic cleanup of expired tokens and scheduled account deletions
type CleanupTask struct {
	logger                *zerolog.Logger
	authTokenRepo         repositories.AuthTokenRepository
	passwordResetRepo     repositories.PasswordResetRepository
	emailVerificationRepo repositories.EmailVerificationRepository
	userRepo              repositories.UserRepository
}

func NewCleanupTask(
	logger *zerolog.Logger,
	authTokenRepo repositories.AuthTokenRepository,
	passwordResetRepo repositories.PasswordResetRepository,
	emailVerificationRepo repositories.EmailVerificationRepository,
	userRepo repositories.UserRepository,
) *CleanupTask {
	return &CleanupTask{
		logger:                logger,
		authTokenRepo:         authTokenRepo,
		passwordResetRepo:     passwordResetRepo,
		emailVerificationRepo: emailVerificationRepo,
		userRepo:              userRepo,
	}
}

func (t *CleanupTask) Handle(ctx context.Context, task *asynq.Task) error {
	meta, err := taskqueue.UnwrapPayload(task.Payload(), &struct{}{})
	if err != nil {
		// For periodic tasks, payload might be empty
		meta = taskqueue.TaskMetadata{}
	}

	ctx = meta.LoggerContext(ctx, t.logger)
	log := logger.Ctx(ctx)

	log.Info().Str("task", TypeCleanup).Msg("starting token cleanup")

	// Cleanup auth tokens
	authDeleted, err := t.authTokenRepo.DeleteExpiredOrRevoked(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to cleanup auth tokens")
		return eris.Wrap(err, "failed to cleanup auth tokens")
	}

	// Cleanup password reset tokens
	passwordDeleted, err := t.passwordResetRepo.DeleteExpiredOrUsed(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to cleanup password reset tokens")
		return eris.Wrap(err, "failed to cleanup password reset tokens")
	}

	// Cleanup email verification tokens
	emailDeleted, err := t.emailVerificationRepo.DeleteExpiredOrUsed(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to cleanup email verification tokens")
		return eris.Wrap(err, "failed to cleanup email verification tokens")
	}

	// Delete users with scheduled deletion date in the past
	usersDeleted, err := t.userRepo.DeleteScheduledUsers(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to delete scheduled users")
		return eris.Wrap(err, "failed to delete scheduled users")
	}

	log.Info().
		Int64("auth_tokens_deleted", authDeleted).
		Int64("password_resets_deleted", passwordDeleted).
		Int64("email_verifications_deleted", emailDeleted).
		Int64("users_deleted", usersDeleted).
		Msg("cleanup completed")

	return nil
}
