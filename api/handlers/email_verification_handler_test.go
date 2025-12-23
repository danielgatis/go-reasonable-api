package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-reasonable-api/api/handlers"
	apperrors "go-reasonable-api/app/errors"
	mocks "go-reasonable-api/app/mocks/services"
	"go-reasonable-api/support/errors"
	zhttp "go-reasonable-api/support/http"
	"go-reasonable-api/support/http/reqctx"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupEmailVerificationEcho() *echo.Echo {
	e := echo.New()
	e.Validator = zhttp.NewValidator()
	return e
}

func TestEmailVerificationHandler_Create(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name           string
		requestBody    string
		setupContext   func(c echo.Context)
		setupMock      func(*mocks.MockEmailVerificationService)
		expectedStatus int
		expectedError  string
		isBindError    bool
	}{
		{
			name:        "sends verification email for authenticated user",
			requestBody: ``,
			setupContext: func(c echo.Context) {
				reqctx.SetUserID(c, userID)
			},
			setupMock: func(emailVerifySvc *mocks.MockEmailVerificationService) {
				emailVerifySvc.EXPECT().Send(mock.Anything, userID).Return(nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:        "returns error when send fails for authenticated user",
			requestBody: ``,
			setupContext: func(c echo.Context) {
				reqctx.SetUserID(c, userID)
			},
			setupMock: func(emailVerifySvc *mocks.MockEmailVerificationService) {
				emailVerifySvc.EXPECT().Send(mock.Anything, userID).
					Return(errors.InternalError("SEND_FAILED", "failed to send email"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "SEND_FAILED",
		},
		{
			name:         "resends verification email for unauthenticated user",
			requestBody:  `{"email":"test@example.com"}`,
			setupContext: func(c echo.Context) {},
			setupMock: func(emailVerifySvc *mocks.MockEmailVerificationService) {
				emailVerifySvc.EXPECT().Resend(mock.Anything, "test@example.com").Return(nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:         "returns accepted even if user not found for resend",
			requestBody:  `{"email":"notfound@example.com"}`,
			setupContext: func(c echo.Context) {},
			setupMock: func(emailVerifySvc *mocks.MockEmailVerificationService) {
				emailVerifySvc.EXPECT().Resend(mock.Anything, "notfound@example.com").Return(apperrors.ErrUserNotFound)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "returns error for invalid JSON",
			requestBody:    `{invalid}`,
			setupContext:   func(c echo.Context) {},
			setupMock:      func(emailVerifySvc *mocks.MockEmailVerificationService) {},
			expectedStatus: http.StatusInternalServerError,
			isBindError:    true,
		},
		{
			name:           "returns error for missing email when unauthenticated",
			requestBody:    `{}`,
			setupContext:   func(c echo.Context) {},
			setupMock:      func(emailVerifySvc *mocks.MockEmailVerificationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "returns error for invalid email when unauthenticated",
			requestBody:    `{"email":"notanemail"}`,
			setupContext:   func(c echo.Context) {},
			setupMock:      func(emailVerifySvc *mocks.MockEmailVerificationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupEmailVerificationEcho()
			mockEmailVerifySvc := mocks.NewMockEmailVerificationService(t)
			tt.setupMock(mockEmailVerifySvc)

			handler := handlers.NewEmailVerificationHandler(mockEmailVerifySvc)

			req := httptest.NewRequest(http.MethodPost, "/email-verifications", strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			tt.setupContext(c)

			err := handler.Create(c)

			if tt.isBindError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "bind")
			} else if tt.expectedError != "" {
				require.Error(t, err)
				var appErr *errors.AppError
				require.True(t, eris.As(err, &appErr))
				assert.Equal(t, tt.expectedError, appErr.Code)
				assert.Equal(t, tt.expectedStatus, appErr.StatusCode)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestEmailVerificationHandler_Update(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		setupMock      func(*mocks.MockEmailVerificationService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:  "verifies email successfully",
			token: "valid-token",
			setupMock: func(emailVerifySvc *mocks.MockEmailVerificationService) {
				emailVerifySvc.EXPECT().Verify(mock.Anything, "valid-token").Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "returns error for missing token",
			token:          "",
			setupMock:      func(emailVerifySvc *mocks.MockEmailVerificationService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_TOKEN",
		},
		{
			name:  "returns error for invalid token",
			token: "invalid-token",
			setupMock: func(emailVerifySvc *mocks.MockEmailVerificationService) {
				emailVerifySvc.EXPECT().Verify(mock.Anything, "invalid-token").
					Return(apperrors.ErrInvalidToken)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_TOKEN",
		},
		{
			name:  "returns error for expired token",
			token: "expired-token",
			setupMock: func(emailVerifySvc *mocks.MockEmailVerificationService) {
				emailVerifySvc.EXPECT().Verify(mock.Anything, "expired-token").
					Return(apperrors.ErrTokenExpired)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "TOKEN_EXPIRED",
		},
		{
			name:  "returns error for already verified email",
			token: "used-token",
			setupMock: func(emailVerifySvc *mocks.MockEmailVerificationService) {
				emailVerifySvc.EXPECT().Verify(mock.Anything, "used-token").
					Return(apperrors.ErrEmailAlreadyVerified)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "EMAIL_ALREADY_VERIFIED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupEmailVerificationEcho()
			mockEmailVerifySvc := mocks.NewMockEmailVerificationService(t)
			tt.setupMock(mockEmailVerifySvc)

			handler := handlers.NewEmailVerificationHandler(mockEmailVerifySvc)

			req := httptest.NewRequest(http.MethodPut, "/email-verifications/"+tt.token, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("token")
			c.SetParamValues(tt.token)

			err := handler.Update(c)

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
