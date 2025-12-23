-- name: CreateEmailVerification :exec
INSERT INTO email_verifications (id, user_id, token_hash, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetEmailVerificationByTokenHash :one
SELECT * FROM email_verifications WHERE token_hash = $1;

-- name: MarkEmailVerificationUsed :exec
UPDATE email_verifications SET used_at = $1 WHERE id = $2;

-- name: InvalidateAllEmailVerificationsForUser :exec
UPDATE email_verifications SET used_at = $1 WHERE user_id = $2 AND used_at IS NULL;

-- name: DeleteExpiredOrUsedEmailVerifications :execrows
DELETE FROM email_verifications WHERE expires_at < $1 OR used_at IS NOT NULL;
