package tasks_test

import (
	"context"
	"database/sql"
	"testing"

	mocks "go-reasonable-api/app/mocks/repositories"
	"go-reasonable-api/app/tasks"

	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestLogger() *zerolog.Logger {
	logger := zerolog.Nop()
	return &logger
}

func TestCleanupTask_Handle(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		setupMock   func(*mocks.MockAuthTokenRepository, *mocks.MockPasswordResetRepository, *mocks.MockEmailVerificationRepository, *mocks.MockUserRepository)
		expectedErr bool
	}{
		{
			name: "cleans up all tokens and users successfully",
			setupMock: func(authRepo *mocks.MockAuthTokenRepository, pwRepo *mocks.MockPasswordResetRepository, emailRepo *mocks.MockEmailVerificationRepository, userRepo *mocks.MockUserRepository) {
				authRepo.EXPECT().DeleteExpiredOrRevoked(mock.Anything).Return(int64(5), nil)
				pwRepo.EXPECT().DeleteExpiredOrUsed(mock.Anything).Return(int64(3), nil)
				emailRepo.EXPECT().DeleteExpiredOrUsed(mock.Anything).Return(int64(2), nil)
				userRepo.EXPECT().DeleteScheduledUsers(mock.Anything).Return(int64(1), nil)
			},
			expectedErr: false,
		},
		{
			name: "returns error when auth token cleanup fails",
			setupMock: func(authRepo *mocks.MockAuthTokenRepository, pwRepo *mocks.MockPasswordResetRepository, emailRepo *mocks.MockEmailVerificationRepository, userRepo *mocks.MockUserRepository) {
				authRepo.EXPECT().DeleteExpiredOrRevoked(mock.Anything).Return(int64(0), sql.ErrConnDone)
			},
			expectedErr: true,
		},
		{
			name: "returns error when password reset cleanup fails",
			setupMock: func(authRepo *mocks.MockAuthTokenRepository, pwRepo *mocks.MockPasswordResetRepository, emailRepo *mocks.MockEmailVerificationRepository, userRepo *mocks.MockUserRepository) {
				authRepo.EXPECT().DeleteExpiredOrRevoked(mock.Anything).Return(int64(5), nil)
				pwRepo.EXPECT().DeleteExpiredOrUsed(mock.Anything).Return(int64(0), sql.ErrConnDone)
			},
			expectedErr: true,
		},
		{
			name: "returns error when email verification cleanup fails",
			setupMock: func(authRepo *mocks.MockAuthTokenRepository, pwRepo *mocks.MockPasswordResetRepository, emailRepo *mocks.MockEmailVerificationRepository, userRepo *mocks.MockUserRepository) {
				authRepo.EXPECT().DeleteExpiredOrRevoked(mock.Anything).Return(int64(5), nil)
				pwRepo.EXPECT().DeleteExpiredOrUsed(mock.Anything).Return(int64(3), nil)
				emailRepo.EXPECT().DeleteExpiredOrUsed(mock.Anything).Return(int64(0), sql.ErrConnDone)
			},
			expectedErr: true,
		},
		{
			name: "returns error when user deletion fails",
			setupMock: func(authRepo *mocks.MockAuthTokenRepository, pwRepo *mocks.MockPasswordResetRepository, emailRepo *mocks.MockEmailVerificationRepository, userRepo *mocks.MockUserRepository) {
				authRepo.EXPECT().DeleteExpiredOrRevoked(mock.Anything).Return(int64(5), nil)
				pwRepo.EXPECT().DeleteExpiredOrUsed(mock.Anything).Return(int64(3), nil)
				emailRepo.EXPECT().DeleteExpiredOrUsed(mock.Anything).Return(int64(2), nil)
				userRepo.EXPECT().DeleteScheduledUsers(mock.Anything).Return(int64(0), sql.ErrConnDone)
			},
			expectedErr: true,
		},
		{
			name: "handles zero deleted tokens and users",
			setupMock: func(authRepo *mocks.MockAuthTokenRepository, pwRepo *mocks.MockPasswordResetRepository, emailRepo *mocks.MockEmailVerificationRepository, userRepo *mocks.MockUserRepository) {
				authRepo.EXPECT().DeleteExpiredOrRevoked(mock.Anything).Return(int64(0), nil)
				pwRepo.EXPECT().DeleteExpiredOrUsed(mock.Anything).Return(int64(0), nil)
				emailRepo.EXPECT().DeleteExpiredOrUsed(mock.Anything).Return(int64(0), nil)
				userRepo.EXPECT().DeleteScheduledUsers(mock.Anything).Return(int64(0), nil)
			},
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuthRepo := mocks.NewMockAuthTokenRepository(t)
			mockPwRepo := mocks.NewMockPasswordResetRepository(t)
			mockEmailRepo := mocks.NewMockEmailVerificationRepository(t)
			mockUserRepo := mocks.NewMockUserRepository(t)
			tt.setupMock(mockAuthRepo, mockPwRepo, mockEmailRepo, mockUserRepo)

			task := tasks.NewCleanupTask(newTestLogger(), mockAuthRepo, mockPwRepo, mockEmailRepo, mockUserRepo)

			// Create an empty asynq task (periodic tasks have empty payload)
			asynqTask := asynq.NewTask(tasks.TypeCleanup, nil)
			err := task.Handle(ctx, asynqTask)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
