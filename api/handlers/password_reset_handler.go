package handlers

import (
	"net/http"

	"go-reasonable-api/api/requests"
	"go-reasonable-api/app/interfaces/services"
	"go-reasonable-api/support/http/bind"
	"go-reasonable-api/support/logger"

	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

type PasswordResetHandler struct {
	passwordResetService services.PasswordResetService
}

func NewPasswordResetHandler(passwordResetService services.PasswordResetService) *PasswordResetHandler {
	return &PasswordResetHandler{
		passwordResetService: passwordResetService,
	}
}

// Create sends a password reset email
// @Summary Request password reset
// @Description Send a password reset email to the user
// @Tags password-resets
// @Accept json
// @Produce json
// @Param request body requests.CreatePasswordResetRequest true "Create password reset request"
// @Success 202
// @Failure 400 {object} errors.AppError
// @Router /password-resets [post]
func (h *PasswordResetHandler) Create(c echo.Context) error {
	var req requests.CreatePasswordResetRequest
	if err := bind.AndValidate(c, &req); err != nil {
		return err
	}

	// Log error but don't return it to prevent user enumeration
	if err := h.passwordResetService.Create(c.Request().Context(), req.Email); err != nil {
		logger.Ctx(c.Request().Context()).Warn().
			Err(err).
			Str("email", req.Email).
			Msg("failed to create password reset")
	}

	return c.NoContent(http.StatusAccepted)
}

// Update resets the user's password using the token
// @Summary Reset password
// @Description Reset user password using the token from email
// @Tags password-resets
// @Accept json
// @Produce json
// @Param token path string true "Password reset token"
// @Param request body requests.UpdatePasswordResetRequest true "Update password reset request"
// @Success 204
// @Failure 400 {object} errors.AppError
// @Failure 422 {object} errors.AppError
// @Router /password-resets/{token} [put]
func (h *PasswordResetHandler) Update(c echo.Context) error {
	token, err := bind.RequiredParam(c, "token")
	if err != nil {
		return err
	}

	var req requests.UpdatePasswordResetRequest
	if err := bind.AndValidate(c, &req); err != nil {
		return err
	}

	if err := h.passwordResetService.Execute(c.Request().Context(), token, req.NewPassword); err != nil {
		return eris.Wrap(err, "failed to reset password")
	}

	return c.NoContent(http.StatusNoContent)
}
