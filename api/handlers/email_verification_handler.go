package handlers

import (
	"net/http"

	"go-reasonable-api/api/requests"
	"go-reasonable-api/app/interfaces/services"
	"go-reasonable-api/support/http/bind"
	"go-reasonable-api/support/http/reqctx"
	"go-reasonable-api/support/logger"

	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

type EmailVerificationHandler struct {
	emailVerificationService services.EmailVerificationService
}

func NewEmailVerificationHandler(emailVerificationService services.EmailVerificationService) *EmailVerificationHandler {
	return &EmailVerificationHandler{
		emailVerificationService: emailVerificationService,
	}
}

// Create sends an email verification
// @Summary Request email verification
// @Description Send an email verification link. If authenticated, sends to current user. If not, requires email in body.
// @Tags email-verifications
// @Accept json
// @Produce json
// @Param request body requests.CreateEmailVerificationRequest false "Email (required if not authenticated)"
// @Success 202
// @Failure 400 {object} errors.AppError
// @Router /email-verifications [post]
func (h *EmailVerificationHandler) Create(c echo.Context) error {
	if userID, ok := reqctx.GetUserID(c); ok {
		if err := h.emailVerificationService.Send(c.Request().Context(), userID); err != nil {
			return eris.Wrap(err, "failed to send email verification")
		}
		return c.NoContent(http.StatusAccepted)
	}

	var req requests.CreateEmailVerificationRequest
	if err := bind.AndValidate(c, &req); err != nil {
		return err
	}

	// Log error but don't return it to prevent user enumeration
	if err := h.emailVerificationService.Resend(c.Request().Context(), req.Email); err != nil {
		logger.Ctx(c.Request().Context()).Warn().
			Err(err).
			Str("email", req.Email).
			Msg("failed to resend email verification")
	}

	return c.NoContent(http.StatusAccepted)
}

// Update verifies the user's email using the token
// @Summary Verify email
// @Description Verify user email using the token from email
// @Tags email-verifications
// @Accept json
// @Produce json
// @Param token path string true "Email verification token"
// @Success 204
// @Failure 400 {object} errors.AppError
// @Failure 422 {object} errors.AppError
// @Router /email-verifications/{token} [put]
func (h *EmailVerificationHandler) Update(c echo.Context) error {
	token, err := bind.RequiredParam(c, "token")
	if err != nil {
		return err
	}

	if err := h.emailVerificationService.Verify(c.Request().Context(), token); err != nil {
		return eris.Wrap(err, "failed to verify email")
	}

	return c.NoContent(http.StatusNoContent)
}
