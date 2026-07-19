package domain_test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/MYusufEka/oxmail/internal/database"
	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSignatureTestDB(t *testing.T) *database.DB {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := database.Open(":memory:", logger)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestSignatureService_GetMissing(t *testing.T) {
	db := setupSignatureTestDB(t)
	svc := domain.NewSignatureService(db.Conn)
	ctx := context.Background()

	signature, err := svc.Get(ctx, "alice@example.com")

	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", signature.Email)
	assert.Equal(t, "", signature.Content)
	assert.False(t, signature.Enabled)
}

func TestSignatureService_UpsertAndGet(t *testing.T) {
	db := setupSignatureTestDB(t)
	svc := domain.NewSignatureService(db.Conn)
	ctx := context.Background()

	saved, err := svc.Upsert(ctx, "alice@example.com", "<p>Regards</p>", true)

	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", saved.Email)
	assert.Equal(t, "<p>Regards</p>", saved.Content)
	assert.True(t, saved.Enabled)

	got, err := svc.Get(ctx, "alice@example.com")
	require.NoError(t, err)
	assert.Equal(t, saved.Email, got.Email)
	assert.Equal(t, saved.Content, got.Content)
	assert.Equal(t, saved.Enabled, got.Enabled)
}

func TestSignatureService_Update(t *testing.T) {
	db := setupSignatureTestDB(t)
	svc := domain.NewSignatureService(db.Conn)
	ctx := context.Background()

	_, err := svc.Upsert(ctx, "alice@example.com", "<p>Old</p>", true)
	require.NoError(t, err)

	updated, err := svc.Upsert(ctx, "alice@example.com", "<p>New</p>", false)

	require.NoError(t, err)
	assert.Equal(t, "alice@example.com", updated.Email)
	assert.Equal(t, "<p>New</p>", updated.Content)
	assert.False(t, updated.Enabled)

	got, err := svc.Get(ctx, "alice@example.com")
	require.NoError(t, err)
	assert.Equal(t, "<p>New</p>", got.Content)
	assert.False(t, got.Enabled)
}

func TestSignatureService_Delete(t *testing.T) {
	db := setupSignatureTestDB(t)
	svc := domain.NewSignatureService(db.Conn)
	ctx := context.Background()

	_, err := svc.Upsert(ctx, "alice@example.com", "<p>Bye</p>", true)
	require.NoError(t, err)

	err = svc.Delete(ctx, "alice@example.com")
	require.NoError(t, err)

	signature, err := svc.Get(ctx, "alice@example.com")
	require.NoError(t, err)
	assert.Equal(t, "", signature.Content)
	assert.False(t, signature.Enabled)

	err = svc.Delete(ctx, "alice@example.com")
	assert.NoError(t, err)
}
