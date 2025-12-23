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

// UserHandler handles user registration and profile operations.
type UserHandler struct {
	userService    services.UserService
	sessionService services.SessionService
}

func NewUserHandler(userService services.UserService, sessionService services.SessionService) *UserHandler {
	return &UserHandler{
		userService:    userService,
		sessionService: sessionService,
	}
}

// Create creates a new user account
// @Summary Create new user account
// @Description Register a new user with name, email and password
// @Tags users
// @Accept json
// @Produce json
// @Param request body requests.CreateUserRequest true "Create user request"
// @Success 201 {object} responses.SessionResponse
// @Failure 400 {object} errors.AppError
// @Failure 422 {object} errors.AppError
// @Router /users [post]
func (h *UserHandler) Create(c echo.Context) error {
	var req requests.CreateUserRequest
	if err := bind.AndValidate(c, &req); err != nil {
		return err
	}

	user, err := h.userService.Create(c.Request().Context(), req.Name, req.Email, req.Password)
	if err != nil {
		return eris.Wrap(err, "failed to create user")
	}

	token, err := h.sessionService.CreateForUser(c.Request().Context(), user.ID)
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

// Me returns the current authenticated user
// @Summary Get current user
// @Description Get the currently authenticated user's information
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} responses.UserResponse
// @Failure 401 {object} errors.AppError
// @Router /users/me [get]
func (h *UserHandler) Me(c echo.Context) error {
	userID, ok := reqctx.GetUserID(c)
	if !ok {
		return apperrors.ErrInvalidToken
	}

	user, err := h.userService.GetByID(c.Request().Context(), userID)
	if err != nil {
		return eris.Wrap(err, "failed to get user")
	}

	return c.JSON(http.StatusOK, responses.UserResponse{
		ID:                  user.ID,
		Name:                user.Name,
		Email:               user.Email,
		EmailVerified:       user.EmailVerifiedAt != nil,
		DeletionScheduledAt: user.DeletionScheduledAt,
	})
}

// Delete schedules the current user's account for deletion
// @Summary Schedule account deletion
// @Description Schedule the current user's account for deletion after 30 days. All sessions will be revoked.
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} errors.AppError
// @Failure 422 {object} errors.AppError
// @Router /users/me [delete]
func (h *UserHandler) Delete(c echo.Context) error {
	userID, ok := reqctx.GetUserID(c)
	if !ok {
		return apperrors.ErrInvalidToken
	}

	if err := h.userService.ScheduleDeletion(c.Request().Context(), userID); err != nil {
		return eris.Wrap(err, "failed to schedule deletion")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "account deletion scheduled",
	})
}
