package domain_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/MYusufEka/oxmail/internal/database"
	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupDomainTestDB(t *testing.T) *database.DB {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := database.Open(":memory:", logger)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestDomainService_Create(t *testing.T) {
	db := setupDomainTestDB(t)
	svc := domain.NewDomainService(db)
	ctx := context.Background()

	t.Run("creates domain successfully", func(t *testing.T) {
		created, err := svc.Create(ctx, "example.com")
		require.NoError(t, err)
		assert.Equal(t, "example.com", created.Name)
		assert.True(t, created.Active)
		assert.NotZero(t, created.ID)
		assert.False(t, created.CreatedAt.IsZero())
		assert.False(t, created.UpdatedAt.IsZero())
	})

	t.Run("rejects duplicate domain", func(t *testing.T) {
		_, err := svc.Create(ctx, "duplicate.com")
		require.NoError(t, err)

		_, err = svc.Create(ctx, "duplicate.com")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrDomainExists)
	})

	t.Run("rejects empty domain name", func(t *testing.T) {
		_, err := svc.Create(ctx, "")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDomain)
	})

	t.Run("rejects domain with spaces", func(t *testing.T) {
		_, err := svc.Create(ctx, "has space.com")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDomain)
	})

	t.Run("rejects domain without TLD", func(t *testing.T) {
		_, err := svc.Create(ctx, "nodot")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDomain)
	})

	t.Run("rejects domain with special characters", func(t *testing.T) {
		_, err := svc.Create(ctx, "bad!domain.com")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDomain)
	})

	t.Run("rejects domain starting with hyphen", func(t *testing.T) {
		_, err := svc.Create(ctx, "-invalid.com")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDomain)
	})

	t.Run("rejects domain with trailing dot", func(t *testing.T) {
		_, err := svc.Create(ctx, "trailing.com.")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrInvalidDomain)
	})

	t.Run("accepts valid subdomain", func(t *testing.T) {
		created, err := svc.Create(ctx, "mail.example.org")
		require.NoError(t, err)
		assert.Equal(t, "mail.example.org", created.Name)
	})
}

func TestDomainService_Get(t *testing.T) {
	db := setupDomainTestDB(t)
	svc := domain.NewDomainService(db)
	ctx := context.Background()

	t.Run("gets existing domain", func(t *testing.T) {
		created, err := svc.Create(ctx, "getme.com")
		require.NoError(t, err)

		got, err := svc.Get(ctx, created.Name)
		require.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, created.Name, got.Name)
	})

	t.Run("returns not found for missing domain", func(t *testing.T) {
		_, err := svc.Get(ctx, "nonexistent.com")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrDomainNotFound)
	})
}

func TestDomainService_List(t *testing.T) {
	t.Run("returns empty list when no domains", func(t *testing.T) {
		db := setupDomainTestDB(t)
		svc := domain.NewDomainService(db)
		ctx := context.Background()

		domains, total, err := svc.List(ctx, 1, 50)
		require.NoError(t, err)
		assert.Empty(t, domains)
		assert.Equal(t, 0, total)
	})

	t.Run("returns paginated results", func(t *testing.T) {
		db := setupDomainTestDB(t)
		svc := domain.NewDomainService(db)
		ctx := context.Background()

		for i := range 5 {
			_, err := svc.Create(ctx, fmt.Sprintf("domain%d.com", i))
			require.NoError(t, err)
		}

		domains, total, err := svc.List(ctx, 1, 2)
		require.NoError(t, err)
		assert.Len(t, domains, 2)
		assert.Equal(t, 5, total)

		domains, total, err = svc.List(ctx, 3, 2)
		require.NoError(t, err)
		assert.Len(t, domains, 1)
		assert.Equal(t, 5, total)
	})
}

func TestDomainService_Delete(t *testing.T) {
	db := setupDomainTestDB(t)
	svc := domain.NewDomainService(db)
	ctx := context.Background()

	t.Run("deletes existing domain", func(t *testing.T) {
		_, err := svc.Create(ctx, "deleteme.com")
		require.NoError(t, err)

		err = svc.Delete(ctx, "deleteme.com")
		require.NoError(t, err)

		_, err = svc.Get(ctx, "deleteme.com")
		assert.ErrorIs(t, err, domain.ErrDomainNotFound)
	})

	t.Run("returns not found for missing domain", func(t *testing.T) {
		err := svc.Delete(ctx, "ghost.com")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrDomainNotFound)
	})
}
