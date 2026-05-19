package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/MYusufEka/oxmail/internal/api"
	"github.com/MYusufEka/oxmail/internal/database"
	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUsersTestDB(t *testing.T) *database.DB {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := database.Open(":memory:", logger)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.Conn.Exec("INSERT INTO domains (name, active) VALUES ('local.test', 1)")
	require.NoError(t, err)

	return db
}

type testDomainLookup struct {
	db *database.DB
}

func (d *testDomainLookup) GetDomainByName(ctx context.Context, name string) (*domain.Domain, error) {
	var dom domain.Domain
	err := d.db.Conn.QueryRowContext(ctx,
		"SELECT id, name, active, created_at, updated_at FROM domains WHERE name = ?", name,
	).Scan(&dom.ID, &dom.Name, &dom.Active, &dom.CreatedAt, &dom.UpdatedAt)
	if err != nil {
		return nil, nil
	}
	return &dom, nil
}

func newUsersHandler(t *testing.T) *api.UsersHandler {
	t.Helper()
	db := setupUsersTestDB(t)
	lookup := &testDomainLookup{db: db}
	userSvc := domain.NewUserService(db, lookup)
	return api.NewUsersHandler(userSvc)
}

func TestUsersHandler_Create(t *testing.T) {
	h := newUsersHandler(t)

	body := `{"email":"alice@local.test","password":"TestPass123!","displayName":"Alice","quota":1024}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var user domain.User
	err := json.NewDecoder(rec.Body).Decode(&user)
	require.NoError(t, err)
	assert.Equal(t, "alice@local.test", user.Email)
	assert.Equal(t, "Alice", user.DisplayName)
	assert.Equal(t, int64(1024), user.Quota)
	assert.True(t, user.Active)
	assert.Empty(t, user.PasswordHash) // json:"-" should hide it
}

func TestUsersHandler_Create_DomainNotFound(t *testing.T) {
	h := newUsersHandler(t)

	body := `{"email":"alice@nonexistent.test","password":"TestPass123!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUsersHandler_Create_DuplicateEmail(t *testing.T) {
	h := newUsersHandler(t)

	body := `{"email":"alice@local.test","password":"TestPass123!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)
	assert.Equal(t, http.StatusCreated, rec.Code)

	// Second attempt
	req = httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestUsersHandler_Create_InvalidJSON(t *testing.T) {
	h := newUsersHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString("{invalid"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUsersHandler_Create_MissingFields(t *testing.T) {
	h := newUsersHandler(t)

	body := `{"email":"","password":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestUsersHandler_Get(t *testing.T) {
	h := newUsersHandler(t)

	// Create user first
	body := `{"email":"alice@local.test","password":"TestPass123!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created domain.User
	json.NewDecoder(rec.Body).Decode(&created)

	// Get user
	req = httptest.NewRequest(http.MethodGet, "/api/users/1", nil)
	rec = httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var fetched domain.User
	err := json.NewDecoder(rec.Body).Decode(&fetched)
	require.NoError(t, err)
	assert.Equal(t, "alice@local.test", fetched.Email)
}

func TestUsersHandler_Get_NotFound(t *testing.T) {
	h := newUsersHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/users/9999", nil)
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUsersHandler_List(t *testing.T) {
	h := newUsersHandler(t)

	// Create users
	for _, email := range []string{"alice@local.test", "bob@local.test"} {
		body := `{"email":"` + email + `","password":"TestPass123!"}`
		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.UserListResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Len(t, resp.Users, 2)
	assert.Equal(t, 2, resp.Pagination.Total)
}

func TestUsersHandler_List_DomainFilter(t *testing.T) {
	h := newUsersHandler(t)

	body := `{"email":"alice@local.test","password":"TestPass123!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	req = httptest.NewRequest(http.MethodGet, "/api/users?domain=local.test", nil)
	rec = httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp api.UserListResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Len(t, resp.Users, 1)
}

func TestUsersHandler_Delete(t *testing.T) {
	h := newUsersHandler(t)

	// Create user
	body := `{"email":"alice@local.test","password":"TestPass123!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/api/users/1", nil)
	rec = httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify gone
	req = httptest.NewRequest(http.MethodGet, "/api/users/1", nil)
	rec = httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUsersHandler_Delete_NotFound(t *testing.T) {
	h := newUsersHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/users/9999", nil)
	rec := httptest.NewRecorder()
	h.Router().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
