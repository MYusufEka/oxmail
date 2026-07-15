package domain

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/MYusufEka/oxmail/internal/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAuditDB(t *testing.T) *AuditService {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	db, err := database.Open(":memory:", logger)
	require.NoError(t, err)
	t.Cleanup(func() { db.Conn.Close() })
	return NewAuditService(db.Conn)
}

func TestAuditService_Log(t *testing.T) {
	svc := newTestAuditDB(t)
	ctx := context.Background()

	err := svc.Log(ctx, "admin", "domain.create", "domain", "example.com", `{"domain":"example.com"}`)
	require.NoError(t, err)

	entries, err := svc.List(ctx, 10, 0)
	require.NoError(t, err)
	require.Len(t, entries, 1)

	e := entries[0]
	assert.Equal(t, "admin", e.Actor)
	assert.Equal(t, "domain.create", e.Action)
	assert.Equal(t, "domain", e.TargetType)
	assert.Equal(t, "example.com", e.TargetID)
	assert.Equal(t, `{"domain":"example.com"}`, e.Detail)
	assert.False(t, e.CreatedAt.IsZero())
}

func TestAuditService_Log_EmptyActor_UsesDefault(t *testing.T) {
	svc := newTestAuditDB(t)
	ctx := context.Background()

	t.Setenv("OXMAIL_MODE", "dev")

	err := svc.Log(ctx, "", "user.delete", "user", "42", `{"id":42}`)
	require.NoError(t, err)

	entries, err := svc.List(ctx, 10, 0)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "admin", entries[0].Actor)
}

func TestAuditService_Log_EmptyDetail_DefaultsToEmpty(t *testing.T) {
	svc := newTestAuditDB(t)
	ctx := context.Background()

	err := svc.Log(ctx, "admin", "dkim.generate", "domain", "mail.test", "")
	require.NoError(t, err)

	entries, err := svc.List(ctx, 10, 0)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "{}", entries[0].Detail)
}

func TestAuditService_List_Pagination(t *testing.T) {
	svc := newTestAuditDB(t)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		err := svc.Log(ctx, "admin", "domain.create", "domain", "example.com", "{}")
		require.NoError(t, err)
	}

	all, err := svc.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.Len(t, all, 5)

	page, err := svc.List(ctx, 2, 0)
	require.NoError(t, err)
	assert.Len(t, page, 2)

	offset, err := svc.List(ctx, 10, 3)
	require.NoError(t, err)
	assert.Len(t, offset, 2)
}

func TestAuditService_List_Empty(t *testing.T) {
	svc := newTestAuditDB(t)
	ctx := context.Background()

	entries, err := svc.List(ctx, 10, 0)
	require.NoError(t, err)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 0)
}
