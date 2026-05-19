package domain_test

import (
	"database/sql"
	"log/slog"
	"os"
	"testing"

	"github.com/MYusufEka/oxmail/internal/database"
	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *database.DB {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := database.Open(":memory:", logger)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	// Seed a domain for alias tests.
	_, err = db.Conn.Exec(`INSERT INTO domains (name, active) VALUES ('local.test', 1)`)
	require.NoError(t, err)

	return db
}

func getDomainID(t *testing.T, conn *sql.DB, name string) int64 {
	t.Helper()
	var id int64
	err := conn.QueryRow(`SELECT id FROM domains WHERE name = ?`, name).Scan(&id)
	require.NoError(t, err)
	return id
}

func TestAliasService_Create(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewAliasService(db.Conn)

	alias, err := svc.Create("info@local.test", "alice@local.test")
	require.NoError(t, err)

	assert.NotZero(t, alias.ID)
	assert.Equal(t, "info@local.test", alias.SourceAddress)
	assert.Equal(t, "alice@local.test", alias.DestinationAddress)
	assert.True(t, alias.Active)
	assert.False(t, alias.CreatedAt.IsZero())
}

func TestAliasService_Create_InvalidSourceDomain(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewAliasService(db.Conn)

	_, err := svc.Create("info@unknown.test", "alice@local.test")
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrDomainNotFound)
}

func TestAliasService_Create_InvalidEmailFormat(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewAliasService(db.Conn)

	_, err := svc.Create("notanemail", "alice@local.test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestAliasService_Create_DuplicateAlias(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewAliasService(db.Conn)

	_, err := svc.Create("info@local.test", "alice@local.test")
	require.NoError(t, err)

	_, err = svc.Create("info@local.test", "alice@local.test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestAliasService_Create_CircularDirect(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewAliasService(db.Conn)

	// alice -> info
	_, err := svc.Create("alice@local.test", "info@local.test")
	require.NoError(t, err)

	// info -> alice would create a cycle
	_, err = svc.Create("info@local.test", "alice@local.test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular")
}

func TestAliasService_Create_CircularTransitive(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewAliasService(db.Conn)

	// a -> b
	_, err := svc.Create("a@local.test", "b@local.test")
	require.NoError(t, err)

	// b -> c
	_, err = svc.Create("b@local.test", "c@local.test")
	require.NoError(t, err)

	// c -> a would create a transitive cycle
	_, err = svc.Create("c@local.test", "a@local.test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular")
}

func TestAliasService_Get(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewAliasService(db.Conn)

	created, err := svc.Create("info@local.test", "alice@local.test")
	require.NoError(t, err)

	alias, err := svc.Get(created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, alias.ID)
	assert.Equal(t, "info@local.test", alias.SourceAddress)
}

func TestAliasService_Get_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewAliasService(db.Conn)

	_, err := svc.Get(9999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAliasService_List(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewAliasService(db.Conn)

	_, err := svc.Create("info@local.test", "alice@local.test")
	require.NoError(t, err)
	_, err = svc.Create("support@local.test", "bob@local.test")
	require.NoError(t, err)

	aliases, err := svc.List()
	require.NoError(t, err)
	assert.Len(t, aliases, 2)
}

func TestAliasService_List_Empty(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewAliasService(db.Conn)

	aliases, err := svc.List()
	require.NoError(t, err)
	assert.Empty(t, aliases)
}

func TestAliasService_Delete(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewAliasService(db.Conn)

	created, err := svc.Create("info@local.test", "alice@local.test")
	require.NoError(t, err)

	err = svc.Delete(created.ID)
	require.NoError(t, err)

	_, err = svc.Get(created.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAliasService_Delete_NotFound(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewAliasService(db.Conn)

	err := svc.Delete(9999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestAliasService_MultipleAliasesSameSource(t *testing.T) {
	db := setupTestDB(t)
	svc := domain.NewAliasService(db.Conn)

	_, err := svc.Create("info@local.test", "alice@local.test")
	require.NoError(t, err)
	_, err = svc.Create("info@local.test", "bob@local.test")
	require.NoError(t, err)

	aliases, err := svc.List()
	require.NoError(t, err)
	assert.Len(t, aliases, 2)
}
