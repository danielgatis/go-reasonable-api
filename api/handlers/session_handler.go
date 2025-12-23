package handlers

import (
	"net/http"

	"go-reasonable-api/api/requests"
	"go-reasonable-api/api/responses"
	apperrors "go-reasonable-api/app/errors"
	"go-reasonable-api/app/interfaces/services"
	"go-reasonable-api/support/http/bind"
	"go-reasonable-api/support/http/reqctx"

	"github.com/labstack/echo/v4"
	"github.com/rotisserie/eris"
)

// SessionHandler handles authentication (login/logout).
type SessionHandler struct {
	sessionService services.SessionService
}

func NewSessionHandler(sessionService services.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

// Create authenticates a user and returns a token
// @Summary Create session (login)
// @Description Authenticate user with email and password
// @Tags sessions
// @Accept json
// @Produce json
// @Param request body requests.CreateSessionRequest true "Create session request"
// @Success 201 {object} responses.SessionResponse
// @Failure 400 {object} errors.AppError
// @Failure 401 {object} errors.AppError
// @Router /sessions [post]
func (h *SessionHandler) Create(c echo.Context) error {
	var req requests.CreateSessionRequest
	if err := bind.AndValidate(c, &req); err != nil {
		return err
	}

	user, token, err := h.sessionService.Create(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return eris.Wrap(err, "failed to create session")
	}

	return c.JSON(http.StatusCreated, responses.SessionResponse{
		User: responses.UserResponse{
			ID:            user.ID,
			Name:          user.Name,
			Email:         user.Email,
			EmailVerified: user.EmailVerifiedAt != nil,
		},
		Token: token,
	})
}

// DeleteCurrent invalidates the current user's token
// @Summary Delete current session (logout)
// @Description Invalidate the current user's authentication token
// @Tags sessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 204
// @Failure 401 {object} errors.AppError
// @Router /sessions/current [delete]
func (h *SessionHandler) DeleteCurrent(c echo.Context) error {
	token, ok := reqctx.GetToken(c)
	if !ok {
		return apperrors.ErrInvalidToken
	}

	if err := h.sessionService.Delete(c.Request().Context(), token); err != nil {
		return eris.Wrap(err, "failed to delete session")
	}

	return c.NoContent(http.StatusNoContent)
}
