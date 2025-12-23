package services_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"go-reasonable-api/app/errors"
	mocks "go-reasonable-api/app/mocks/repositories"
	mocksSupport "go-reasonable-api/app/mocks/support"
	"go-reasonable-api/app/services"
	"go-reasonable-api/db/sqlcgen"
	"go-reasonable-api/support/config"
	"go-reasonable-api/support/db"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func newTestConfig() *config.Config {
	return &config.Config{
		Auth: config.AuthConfig{
			BcryptCost:           bcrypt.MinCost,
			AccountDeletionDelay: 720 * time.Hour, // 30 days
		},
		Worker: config.WorkerConfig{
			EmailMaxRetry:  5,
			EmailTimeout:   30 * time.Second,
			EmailRetention: 24 * time.Hour,
		},
	}
}

func TestUserService_Create(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		email       string
		password    string
		userName    string
		setupMock   func(*mocks.MockUserRepository)
		expectedErr error
	}{
		{
			name:     "creates user successfully",
			email:    "test@example.com",
			password: "password123",
			userName: "Test User",
			setupMock: func(m *mocks.MockUserRepository) {
				m.EXPECT().EmailExists(mock.Anything, "test@example.com").Return(false, nil)
				m.EXPECT().Create(mock.Anything, "Test User", "test@example.com", mock.AnythingOfType("string")).
					Return(&sqlcgen.User{
						ID:    uuid.New(),
						Name:  "Test User",
						Email: "test@example.com",
					}, nil)
			},
			expectedErr: nil,
		},
		{
			name:     "returns error when email already exists",
			email:    "existing@example.com",
			password: "password123",
			userName: "Test User",
			setupMock: func(m *mocks.MockUserRepository) {
				m.EXPECT().EmailExists(mock.Anything, "existing@example.com").Return(true, nil)
			},
			expectedErr: errors.ErrEmailAlreadyExists,
		},
		{
			name:     "returns error when email check fails",
			email:    "test@example.com",
			password: "password123",
			userName: "Test User",
			setupMock: func(m *mocks.MockUserRepository) {
				m.EXPECT().EmailExists(mock.Anything, "test@example.com").Return(false, sql.ErrConnDone)
			},
			expectedErr: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockUserRepository(t)
			mockAuthTokenRepo := mocks.NewMockAuthTokenRepository(t)
			tt.setupMock(mockRepo)

			service := services.NewUserService(newTestConfig(), nil, mockRepo, mockAuthTokenRepo, nil)
			user, err := service.Create(ctx, tt.userName, tt.email, tt.password)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
			}
		})
	}
}

func TestUserService_GetByID(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name        string
		userID      uuid.UUID
		setupMock   func(*mocks.MockUserRepository)
		expectedErr error
	}{
		{
			name:   "returns user when found",
			userID: userID,
			setupMock: func(m *mocks.MockUserRepository) {
				m.EXPECT().GetByID(mock.Anything, userID).Return(&sqlcgen.User{
					ID:    userID,
					Email: "test@example.com",
				}, nil)
			},
			expectedErr: nil,
		},
		{
			name:   "returns ErrUserNotFound when user does not exist",
			userID: userID,
			setupMock: func(m *mocks.MockUserRepository) {
				m.EXPECT().GetByID(mock.Anything, userID).Return(nil, sql.ErrNoRows)
			},
			expectedErr: errors.ErrUserNotFound,
		},
		{
			name:   "returns error when repository fails",
			userID: userID,
			setupMock: func(m *mocks.MockUserRepository) {
				m.EXPECT().GetByID(mock.Anything, userID).Return(nil, sql.ErrConnDone)
			},
			expectedErr: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockUserRepository(t)
			mockAuthTokenRepo := mocks.NewMockAuthTokenRepository(t)
			tt.setupMock(mockRepo)

			service := services.NewUserService(newTestConfig(), nil, mockRepo, mockAuthTokenRepo, nil)
			user, err := service.GetByID(ctx, tt.userID)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.userID, user.ID)
			}
		})
	}
}

func TestUserService_GetByEmail(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		email       string
		setupMock   func(*mocks.MockUserRepository)
		expectedErr error
	}{
		{
			name:  "returns user when found",
			email: "test@example.com",
			setupMock: func(m *mocks.MockUserRepository) {
				m.EXPECT().GetByEmail(mock.Anything, "test@example.com").Return(&sqlcgen.User{
					ID:    uuid.New(),
					Email: "test@example.com",
				}, nil)
			},
			expectedErr: nil,
		},
		{
			name:  "returns ErrUserNotFound when user does not exist",
			email: "notfound@example.com",
			setupMock: func(m *mocks.MockUserRepository) {
				m.EXPECT().GetByEmail(mock.Anything, "notfound@example.com").Return(nil, sql.ErrNoRows)
			},
			expectedErr: errors.ErrUserNotFound,
		},
		{
			name:  "returns error when repository fails",
			email: "test@example.com",
			setupMock: func(m *mocks.MockUserRepository) {
				m.EXPECT().GetByEmail(mock.Anything, "test@example.com").Return(nil, sql.ErrConnDone)
			},
			expectedErr: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mocks.NewMockUserRepository(t)
			mockAuthTokenRepo := mocks.NewMockAuthTokenRepository(t)
			tt.setupMock(mockRepo)

			service := services.NewUserService(newTestConfig(), nil, mockRepo, mockAuthTokenRepo, nil)
			user, err := service.GetByEmail(ctx, tt.email)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.email, user.Email)
			}
		})
	}
}

func TestUserService_ScheduleDeletion(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("returns error when user not found", func(t *testing.T) {
		mockDB, mockSQL, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		mockSQL.ExpectBegin()
		mockSQL.ExpectRollback()

		txManager := db.NewTxManager(mockDB)
		mockRepo := mocks.NewMockUserRepository(t)
		mockAuthTokenRepo := mocks.NewMockAuthTokenRepository(t)
		mockTaskClient := mocksSupport.NewMockTaskClient(t)

		// WithTx returns itself for chaining
		mockRepo.EXPECT().WithTx(mock.Anything).Return(mockRepo)
		mockAuthTokenRepo.EXPECT().WithTx(mock.Anything).Return(mockAuthTokenRepo)
		mockRepo.EXPECT().GetByID(mock.Anything, userID).Return(nil, sql.ErrNoRows)

		service := services.NewUserService(newTestConfig(), txManager, mockRepo, mockAuthTokenRepo, mockTaskClient)
		err = service.ScheduleDeletion(ctx, userID)

		assert.ErrorIs(t, err, errors.ErrUserNotFound)
	})

	t.Run("returns error when deletion already scheduled", func(t *testing.T) {
		mockDB, mockSQL, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		mockSQL.ExpectBegin()
		mockSQL.ExpectRollback()

		txManager := db.NewTxManager(mockDB)
		mockRepo := mocks.NewMockUserRepository(t)
		mockAuthTokenRepo := mocks.NewMockAuthTokenRepository(t)
		mockTaskClient := mocksSupport.NewMockTaskClient(t)

		scheduledAt := time.Now().Add(24 * time.Hour)
		mockRepo.EXPECT().WithTx(mock.Anything).Return(mockRepo)
		mockAuthTokenRepo.EXPECT().WithTx(mock.Anything).Return(mockAuthTokenRepo)
		mockRepo.EXPECT().GetByID(mock.Anything, userID).Return(&sqlcgen.User{
			ID:                  userID,
			Email:               "test@example.com",
			DeletionScheduledAt: &scheduledAt,
		}, nil)

		service := services.NewUserService(newTestConfig(), txManager, mockRepo, mockAuthTokenRepo, mockTaskClient)
		err = service.ScheduleDeletion(ctx, userID)

		assert.ErrorIs(t, err, errors.ErrDeletionAlreadyScheduled)
	})

	t.Run("schedules deletion successfully", func(t *testing.T) {
		mockDB, mockSQL, err := sqlmock.New()
		require.NoError(t, err)
		defer func() { _ = mockDB.Close() }()

		mockSQL.ExpectBegin()
		mockSQL.ExpectCommit()

		txManager := db.NewTxManager(mockDB)
		mockRepo := mocks.NewMockUserRepository(t)
		mockAuthTokenRepo := mocks.NewMockAuthTokenRepository(t)
		mockTaskClient := mocksSupport.NewMockTaskClient(t)

		mockRepo.EXPECT().WithTx(mock.Anything).Return(mockRepo)
		mockAuthTokenRepo.EXPECT().WithTx(mock.Anything).Return(mockAuthTokenRepo)
		mockRepo.EXPECT().GetByID(mock.Anything, userID).Return(&sqlcgen.User{
			ID:                  userID,
			Name:                "Test User",
			Email:               "test@example.com",
			DeletionScheduledAt: nil,
		}, nil)
		mockRepo.EXPECT().ScheduleDeletion(mock.Anything, userID, mock.AnythingOfType("time.Time")).Return(nil)
		mockAuthTokenRepo.EXPECT().RevokeAllForUser(mock.Anything, userID).Return(nil)
		mockTaskClient.EXPECT().EnqueueCtx(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		service := services.NewUserService(newTestConfig(), txManager, mockRepo, mockAuthTokenRepo, mockTaskClient)
		err = service.ScheduleDeletion(ctx, userID)

		require.NoError(t, err)
	})
}
