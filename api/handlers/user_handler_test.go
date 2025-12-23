package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-reasonable-api/api/handlers"
	"go-reasonable-api/api/responses"
	apperrors "go-reasonable-api/app/errors"
	mocks "go-reasonable-api/app/mocks/services"
	"go-reasonable-api/db/sqlcgen"
	"go-reasonable-api/support/errors"
	zhttp "go-reasonable-api/support/http"
	"go-reasonable-api/support/http/reqctx"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupUserHandlerEcho() *echo.Echo {
	e := echo.New()
	e.Validator = zhttp.NewValidator()
	return e
}

func TestUserHandler_Create(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		requestBody    string
		setupMock      func(*mocks.MockUserService, *mocks.MockSessionService)
		expectedStatus int
		expectedError  string
		isBindError    bool
	}{
		{
			name:        "creates user successfully",
			requestBody: `{"name":"Test User","email":"test@example.com","password":"password123"}`,
			setupMock: func(userSvc *mocks.MockUserService, sessionSvc *mocks.MockSessionService) {
				userSvc.EXPECT().Create(mock.Anything, "Test User", "test@example.com", "password123").
					Return(&sqlcgen.User{
						ID:    userID,
						Name:  "Test User",
						Email: "test@example.com",
					}, nil)
				sessionSvc.EXPECT().CreateForUser(mock.Anything, userID).
					Return("token123", nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "returns error for invalid JSON",
			requestBody:    `{invalid}`,
			setupMock:      func(userSvc *mocks.MockUserService, sessionSvc *mocks.MockSessionService) {},
			expectedStatus: http.StatusInternalServerError,
			isBindError:    true,
		},
		{
			name:           "returns error for missing name",
			requestBody:    `{"email":"test@example.com","password":"password123"}`,
			setupMock:      func(userSvc *mocks.MockUserService, sessionSvc *mocks.MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "returns error for missing email",
			requestBody:    `{"name":"Test User","password":"password123"}`,
			setupMock:      func(userSvc *mocks.MockUserService, sessionSvc *mocks.MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "returns error for invalid email",
			requestBody:    `{"name":"Test User","email":"notanemail","password":"password123"}`,
			setupMock:      func(userSvc *mocks.MockUserService, sessionSvc *mocks.MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "returns error for missing password",
			requestBody:    `{"name":"Test User","email":"test@example.com"}`,
			setupMock:      func(userSvc *mocks.MockUserService, sessionSvc *mocks.MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "returns error for short password",
			requestBody:    `{"name":"Test User","email":"test@example.com","password":"short"}`,
			setupMock:      func(userSvc *mocks.MockUserService, sessionSvc *mocks.MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:        "returns error when email already exists",
			requestBody: `{"name":"Test User","email":"exists@example.com","password":"password123"}`,
			setupMock: func(userSvc *mocks.MockUserService, sessionSvc *mocks.MockSessionService) {
				userSvc.EXPECT().Create(mock.Anything, "Test User", "exists@example.com", "password123").
					Return(nil, apperrors.ErrEmailAlreadyExists)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "EMAIL_ALREADY_EXISTS",
		},
		{
			name:        "returns error when session creation fails",
			requestBody: `{"name":"Test User","email":"test@example.com","password":"password123"}`,
			setupMock: func(userSvc *mocks.MockUserService, sessionSvc *mocks.MockSessionService) {
				userSvc.EXPECT().Create(mock.Anything, "Test User", "test@example.com", "password123").
					Return(&sqlcgen.User{
						ID:    userID,
						Name:  "Test User",
						Email: "test@example.com",
					}, nil)
				sessionSvc.EXPECT().CreateForUser(mock.Anything, userID).
					Return("", errors.InternalError("SESSION_CREATION_FAILED", "failed to create session"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "SESSION_CREATION_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupUserHandlerEcho()
			mockUserSvc := mocks.NewMockUserService(t)
			mockSessionSvc := mocks.NewMockSessionService(t)
			tt.setupMock(mockUserSvc, mockSessionSvc)

			handler := handlers.NewUserHandler(mockUserSvc, mockSessionSvc)

			req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.Create(c)

			if tt.isBindError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "bind")
			} else if tt.expectedError != "" {
				require.Error(t, err)
				var appErr *errors.AppError
				require.ErrorAs(t, err, &appErr)
				assert.Equal(t, tt.expectedError, appErr.Code)
				assert.Equal(t, tt.expectedStatus, appErr.StatusCode)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)

				var resp responses.SessionResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
				assert.Equal(t, userID, resp.User.ID)
				assert.Equal(t, "Test User", resp.User.Name)
				assert.Equal(t, "test@example.com", resp.User.Email)
				assert.Equal(t, "token123", resp.Token)
			}
		})
	}
}

func TestUserHandler_Me(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		setupContext   func(c echo.Context)
		setupMock      func(*mocks.MockUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "returns user successfully",
			setupContext: func(c echo.Context) {
				reqctx.SetUserID(c, userID)
			},
			setupMock: func(userSvc *mocks.MockUserService) {
				userSvc.EXPECT().GetByID(mock.Anything, userID).
					Return(&sqlcgen.User{
						ID:    userID,
						Name:  "Test User",
						Email: "test@example.com",
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "returns error when user not in context",
			setupContext:   func(c echo.Context) {},
			setupMock:      func(userSvc *mocks.MockUserService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_TOKEN",
		},
		{
			name: "returns error when user not found",
			setupContext: func(c echo.Context) {
				reqctx.SetUserID(c, userID)
			},
			setupMock: func(userSvc *mocks.MockUserService) {
				userSvc.EXPECT().GetByID(mock.Anything, userID).
					Return(nil, apperrors.ErrUserNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "USER_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupUserHandlerEcho()
			mockUserSvc := mocks.NewMockUserService(t)
			mockSessionSvc := mocks.NewMockSessionService(t)
			tt.setupMock(mockUserSvc)

			handler := handlers.NewUserHandler(mockUserSvc, mockSessionSvc)

			req := httptest.NewRequest(http.MethodGet, "/users/me", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			tt.setupContext(c)

			err := handler.Me(c)

			if tt.expectedError != "" {
				require.Error(t, err)
				var appErr *errors.AppError
				require.ErrorAs(t, err, &appErr)
				assert.Equal(t, tt.expectedError, appErr.Code)
				assert.Equal(t, tt.expectedStatus, appErr.StatusCode)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)

				var resp responses.UserResponse
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
				assert.Equal(t, userID, resp.ID)
				assert.Equal(t, "Test User", resp.Name)
				assert.Equal(t, "test@example.com", resp.Email)
			}
		})
	}
}
