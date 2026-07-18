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
	"golang.org/x/crypto/bcrypt"
)

func setupContactTestDB(t *testing.T) *database.DB {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := database.Open(":memory:", logger)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func seedContactDomainAndUser(t *testing.T, db *database.DB, domainName, email string) {
	t.Helper()
	_, err := db.Conn.Exec("INSERT INTO domains (name, active) VALUES (?, 1)", domainName)
	require.NoError(t, err)

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	require.NoError(t, err)

	_, err = db.Conn.Exec(
		"INSERT INTO users (email, password_hash, domain_id, display_name, active) VALUES (?, ?, (SELECT id FROM domains WHERE name = ?), ?, 1)",
		email, string(hash), domainName, "Test User",
	)
	require.NoError(t, err)
}

func TestContactService_Create(t *testing.T) {
	db := setupContactTestDB(t)
	seedContactDomainAndUser(t, db, "example.com", "alice@example.com")
	svc := domain.NewContactService(db.Conn)
	ctx := context.Background()

	t.Run("creates contact successfully", func(t *testing.T) {
		contact, err := svc.Create(ctx, "alice@example.com", domain.CreateContactRequest{
			Name:  "Bob Smith",
			Email: "bob@example.com",
			Phone: "+1234567890",
		})
		require.NoError(t, err)
		assert.Equal(t, "Bob Smith", contact.Name)
		assert.Equal(t, "bob@example.com", contact.Email)
		assert.Equal(t, "+1234567890", contact.Phone)
		assert.Equal(t, "alice@example.com", contact.UserEmail)
		assert.NotZero(t, contact.ID)
		assert.False(t, contact.CreatedAt.IsZero())
	})

	t.Run("rejects duplicate contact", func(t *testing.T) {
		_, err := svc.Create(ctx, "alice@example.com", domain.CreateContactRequest{
			Name:  "Bob Dup",
			Email: "bobdup@example.com",
		})
		require.NoError(t, err)

		_, err = svc.Create(ctx, "alice@example.com", domain.CreateContactRequest{
			Name:  "Bob Dup Again",
			Email: "bobdup@example.com",
		})
		require.ErrorIs(t, err, domain.ErrContactExists)
	})

	t.Run("same email different user is allowed", func(t *testing.T) {
		seedContactDomainAndUser(t, db, "other.com", "carol@other.com")

		_, err := svc.Create(ctx, "alice@example.com", domain.CreateContactRequest{
			Name:  "Dave",
			Email: "dave@example.com",
		})
		require.NoError(t, err)

		_, err = svc.Create(ctx, "carol@other.com", domain.CreateContactRequest{
			Name:  "Dave",
			Email: "dave@example.com",
		})
		require.NoError(t, err)
	})

	t.Run("requires name", func(t *testing.T) {
		_, err := svc.Create(ctx, "alice@example.com", domain.CreateContactRequest{
			Email: "nobody@example.com",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "name is required")
	})

	t.Run("requires email", func(t *testing.T) {
		_, err := svc.Create(ctx, "alice@example.com", domain.CreateContactRequest{
			Name: "No Email",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "email is required")
	})
}

func TestContactService_Get(t *testing.T) {
	db := setupContactTestDB(t)
	seedContactDomainAndUser(t, db, "example.com", "alice@example.com")
	svc := domain.NewContactService(db.Conn)
	ctx := context.Background()

	created, err := svc.Create(ctx, "alice@example.com", domain.CreateContactRequest{
		Name:  "Charlie",
		Email: "charlie@example.com",
	})
	require.NoError(t, err)

	t.Run("gets existing contact", func(t *testing.T) {
		got, err := svc.Get(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, got.ID)
		assert.Equal(t, "Charlie", got.Name)
		assert.Equal(t, "charlie@example.com", got.Email)
	})

	t.Run("returns not found for missing ID", func(t *testing.T) {
		_, err := svc.Get(ctx, 99999)
		require.ErrorIs(t, err, domain.ErrContactNotFound)
	})
}

func TestContactService_List(t *testing.T) {
	db := setupContactTestDB(t)
	seedContactDomainAndUser(t, db, "example.com", "alice@example.com")
	svc := domain.NewContactService(db.Conn)
	ctx := context.Background()

	t.Run("returns empty list when no contacts", func(t *testing.T) {
		contacts, err := svc.List(ctx, "alice@example.com")
		require.NoError(t, err)
		assert.Empty(t, contacts)
	})

	t.Run("returns contacts ordered by name", func(t *testing.T) {
		_, err := svc.Create(ctx, "alice@example.com", domain.CreateContactRequest{
			Name:  "Zoe",
			Email: "zoe@example.com",
		})
		require.NoError(t, err)

		_, err = svc.Create(ctx, "alice@example.com", domain.CreateContactRequest{
			Name:  "Adam",
			Email: "adam@example.com",
		})
		require.NoError(t, err)

		contacts, err := svc.List(ctx, "alice@example.com")
		require.NoError(t, err)
		require.Len(t, contacts, 2)
		assert.Equal(t, "Adam", contacts[0].Name)
		assert.Equal(t, "Zoe", contacts[1].Name)
	})

	t.Run("returns only contacts for given user", func(t *testing.T) {
		seedContactDomainAndUser(t, db, "other.com", "bob@other.com")

		contacts, err := svc.List(ctx, "bob@other.com")
		require.NoError(t, err)
		assert.Empty(t, contacts)
	})
}

func TestContactService_Update(t *testing.T) {
	db := setupContactTestDB(t)
	seedContactDomainAndUser(t, db, "example.com", "alice@example.com")
	svc := domain.NewContactService(db.Conn)
	ctx := context.Background()

	created, err := svc.Create(ctx, "alice@example.com", domain.CreateContactRequest{
		Name:  "Old Name",
		Email: "old@example.com",
		Phone: "111",
	})
	require.NoError(t, err)

	t.Run("updates name only", func(t *testing.T) {
		newName := "New Name"
		updated, err := svc.Update(ctx, created.ID, domain.UpdateContactRequest{
			Name: &newName,
		})
		require.NoError(t, err)
		assert.Equal(t, "New Name", updated.Name)
		assert.Equal(t, "old@example.com", updated.Email)
		assert.Equal(t, "111", updated.Phone)
	})

	t.Run("updates email only", func(t *testing.T) {
		newEmail := "newemail@example.com"
		updated, err := svc.Update(ctx, created.ID, domain.UpdateContactRequest{
			Email: &newEmail,
		})
		require.NoError(t, err)
		assert.Equal(t, "newemail@example.com", updated.Email)
	})

	t.Run("no changes when no fields provided", func(t *testing.T) {
		updated, err := svc.Update(ctx, created.ID, domain.UpdateContactRequest{})
		require.NoError(t, err)
		assert.NotNil(t, updated)
	})

	t.Run("returns not found for missing contact", func(t *testing.T) {
		name := "Ghost"
		_, err := svc.Update(ctx, 99999, domain.UpdateContactRequest{
			Name: &name,
		})
		require.ErrorIs(t, err, domain.ErrContactNotFound)
	})
}

func TestContactService_Delete(t *testing.T) {
	db := setupContactTestDB(t)
	seedContactDomainAndUser(t, db, "example.com", "alice@example.com")
	svc := domain.NewContactService(db.Conn)
	ctx := context.Background()

	created, err := svc.Create(ctx, "alice@example.com", domain.CreateContactRequest{
		Name:  "Deletable",
		Email: "delete@example.com",
	})
	require.NoError(t, err)

	t.Run("deletes existing contact", func(t *testing.T) {
		err := svc.Delete(ctx, created.ID)
		require.NoError(t, err)

		_, err = svc.Get(ctx, created.ID)
		require.ErrorIs(t, err, domain.ErrContactNotFound)
	})

	t.Run("returns not found for missing contact", func(t *testing.T) {
		err := svc.Delete(ctx, 99999)
		require.ErrorIs(t, err, domain.ErrContactNotFound)
	})
}

func TestContactService_ResolveName(t *testing.T) {
	db := setupContactTestDB(t)
	seedContactDomainAndUser(t, db, "example.com", "alice@example.com")
	svc := domain.NewContactService(db.Conn)
	ctx := context.Background()

	_, err := svc.Create(ctx, "alice@example.com", domain.CreateContactRequest{
		Name:  "Bobbie",
		Email: "bobbie@example.com",
	})
	require.NoError(t, err)

	t.Run("resolves existing contact", func(t *testing.T) {
		name := svc.ResolveName(ctx, "alice@example.com", "bobbie@example.com")
		assert.Equal(t, "Bobbie", name)
	})

	t.Run("returns empty for nonexistent contact", func(t *testing.T) {
		name := svc.ResolveName(ctx, "alice@example.com", "nobody@example.com")
		assert.Equal(t, "", name)
	})

	t.Run("returns empty for wrong user", func(t *testing.T) {
		name := svc.ResolveName(ctx, "wrong@example.com", "bobbie@example.com")
		assert.Equal(t, "", name)
	})
}
