package repositories

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository(t *testing.T) {
	tx := setupTest(t)
	repo := NewUserRepository(testDB).WithTx(tx)
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		user, err := repo.Create(ctx, "John Doe", "john@example.com", "hashedpassword")
		require.NoError(t, err)
		assert.NotEmpty(t, user.ID)
		assert.Equal(t, "John Doe", user.Name)
		assert.Equal(t, "john@example.com", user.Email)
		assert.Equal(t, "hashedpassword", user.PasswordHash)
		assert.NotZero(t, user.CreatedAt)
		assert.NotZero(t, user.UpdatedAt)
	})

	t.Run("GetByID", func(t *testing.T) {
		created, err := repo.Create(ctx, "Jane Doe", "jane@example.com", "hash123")
		require.NoError(t, err)

		found, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.Name, found.Name)
		assert.Equal(t, created.Email, found.Email)
	})

	t.Run("GetByID_NotFound", func(t *testing.T) {
		user, err := repo.GetByID(ctx, uuid.New())
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, user)
	})

	t.Run("GetByEmail", func(t *testing.T) {
		created, err := repo.Create(ctx, "Bob Smith", "bob@example.com", "hash456")
		require.NoError(t, err)

		found, err := repo.GetByEmail(ctx, "bob@example.com")
		require.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, created.Email, found.Email)
	})

	t.Run("GetByEmail_NotFound", func(t *testing.T) {
		user, err := repo.GetByEmail(ctx, "notfound@example.com")
		require.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, user)
	})

	t.Run("UpdatePassword", func(t *testing.T) {
		user, err := repo.Create(ctx, "Alice", "alice@example.com", "oldpass")
		require.NoError(t, err)

		err = repo.UpdatePassword(ctx, user.ID, "newpass")
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "newpass", updated.PasswordHash)
	})

	t.Run("MarkEmailVerified", func(t *testing.T) {
		user, err := repo.Create(ctx, "Charlie", "charlie@example.com", "pass")
		require.NoError(t, err)
		assert.Nil(t, user.EmailVerifiedAt)

		err = repo.MarkEmailVerified(ctx, user.ID)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.NotNil(t, updated.EmailVerifiedAt)
	})

	t.Run("EmailExists", func(t *testing.T) {
		_, err := repo.Create(ctx, "Dave", "dave@example.com", "pass")
		require.NoError(t, err)

		exists, err := repo.EmailExists(ctx, "dave@example.com")
		require.NoError(t, err)
		assert.True(t, exists)

		exists, err = repo.EmailExists(ctx, "notexists@example.com")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}
