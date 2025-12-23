package requests

type CreateEmailVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}
