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

func setupEcho() *echo.Echo {
	e := echo.New()
	e.Validator = zhttp.NewValidator()
	return e
}

func TestSessionHandler_Create(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		requestBody    string
		setupMock      func(*mocks.MockSessionService)
		expectedStatus int
		expectedError  string
		isBindError    bool
	}{
		{
			name:        "creates session successfully",
			requestBody: `{"email":"test@example.com","password":"password123"}`,
			setupMock: func(sessionSvc *mocks.MockSessionService) {
				sessionSvc.EXPECT().Create(mock.Anything, "test@example.com", "password123").
					Return(&sqlcgen.User{
						ID:    userID,
						Name:  "Test User",
						Email: "test@example.com",
					}, "token123", nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "returns error for invalid JSON",
			requestBody:    `{invalid}`,
			setupMock:      func(sessionSvc *mocks.MockSessionService) {},
			expectedStatus: http.StatusInternalServerError,
			isBindError:    true,
		},
		{
			name:           "returns error for missing email",
			requestBody:    `{"password":"password123"}`,
			setupMock:      func(sessionSvc *mocks.MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "returns error for invalid email",
			requestBody:    `{"email":"notanemail","password":"password123"}`,
			setupMock:      func(sessionSvc *mocks.MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "returns error for missing password",
			requestBody:    `{"email":"test@example.com"}`,
			setupMock:      func(sessionSvc *mocks.MockSessionService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:        "returns error for invalid credentials",
			requestBody: `{"email":"test@example.com","password":"wrongpassword"}`,
			setupMock: func(sessionSvc *mocks.MockSessionService) {
				sessionSvc.EXPECT().Create(mock.Anything, "test@example.com", "wrongpassword").
					Return(nil, "", apperrors.ErrInvalidCredentials)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_CREDENTIALS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupEcho()
			mockSessionSvc := mocks.NewMockSessionService(t)
			tt.setupMock(mockSessionSvc)

			handler := handlers.NewSessionHandler(mockSessionSvc)

			req := httptest.NewRequest(http.MethodPost, "/sessions", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
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
				assert.Equal(t, "token123", resp.Token)
			}
		})
	}
}

func TestSessionHandler_DeleteCurrent(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(echo.Context)
		setupMock      func(*mocks.MockSessionService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "deletes session successfully",
			setupContext: func(c echo.Context) {
				reqctx.SetToken(c, "valid-token")
			},
			setupMock: func(sessionSvc *mocks.MockSessionService) {
				sessionSvc.EXPECT().Delete(mock.Anything, "valid-token").Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "returns error when token not in context",
			setupContext:   func(c echo.Context) {},
			setupMock:      func(sessionSvc *mocks.MockSessionService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_TOKEN",
		},
		{
			name: "returns error when delete fails",
			setupContext: func(c echo.Context) {
				reqctx.SetToken(c, "valid-token")
			},
			setupMock: func(sessionSvc *mocks.MockSessionService) {
				sessionSvc.EXPECT().Delete(mock.Anything, "valid-token").
					Return(errors.InternalError("DELETE_FAILED", "failed to delete token"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "DELETE_FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupEcho()
			mockSessionSvc := mocks.NewMockSessionService(t)
			tt.setupMock(mockSessionSvc)

			handler := handlers.NewSessionHandler(mockSessionSvc)

			req := httptest.NewRequest(http.MethodDelete, "/sessions/current", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			tt.setupContext(c)

			err := handler.DeleteCurrent(c)

			if tt.expectedError != "" {
				require.Error(t, err)
				var appErr *errors.AppError
				require.ErrorAs(t, err, &appErr)
				assert.Equal(t, tt.expectedError, appErr.Code)
				assert.Equal(t, tt.expectedStatus, appErr.StatusCode)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}
		})
	}
}
