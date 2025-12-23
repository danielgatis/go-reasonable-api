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

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupPasswordResetEcho() *echo.Echo {
	e := echo.New()
	e.Validator = zhttp.NewValidator()
	return e
}

func TestPasswordResetHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		setupMock      func(*mocks.MockPasswordResetService)
		expectedStatus int
		expectedError  string
		isBindError    bool
	}{
		{
			name:        "sends password reset email successfully",
			requestBody: `{"email":"test@example.com"}`,
			setupMock: func(pwResetSvc *mocks.MockPasswordResetService) {
				pwResetSvc.EXPECT().Create(mock.Anything, "test@example.com").Return(nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:        "returns accepted even if user not found",
			requestBody: `{"email":"notfound@example.com"}`,
			setupMock: func(pwResetSvc *mocks.MockPasswordResetService) {
				pwResetSvc.EXPECT().Create(mock.Anything, "notfound@example.com").Return(apperrors.ErrUserNotFound)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "returns error for invalid JSON",
			requestBody:    `{invalid}`,
			setupMock:      func(pwResetSvc *mocks.MockPasswordResetService) {},
			expectedStatus: http.StatusInternalServerError,
			isBindError:    true,
		},
		{
			name:           "returns error for missing email",
			requestBody:    `{}`,
			setupMock:      func(pwResetSvc *mocks.MockPasswordResetService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "returns error for invalid email",
			requestBody:    `{"email":"notanemail"}`,
			setupMock:      func(pwResetSvc *mocks.MockPasswordResetService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupPasswordResetEcho()
			mockPwResetSvc := mocks.NewMockPasswordResetService(t)
			tt.setupMock(mockPwResetSvc)

			handler := handlers.NewPasswordResetHandler(mockPwResetSvc)

			req := httptest.NewRequest(http.MethodPost, "/password-resets", strings.NewReader(tt.requestBody))
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
			}
		})
	}
}

func TestPasswordResetHandler_Update(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		requestBody    string
		setupMock      func(*mocks.MockPasswordResetService)
		expectedStatus int
		expectedError  string
		isBindError    bool
	}{
		{
			name:        "resets password successfully",
			token:       "valid-token",
			requestBody: `{"new_password":"newpassword123"}`,
			setupMock: func(pwResetSvc *mocks.MockPasswordResetService) {
				pwResetSvc.EXPECT().Execute(mock.Anything, "valid-token", "newpassword123").Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "returns error for missing token",
			token:          "",
			requestBody:    `{"new_password":"newpassword123"}`,
			setupMock:      func(pwResetSvc *mocks.MockPasswordResetService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "MISSING_TOKEN",
		},
		{
			name:           "returns error for invalid JSON",
			token:          "valid-token",
			requestBody:    `{invalid}`,
			setupMock:      func(pwResetSvc *mocks.MockPasswordResetService) {},
			expectedStatus: http.StatusInternalServerError,
			isBindError:    true,
		},
		{
			name:           "returns error for missing password",
			token:          "valid-token",
			requestBody:    `{}`,
			setupMock:      func(pwResetSvc *mocks.MockPasswordResetService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "returns error for short password",
			token:          "valid-token",
			requestBody:    `{"new_password":"short"}`,
			setupMock:      func(pwResetSvc *mocks.MockPasswordResetService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:        "returns error for invalid token",
			token:       "invalid-token",
			requestBody: `{"new_password":"newpassword123"}`,
			setupMock: func(pwResetSvc *mocks.MockPasswordResetService) {
				pwResetSvc.EXPECT().Execute(mock.Anything, "invalid-token", "newpassword123").
					Return(apperrors.ErrInvalidToken)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "INVALID_TOKEN",
		},
		{
			name:        "returns error for expired token",
			token:       "expired-token",
			requestBody: `{"new_password":"newpassword123"}`,
			setupMock: func(pwResetSvc *mocks.MockPasswordResetService) {
				pwResetSvc.EXPECT().Execute(mock.Anything, "expired-token", "newpassword123").
					Return(apperrors.ErrTokenExpired)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "TOKEN_EXPIRED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := setupPasswordResetEcho()
			mockPwResetSvc := mocks.NewMockPasswordResetService(t)
			tt.setupMock(mockPwResetSvc)

			handler := handlers.NewPasswordResetHandler(mockPwResetSvc)

			req := httptest.NewRequest(http.MethodPut, "/password-resets/"+tt.token, strings.NewReader(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("token")
			c.SetParamValues(tt.token)

			err := handler.Update(c)

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
			}
		})
	}
}
