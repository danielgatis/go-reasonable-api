package services_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"go-reasonable-api/app/errors"
	mocks "go-reasonable-api/app/mocks/repositories"
	"go-reasonable-api/app/services"
	"go-reasonable-api/db/sqlcgen"
	"go-reasonable-api/support/config"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func newSessionTestConfig() *config.Config {
	return &config.Config{
		Auth: config.AuthConfig{
			BcryptCost:   bcrypt.MinCost,
			AuthTokenTTL: time.Hour,
		},
	}
}

func TestSessionService_Create(t *testing.T) {
	ctx := context.Background()
	password := "password123"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	userID := uuid.New()

	tests := []struct {
		name        string
		email       string
		password    string
		setupMock   func(*mocks.MockUserRepository, *mocks.MockAuthTokenRepository)
		expectUser  bool
		expectToken bool
		expectedErr error
	}{
		{
			name:     "creates session successfully",
			email:    "test@example.com",
			password: password,
			setupMock: func(userRepo *mocks.MockUserRepository, authRepo *mocks.MockAuthTokenRepository) {
				userRepo.EXPECT().GetByEmail(mock.Anything, "test@example.com").Return(&sqlcgen.User{
					ID:           userID,
					Email:        "test@example.com",
					PasswordHash: string(passwordHash),
				}, nil)
				authRepo.EXPECT().Create(mock.Anything, userID, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
					Return(&sqlcgen.AuthToken{ID: uuid.New()}, nil)
			},
			expectUser:  true,
			expectToken: true,
			expectedErr: nil,
		},
		{
			name:     "returns error when user not found",
			email:    "notfound@example.com",
			password: password,
			setupMock: func(userRepo *mocks.MockUserRepository, authRepo *mocks.MockAuthTokenRepository) {
				userRepo.EXPECT().GetByEmail(mock.Anything, "notfound@example.com").Return(nil, sql.ErrNoRows)
			},
			expectUser:  false,
			expectToken: false,
			expectedErr: errors.ErrInvalidCredentials,
		},
		{
			name:     "returns error when password is wrong",
			email:    "test@example.com",
			password: "wrongpassword",
			setupMock: func(userRepo *mocks.MockUserRepository, authRepo *mocks.MockAuthTokenRepository) {
				userRepo.EXPECT().GetByEmail(mock.Anything, "test@example.com").Return(&sqlcgen.User{
					ID:           userID,
					Email:        "test@example.com",
					PasswordHash: string(passwordHash),
				}, nil)
			},
			expectUser:  false,
			expectToken: false,
			expectedErr: errors.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := mocks.NewMockUserRepository(t)
			mockAuthRepo := mocks.NewMockAuthTokenRepository(t)
			tt.setupMock(mockUserRepo, mockAuthRepo)

			service := services.NewSessionService(newSessionTestConfig(), mockUserRepo, mockAuthRepo)
			user, token, err := service.Create(ctx, tt.email, tt.password)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, user)
				assert.Empty(t, token)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, user)
				assert.NotEmpty(t, token)
			}
		})
	}
}

func TestSessionService_ValidateToken(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	tokenID := uuid.New()

	tests := []struct {
		name        string
		token       string
		setupMock   func(*mocks.MockAuthTokenRepository)
		expectedErr error
	}{
		{
			name:  "validates token successfully",
			token: "valid-token",
			setupMock: func(authRepo *mocks.MockAuthTokenRepository) {
				authRepo.EXPECT().GetByHash(mock.Anything, mock.AnythingOfType("string")).Return(&sqlcgen.AuthToken{
					ID:        tokenID,
					UserID:    userID,
					ExpiresAt: time.Now().UTC().Add(time.Hour),
					RevokedAt: nil,
				}, nil)
			},
			expectedErr: nil,
		},
		{
			name:  "returns error when token not found",
			token: "invalid-token",
			setupMock: func(authRepo *mocks.MockAuthTokenRepository) {
				authRepo.EXPECT().GetByHash(mock.Anything, mock.AnythingOfType("string")).Return(nil, sql.ErrNoRows)
			},
			expectedErr: errors.ErrInvalidToken,
		},
		{
			name:  "returns error when token is revoked",
			token: "revoked-token",
			setupMock: func(authRepo *mocks.MockAuthTokenRepository) {
				revokedAt := time.Now().UTC()
				authRepo.EXPECT().GetByHash(mock.Anything, mock.AnythingOfType("string")).Return(&sqlcgen.AuthToken{
					ID:        tokenID,
					UserID:    userID,
					ExpiresAt: time.Now().UTC().Add(time.Hour),
					RevokedAt: &revokedAt,
				}, nil)
			},
			expectedErr: errors.ErrTokenRevoked,
		},
		{
			name:  "returns error when token is expired",
			token: "expired-token",
			setupMock: func(authRepo *mocks.MockAuthTokenRepository) {
				authRepo.EXPECT().GetByHash(mock.Anything, mock.AnythingOfType("string")).Return(&sqlcgen.AuthToken{
					ID:        tokenID,
					UserID:    userID,
					ExpiresAt: time.Now().UTC().Add(-time.Hour),
					RevokedAt: nil,
				}, nil)
			},
			expectedErr: errors.ErrTokenExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := mocks.NewMockUserRepository(t)
			mockAuthRepo := mocks.NewMockAuthTokenRepository(t)
			tt.setupMock(mockAuthRepo)

			service := services.NewSessionService(newSessionTestConfig(), mockUserRepo, mockAuthRepo)
			authToken, err := service.ValidateToken(ctx, tt.token)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, authToken)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, authToken)
			}
		})
	}
}

func TestSessionService_Delete(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		token       string
		setupMock   func(*mocks.MockAuthTokenRepository)
		expectedErr error
	}{
		{
			name:  "deletes session successfully",
			token: "valid-token",
			setupMock: func(authRepo *mocks.MockAuthTokenRepository) {
				authRepo.EXPECT().RevokeByHash(mock.Anything, mock.AnythingOfType("string")).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name:  "returns error when revoke fails",
			token: "valid-token",
			setupMock: func(authRepo *mocks.MockAuthTokenRepository) {
				authRepo.EXPECT().RevokeByHash(mock.Anything, mock.AnythingOfType("string")).Return(sql.ErrConnDone)
			},
			expectedErr: sql.ErrConnDone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := mocks.NewMockUserRepository(t)
			mockAuthRepo := mocks.NewMockAuthTokenRepository(t)
			tt.setupMock(mockAuthRepo)

			service := services.NewSessionService(newSessionTestConfig(), mockUserRepo, mockAuthRepo)
			err := service.Delete(ctx, tt.token)

			if tt.expectedErr != nil {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
