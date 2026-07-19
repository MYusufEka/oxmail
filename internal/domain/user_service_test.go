package domain_test

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/MYusufEka/oxmail/internal/database"
	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func setupUserTestDB(t *testing.T) *database.DB {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := database.Open(":memory:", logger)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func seedTestDomain(t *testing.T, db *database.DB, name string) int64 {
	t.Helper()
	res, err := db.Conn.Exec(
		"INSERT INTO domains (name, active) VALUES (?, 1)", name,
	)
	require.NoError(t, err)
	id, err := res.LastInsertId()
	require.NoError(t, err)
	return id
}

// domainLookup implements domain.DomainLookup for testing.
type domainLookup struct {
	db *database.DB
}

func (d *domainLookup) GetDomainByName(ctx context.Context, name string) (*domain.Domain, error) {
	var dom domain.Domain
	err := d.db.Conn.QueryRowContext(ctx,
		"SELECT id, name, active, created_at, updated_at FROM domains WHERE name = ?", name,
	).Scan(&dom.ID, &dom.Name, &dom.Active, &dom.CreatedAt, &dom.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &dom, nil
}

func newTestUserService(t *testing.T) (*domain.UserService, *database.DB) {
	t.Helper()
	db := setupUserTestDB(t)
	seedTestDomain(t, db, "local.test")
	lookup := &domainLookup{db: db}
	svc := domain.NewUserService(db, lookup)
	return svc, db
}

func TestUserService_Create(t *testing.T) {
	svc, _ := newTestUserService(t)
	ctx := context.Background()

	user, err := svc.Create(ctx, domain.CreateUserRequest{
		Email:       "alice@local.test",
		Password:    "TestPass123!",
		DisplayName: "Alice",
		Quota:       1024,
	})

	require.NoError(t, err)
	assert.Equal(t, "alice@local.test", user.Email)
	assert.Equal(t, "Alice", user.DisplayName)
	assert.Equal(t, int64(1024), user.Quota)
	assert.True(t, user.Active)
	assert.NotZero(t, user.ID)
	assert.NotZero(t, user.CreatedAt)
	// Password hash should never be empty but also never be the plaintext
	assert.NotEmpty(t, user.PasswordHash)
	assert.NotEqual(t, "TestPass123!", user.PasswordHash)
}

func TestUserService_Create_BcryptHash(t *testing.T) {
	svc, _ := newTestUserService(t)
	ctx := context.Background()

	user, err := svc.Create(ctx, domain.CreateUserRequest{
		Email:    "bob@local.test",
		Password: "SecurePass456!",
	})

	require.NoError(t, err)
	// Hash must start with $2a$ or $2b$ (bcrypt)
	assert.Regexp(t, `^\$2[ab]\$12\$`, user.PasswordHash)
}

func TestUserService_Create_DomainNotFound(t *testing.T) {
	svc, _ := newTestUserService(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, domain.CreateUserRequest{
		Email:    "alice@nonexistent.test",
		Password: "TestPass123!",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrDomainNotFound)
}

func TestUserService_Create_DuplicateEmail(t *testing.T) {
	svc, _ := newTestUserService(t)
	ctx := context.Background()

	_, err := svc.Create(ctx, domain.CreateUserRequest{
		Email:    "alice@local.test",
		Password: "TestPass123!",
	})
	require.NoError(t, err)

	_, err = svc.Create(ctx, domain.CreateUserRequest{
		Email:    "alice@local.test",
		Password: "AnotherPass456!",
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserExists)
}

func TestUserService_Create_InvalidEmail(t *testing.T) {
	svc, _ := newTestUserService(t)
	ctx := context.Background()

	cases := []string{
		"",
		"noatsign",
		"@nodomain",
		"nouser@",
		"spaces in@email.com",
	}

	for _, email := range cases {
		_, err := svc.Create(ctx, domain.CreateUserRequest{
			Email:    email,
			Password: "TestPass123!",
		})
		assert.Error(t, err, "expected error for email: %s", email)
		assert.ErrorIs(t, err, domain.ErrInvalidEmail)
	}
}

func TestUserService_Get(t *testing.T) {
	svc, _ := newTestUserService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, domain.CreateUserRequest{
		Email:    "alice@local.test",
		Password: "TestPass123!",
	})
	require.NoError(t, err)

	fetched, err := svc.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.Email, fetched.Email)
	assert.Equal(t, created.ID, fetched.ID)
}

func TestUserService_Get_NotFound(t *testing.T) {
	svc, _ := newTestUserService(t)
	ctx := context.Background()

	_, err := svc.GetByID(ctx, 9999)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestUserService_List(t *testing.T) {
	svc, _ := newTestUserService(t)
	ctx := context.Background()

	for i := range 5 {
		_, err := svc.Create(ctx, domain.CreateUserRequest{
			Email:    fmt.Sprintf("user%d@local.test", i),
			Password: "TestPass123!",
		})
		require.NoError(t, err)
	}

	users, total, err := svc.List(ctx, domain.UserListParams{
		Page:  1,
		Limit: 3,
	})
	require.NoError(t, err)
	assert.Len(t, users, 3)
	assert.Equal(t, 5, total)
}

func TestUserService_List_FilterByDomain(t *testing.T) {
	svc, db := newTestUserService(t)
	ctx := context.Background()
	seedTestDomain(t, db, "other.test")

	_, err := svc.Create(ctx, domain.CreateUserRequest{
		Email:    "alice@local.test",
		Password: "TestPass123!",
	})
	require.NoError(t, err)

	_, err = svc.Create(ctx, domain.CreateUserRequest{
		Email:    "bob@other.test",
		Password: "TestPass123!",
	})
	require.NoError(t, err)

	users, total, err := svc.List(ctx, domain.UserListParams{
		Domain: "local.test",
		Page:   1,
		Limit:  10,
	})
	require.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, 1, total)
	assert.Equal(t, "alice@local.test", users[0].Email)
}

func TestUserService_Delete(t *testing.T) {
	svc, _ := newTestUserService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, domain.CreateUserRequest{
		Email:    "alice@local.test",
		Password: "TestPass123!",
	})
	require.NoError(t, err)

	err = svc.Delete(ctx, created.ID)
	require.NoError(t, err)

	_, err = svc.GetByID(ctx, created.ID)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestUserService_Delete_NotFound(t *testing.T) {
	svc, _ := newTestUserService(t)
	ctx := context.Background()

	err := svc.Delete(ctx, 9999)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestUserService_GetByID_MustChangePassword(t *testing.T) {
	svc, db := newTestUserService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, domain.CreateUserRequest{
		Email:    "alice@local.test",
		Password: "TestPass123!",
	})
	require.NoError(t, err)

	_, err = db.Conn.ExecContext(ctx, "UPDATE users SET must_change_password = 1 WHERE id = ?", created.ID)
	require.NoError(t, err)

	fetched, err := svc.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.True(t, fetched.MustChangePassword)
}

func TestUserService_List_MustChangePassword(t *testing.T) {
	svc, db := newTestUserService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, domain.CreateUserRequest{
		Email:    "alice@local.test",
		Password: "TestPass123!",
	})
	require.NoError(t, err)

	_, err = db.Conn.ExecContext(ctx, "UPDATE users SET must_change_password = 1 WHERE id = ?", created.ID)
	require.NoError(t, err)

	users, _, err := svc.List(ctx, domain.UserListParams{Page: 1, Limit: 10})
	require.NoError(t, err)
	require.Len(t, users, 1)
	assert.True(t, users[0].MustChangePassword)
}

func TestUserService_GetByEmail_MustChangePassword(t *testing.T) {
	svc, db := newTestUserService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, domain.CreateUserRequest{
		Email:    "alice@local.test",
		Password: "TestPass123!",
	})
	require.NoError(t, err)

	_, err = db.Conn.ExecContext(ctx, "UPDATE users SET must_change_password = 1 WHERE id = ?", created.ID)
	require.NoError(t, err)

	fetched, err := svc.GetByEmail(ctx, created.Email)
	require.NoError(t, err)
	assert.True(t, fetched.MustChangePassword)
}

func TestUserService_UpdatePasswordClearsMustChangePassword(t *testing.T) {
	svc, db := newTestUserService(t)
	ctx := context.Background()

	created, err := svc.Create(ctx, domain.CreateUserRequest{
		Email:    "alice@local.test",
		Password: "OldPass123!",
	})
	require.NoError(t, err)

	_, err = db.Conn.ExecContext(ctx, "UPDATE users SET must_change_password = 1 WHERE id = ?", created.ID)
	require.NoError(t, err)

	newPassword := "NewPass456!"
	updated, err := svc.Update(ctx, created.ID, domain.UpdateUserRequest{Password: &newPassword})
	require.NoError(t, err)

	assert.False(t, updated.MustChangePassword)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(updated.PasswordHash), []byte(newPassword)))

	fetched, err := svc.GetByID(ctx, created.ID)
	require.NoError(t, err)
	assert.False(t, fetched.MustChangePassword)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(fetched.PasswordHash), []byte(newPassword)))
}
