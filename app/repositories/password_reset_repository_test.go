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

func TestPasswordResetRepository(t *testing.T) {
	tx := setupTest(t)
	userRepo := NewUserRepository(testDB).WithTx(tx)
	repo := NewPasswordResetRepository(testDB).WithTx(tx)
	ctx := context.Background()

	createUser := func(t *testing.T) uuid.UUID {
		user, err := userRepo.Create(ctx, "Test User", uuid.NewString()+"@example.com", "hash")
		require.NoError(t, err)
		return user.ID
	}

	t.Run("Create", func(t *testing.T) {
		userID := createUser(t)
		expiresAt := time.Now().Add(24 * time.Hour)

		reset, err := repo.Create(ctx, userID, "resethash123", expiresAt)
		require.NoError(t, err)
		assert.NotEmpty(t, reset.ID)
		assert.Equal(t, userID, reset.UserID)
		assert.Equal(t, "resethash123", reset.TokenHash)
		assert.Nil(t, reset.UsedAt)
		assert.NotZero(t, reset.CreatedAt)
	})

	t.Run("GetByTokenHash", func(t *testing.T) {
		userID := createUser(t)
		expiresAt := time.Now().Add(24 * time.Hour)

		created, err := repo.Create(ctx, userID, "findresethash", expiresAt)
		require.NoError(t, err)

		found, err := repo.GetByTokenHash(ctx, "findresethash")
		require.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.TokenHash, found.TokenHash)
	})

	t.Run("GetByTokenHash_NotFound", func(t *testing.T) {
		reset, err := repo.GetByTokenHash(ctx, "nonexistenthash")
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, reset)
	})

	t.Run("MarkUsed", func(t *testing.T) {
		userID := createUser(t)
		reset, err := repo.Create(ctx, userID, "markusedresethash", time.Now().Add(time.Hour))
		require.NoError(t, err)
		assert.Nil(t, reset.UsedAt)

		err = repo.MarkUsed(ctx, reset.ID)
		require.NoError(t, err)

		found, err := repo.GetByTokenHash(ctx, "markusedresethash")
		require.NoError(t, err)
		assert.NotNil(t, found.UsedAt)
	})

	t.Run("InvalidateAllForUser", func(t *testing.T) {
		userID := createUser(t)

		_, err := repo.Create(ctx, userID, "invalidatereset1", time.Now().Add(time.Hour))
		require.NoError(t, err)
		_, err = repo.Create(ctx, userID, "invalidatereset2", time.Now().Add(time.Hour))
		require.NoError(t, err)

		err = repo.InvalidateAllForUser(ctx, userID)
		require.NoError(t, err)

		r1, _ := repo.GetByTokenHash(ctx, "invalidatereset1")
		r2, _ := repo.GetByTokenHash(ctx, "invalidatereset2")
		assert.NotNil(t, r1.UsedAt)
		assert.NotNil(t, r2.UsedAt)
	})

	t.Run("DeleteExpiredOrUsed", func(t *testing.T) {
		userID := createUser(t)

		// Create expired reset
		_, err := repo.Create(ctx, userID, "expiredreset", time.Now().Add(-time.Hour))
		require.NoError(t, err)

		// Create used reset
		used, err := repo.Create(ctx, userID, "usedreset", time.Now().Add(time.Hour))
		require.NoError(t, err)
		err = repo.MarkUsed(ctx, used.ID)
		require.NoError(t, err)

		// Create valid reset
		_, err = repo.Create(ctx, userID, "validreset", time.Now().Add(time.Hour))
		require.NoError(t, err)

		deleted, err := repo.DeleteExpiredOrUsed(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, deleted, int64(2))

		// Valid reset should still exist
		_, err = repo.GetByTokenHash(ctx, "validreset")
		assert.NoError(t, err)

		// Expired and used should be deleted
		_, err = repo.GetByTokenHash(ctx, "expiredreset")
		require.ErrorIs(t, err, sql.ErrNoRows)
		_, err = repo.GetByTokenHash(ctx, "usedreset")
		require.ErrorIs(t, err, sql.ErrNoRows)
	})
}
