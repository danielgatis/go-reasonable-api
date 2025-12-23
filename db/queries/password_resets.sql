-- name: CreatePasswordReset :exec
INSERT INTO password_resets (id, user_id, token_hash, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetPasswordResetByTokenHash :one
SELECT * FROM password_resets WHERE token_hash = $1;

-- name: MarkPasswordResetUsed :exec
UPDATE password_resets SET used_at = $1 WHERE id = $2;

-- name: InvalidateAllPasswordResetsForUser :exec
UPDATE password_resets SET used_at = $1 WHERE user_id = $2 AND used_at IS NULL;

-- name: DeleteExpiredOrUsedPasswordResets :execrows
DELETE FROM password_resets WHERE expires_at < $1 OR used_at IS NOT NULL;
