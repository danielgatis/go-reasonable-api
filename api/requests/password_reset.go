package requests

type CreatePasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type UpdatePasswordResetRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=8"`
}
