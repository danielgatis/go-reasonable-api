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

func TestEmailVerificationRepository(t *testing.T) {
	tx := setupTest(t)
	userRepo := NewUserRepository(testDB).WithTx(tx)
	repo := NewEmailVerificationRepository(testDB).WithTx(tx)
	ctx := context.Background()

	createUser := func(t *testing.T) uuid.UUID {
		user, err := userRepo.Create(ctx, "Test User", uuid.NewString()+"@example.com", "hash")
		require.NoError(t, err)
		return user.ID
	}

	t.Run("Create", func(t *testing.T) {
		userID := createUser(t)
		expiresAt := time.Now().Add(24 * time.Hour)

		verification, err := repo.Create(ctx, userID, "verifyhash123", expiresAt)
		require.NoError(t, err)
		assert.NotEmpty(t, verification.ID)
		assert.Equal(t, userID, verification.UserID)
		assert.Equal(t, "verifyhash123", verification.TokenHash)
		assert.Nil(t, verification.UsedAt)
		assert.NotZero(t, verification.CreatedAt)
	})

	t.Run("GetByTokenHash", func(t *testing.T) {
		userID := createUser(t)
		expiresAt := time.Now().Add(24 * time.Hour)

		created, err := repo.Create(ctx, userID, "findverifyhash", expiresAt)
		require.NoError(t, err)

		found, err := repo.GetByTokenHash(ctx, "findverifyhash")
		require.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.TokenHash, found.TokenHash)
	})

	t.Run("GetByTokenHash_NotFound", func(t *testing.T) {
		verification, err := repo.GetByTokenHash(ctx, "nonexistenthash")
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, verification)
	})

	t.Run("MarkUsed", func(t *testing.T) {
		userID := createUser(t)
		verification, err := repo.Create(ctx, userID, "markusedhash", time.Now().Add(time.Hour))
		require.NoError(t, err)
		assert.Nil(t, verification.UsedAt)

		err = repo.MarkUsed(ctx, verification.ID)
		require.NoError(t, err)

		found, err := repo.GetByTokenHash(ctx, "markusedhash")
		require.NoError(t, err)
		assert.NotNil(t, found.UsedAt)
	})

	t.Run("InvalidateAllForUser", func(t *testing.T) {
		userID := createUser(t)

		_, err := repo.Create(ctx, userID, "invalidate1", time.Now().Add(time.Hour))
		require.NoError(t, err)
		_, err = repo.Create(ctx, userID, "invalidate2", time.Now().Add(time.Hour))
		require.NoError(t, err)

		err = repo.InvalidateAllForUser(ctx, userID)
		require.NoError(t, err)

		v1, _ := repo.GetByTokenHash(ctx, "invalidate1")
		v2, _ := repo.GetByTokenHash(ctx, "invalidate2")
		assert.NotNil(t, v1.UsedAt)
		assert.NotNil(t, v2.UsedAt)
	})

	t.Run("DeleteExpiredOrUsed", func(t *testing.T) {
		userID := createUser(t)

		// Create expired verification
		_, err := repo.Create(ctx, userID, "expiredverify", time.Now().Add(-time.Hour))
		require.NoError(t, err)

		// Create used verification
		used, err := repo.Create(ctx, userID, "usedverify", time.Now().Add(time.Hour))
		require.NoError(t, err)
		err = repo.MarkUsed(ctx, used.ID)
		require.NoError(t, err)

		// Create valid verification
		_, err = repo.Create(ctx, userID, "validverify", time.Now().Add(time.Hour))
		require.NoError(t, err)

		deleted, err := repo.DeleteExpiredOrUsed(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, deleted, int64(2))

		// Valid verification should still exist
		_, err = repo.GetByTokenHash(ctx, "validverify")
		assert.NoError(t, err)

		// Expired and used should be deleted
		_, err = repo.GetByTokenHash(ctx, "expiredverify")
		require.ErrorIs(t, err, sql.ErrNoRows)
		_, err = repo.GetByTokenHash(ctx, "usedverify")
		require.ErrorIs(t, err, sql.ErrNoRows)
	})
}
