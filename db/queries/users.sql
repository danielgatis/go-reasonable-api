-- name: CreateUser :one
INSERT INTO users (id, name, email, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3;

-- name: MarkUserEmailVerified :exec
UPDATE users SET email_verified_at = $1, updated_at = $2 WHERE id = $3;

-- name: EmailExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);

-- name: ScheduleUserDeletion :exec
UPDATE users SET deletion_scheduled_at = $1, updated_at = $2 WHERE id = $3;

-- name: CancelUserDeletion :exec
UPDATE users SET deletion_scheduled_at = NULL, updated_at = $1 WHERE id = $2;

-- name: DeleteScheduledUsers :execrows
DELETE FROM users WHERE deletion_scheduled_at IS NOT NULL AND deletion_scheduled_at <= $1;
