-- name: CreateAuthToken :exec
INSERT INTO auth_tokens (id, user_id, token_hash, expires_at, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetAuthTokenByHash :one
SELECT * FROM auth_tokens WHERE token_hash = $1;

-- name: RevokeAuthToken :exec
UPDATE auth_tokens SET revoked_at = $1 WHERE id = $2;

-- name: RevokeAuthTokenByHash :exec
UPDATE auth_tokens SET revoked_at = $1 WHERE token_hash = $2;

-- name: RevokeAllAuthTokensForUser :exec
UPDATE auth_tokens SET revoked_at = $1 WHERE user_id = $2 AND revoked_at IS NULL;

-- name: DeleteExpiredOrRevokedAuthTokens :execrows
DELETE FROM auth_tokens WHERE expires_at < $1 OR revoked_at IS NOT NULL;
