package responses

import (
	"time"

	"github.com/google/uuid"
)

type UserResponse struct {
	ID                  uuid.UUID  `json:"id"`
	Name                string     `json:"name"`
	Email               string     `json:"email"`
	EmailVerified       bool       `json:"email_verified"`
	DeletionScheduledAt *time.Time `json:"deletion_scheduled_at,omitempty"`
}
