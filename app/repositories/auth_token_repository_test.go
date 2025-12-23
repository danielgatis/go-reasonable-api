package repositories

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthTokenRepository(t *testing.T) {
	tx := setupTest(t)
	userRepo := NewUserRepository(testDB).WithTx(tx)
	repo := NewAuthTokenRepository(testDB).WithTx(tx)
	ctx := context.Background()

	createUser := func(t *testing.T) uuid.UUID {
		user, err := userRepo.Create(ctx, "Test User", uuid.NewString()+"@example.com", "hash")
		require.NoError(t, err)
		return user.ID
	}

	t.Run("Create", func(t *testing.T) {
		userID := createUser(t)
		expiresAt := time.Now().Add(24 * time.Hour)

		token, err := repo.Create(ctx, userID, "tokenhash123", expiresAt)
		require.NoError(t, err)
		assert.NotEmpty(t, token.ID)
		assert.Equal(t, userID, token.UserID)
		assert.Equal(t, "tokenhash123", token.TokenHash)
		assert.NotZero(t, token.CreatedAt)
	})

	t.Run("GetByHash", func(t *testing.T) {
		userID := createUser(t)
		expiresAt := time.Now().Add(24 * time.Hour)

		created, err := repo.Create(ctx, userID, "findhash123", expiresAt)
		require.NoError(t, err)

		found, err := repo.GetByHash(ctx, "findhash123")
		require.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.TokenHash, found.TokenHash)
	})

	t.Run("GetByHash_NotFound", func(t *testing.T) {
		token, err := repo.GetByHash(ctx, "nonexistenthash")
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, token)
	})

	t.Run("Revoke", func(t *testing.T) {
		userID := createUser(t)
		token, err := repo.Create(ctx, userID, "revokehash", time.Now().Add(time.Hour))
		require.NoError(t, err)
		assert.Nil(t, token.RevokedAt)

		err = repo.Revoke(ctx, token.ID)
		require.NoError(t, err)

		found, err := repo.GetByHash(ctx, "revokehash")
		require.NoError(t, err)
		assert.NotNil(t, found.RevokedAt)
	})

	t.Run("RevokeByHash", func(t *testing.T) {
		userID := createUser(t)
		_, err := repo.Create(ctx, userID, "revokebyhash", time.Now().Add(time.Hour))
		require.NoError(t, err)

		err = repo.RevokeByHash(ctx, "revokebyhash")
		require.NoError(t, err)

		found, err := repo.GetByHash(ctx, "revokebyhash")
		require.NoError(t, err)
		assert.NotNil(t, found.RevokedAt)
	})

	t.Run("RevokeAllForUser", func(t *testing.T) {
		userID := createUser(t)

		_, err := repo.Create(ctx, userID, "revokeall1", time.Now().Add(time.Hour))
		require.NoError(t, err)
		_, err = repo.Create(ctx, userID, "revokeall2", time.Now().Add(time.Hour))
		require.NoError(t, err)

		err = repo.RevokeAllForUser(ctx, userID)
		require.NoError(t, err)

		t1, _ := repo.GetByHash(ctx, "revokeall1")
		t2, _ := repo.GetByHash(ctx, "revokeall2")
		assert.NotNil(t, t1.RevokedAt)
		assert.NotNil(t, t2.RevokedAt)
	})

	t.Run("DeleteExpiredOrRevoked", func(t *testing.T) {
		userID := createUser(t)

		// Create expired token
		_, err := repo.Create(ctx, userID, "expired", time.Now().Add(-time.Hour))
		require.NoError(t, err)

		// Create revoked token
		revoked, err := repo.Create(ctx, userID, "revoked", time.Now().Add(time.Hour))
		require.NoError(t, err)
		err = repo.Revoke(ctx, revoked.ID)
		require.NoError(t, err)

		// Create valid token
		_, err = repo.Create(ctx, userID, "valid", time.Now().Add(time.Hour))
		require.NoError(t, err)

		deleted, err := repo.DeleteExpiredOrRevoked(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, deleted, int64(2))

		// Valid token should still exist
		_, err = repo.GetByHash(ctx, "valid")
		assert.NoError(t, err)

		// Expired and revoked should be deleted
		_, err = repo.GetByHash(ctx, "expired")
		require.ErrorIs(t, err, sql.ErrNoRows)
		_, err = repo.GetByHash(ctx, "revoked")
		require.ErrorIs(t, err, sql.ErrNoRows)
	})
}
