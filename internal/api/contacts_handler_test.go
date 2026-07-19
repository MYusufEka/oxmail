package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/MYusufEka/oxmail/internal/api"
	"github.com/MYusufEka/oxmail/internal/database"
	"github.com/MYusufEka/oxmail/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type contactsTestFixture struct {
	db      *database.DB
	handler *api.ContactsHandler
}

func setupContactsTest(t *testing.T) *contactsTestFixture {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := database.Open(":memory:", logger)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	svc := domain.NewContactService(db.Conn)
	return &contactsTestFixture{
		db:      db,
		handler: api.NewContactsHandler(svc),
	}
}

func (fixture *contactsTestFixture) Router() *chi.Mux {
	return fixture.handler.Router()
}

func setupContactsTestWithUser(t *testing.T) *contactsTestFixture {
	t.Helper()
	fixture := setupContactsTest(t)
	fixture.seedUser(t, "example.com", "user@example.com")
	return fixture
}

func (fixture *contactsTestFixture) seedUser(t *testing.T, domainName, email string) {
	t.Helper()
	_, err := fixture.db.Conn.Exec("INSERT INTO domains (name, active) VALUES (?, 1)", domainName)
	require.NoError(t, err)

	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	require.NoError(t, err)

	_, err = fixture.db.Conn.Exec(
		"INSERT INTO users (email, password_hash, domain_id, display_name, active) VALUES (?, ?, (SELECT id FROM domains WHERE name = ?), ?, 1)",
		email, string(hash), domainName, "Test User",
	)
	require.NoError(t, err)
}

func TestContactsHandler_List(t *testing.T) {
	t.Run("accepts URL-encoded user email", func(t *testing.T) {
		h := setupContactsTestWithUser(t)

		body := `{"name":"Encoded","email":"encoded@test.com"}`
		req := httptest.NewRequest(http.MethodPost, "/api/contacts/user%40example.com", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		req = httptest.NewRequest(http.MethodGet, "/api/contacts/user%40example.com", nil)
		rec = httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var contacts []domain.Contact
		err := json.NewDecoder(rec.Body).Decode(&contacts)
		require.NoError(t, err)
		require.Len(t, contacts, 1)
		assert.Equal(t, "user@example.com", contacts[0].UserEmail)
		assert.Equal(t, "Encoded", contacts[0].Name)
	})

	t.Run("returns empty list", func(t *testing.T) {
		h := setupContactsTestWithUser(t)

		req := httptest.NewRequest(http.MethodGet, "/api/contacts/user@example.com", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var contacts []domain.Contact
		err := json.NewDecoder(rec.Body).Decode(&contacts)
		require.NoError(t, err)
		assert.Empty(t, contacts)
	})

	t.Run("returns contacts after creation", func(t *testing.T) {
		h := setupContactsTestWithUser(t)

		body := `{"name":"Alice","email":"alice@test.com"}`
		req := httptest.NewRequest(http.MethodPost, "/api/contacts/user@example.com", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		req = httptest.NewRequest(http.MethodGet, "/api/contacts/user@example.com", nil)
		rec = httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var contacts []domain.Contact
		err := json.NewDecoder(rec.Body).Decode(&contacts)
		require.NoError(t, err)
		require.Len(t, contacts, 1)
		assert.Equal(t, "Alice", contacts[0].Name)
		assert.Equal(t, "alice@test.com", contacts[0].Email)
	})

	t.Run("returns 404 for missing email segment", func(t *testing.T) {
		h := setupContactsTestWithUser(t)

		req := httptest.NewRequest(http.MethodGet, "/api/contacts/", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestContactsHandler_Create(t *testing.T) {
	t.Run("creates contact successfully", func(t *testing.T) {
		h := setupContactsTestWithUser(t)

		body := `{"name":"Bob","email":"bob@test.com","phone":"+123456789"}`
		req := httptest.NewRequest(http.MethodPost, "/api/contacts/user@example.com", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var contact domain.Contact
		err := json.NewDecoder(rec.Body).Decode(&contact)
		require.NoError(t, err)
		assert.Equal(t, "Bob", contact.Name)
		assert.Equal(t, "bob@test.com", contact.Email)
		assert.Equal(t, "+123456789", contact.Phone)
		assert.NotZero(t, contact.ID)
	})

	t.Run("returns 409 for duplicate contact", func(t *testing.T) {
		h := setupContactsTestWithUser(t)

		body := `{"name":"Dup","email":"dup@test.com"}`
		req := httptest.NewRequest(http.MethodPost, "/api/contacts/user@example.com", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		req = httptest.NewRequest(http.MethodPost, "/api/contacts/user@example.com", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusConflict, rec.Code)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		h := setupContactsTestWithUser(t)

		req := httptest.NewRequest(http.MethodPost, "/api/contacts/user@example.com", bytes.NewBufferString("{invalid"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 400 for missing name and email", func(t *testing.T) {
		h := setupContactsTestWithUser(t)

		body := `{"name":"","email":""}`
		req := httptest.NewRequest(http.MethodPost, "/api/contacts/user@example.com", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 404 for missing email segment", func(t *testing.T) {
		h := setupContactsTestWithUser(t)

		body := `{"name":"Test","email":"test@test.com"}`
		req := httptest.NewRequest(http.MethodPost, "/api/contacts/", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestContactsHandler_Update(t *testing.T) {
	t.Run("updates contact successfully", func(t *testing.T) {
		h := setupContactsTestWithUser(t)

		body := `{"name":"Original","email":"orig@test.com"}`
		req := httptest.NewRequest(http.MethodPost, "/api/contacts/user@example.com", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		var created domain.Contact
		err := json.NewDecoder(rec.Body).Decode(&created)
		require.NoError(t, err)

		updateBody := `{"name":"Updated","phone":"+987654321"}`
		path := fmt.Sprintf("/api/contacts/user@example.com/%d", created.ID)
		req = httptest.NewRequest(http.MethodPut, path, bytes.NewBufferString(updateBody))
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var updated domain.Contact
		err = json.NewDecoder(rec.Body).Decode(&updated)
		require.NoError(t, err)
		assert.Equal(t, "Updated", updated.Name)
		assert.Equal(t, "orig@test.com", updated.Email)
		assert.Equal(t, "+987654321", updated.Phone)
	})

	t.Run("returns 404 for non-existent contact", func(t *testing.T) {
		h := setupContactsTestWithUser(t)

		body := `{"name":"Ghost","email":"ghost@test.com"}`
		req := httptest.NewRequest(http.MethodPut, "/api/contacts/user@example.com/99999", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("returns 409 when update duplicates another contact email", func(t *testing.T) {
		h := setupContactsTestWithUser(t)

		firstBody := `{"name":"First","email":"first@test.com"}`
		req := httptest.NewRequest(http.MethodPost, "/api/contacts/user@example.com", bytes.NewBufferString(firstBody))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		secondBody := `{"name":"Second","email":"second@test.com"}`
		req = httptest.NewRequest(http.MethodPost, "/api/contacts/user@example.com", bytes.NewBufferString(secondBody))
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		var second domain.Contact
		err := json.NewDecoder(rec.Body).Decode(&second)
		require.NoError(t, err)

		updateBody := `{"email":"first@test.com"}`
		path := fmt.Sprintf("/api/contacts/user@example.com/%d", second.ID)
		req = httptest.NewRequest(http.MethodPut, path, bytes.NewBufferString(updateBody))
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusConflict, rec.Code)

		var resp domain.ErrorResponse
		err = json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "contact_exists", resp.Error.Code)
	})

	t.Run("returns 400 for invalid contact ID", func(t *testing.T) {
		h := setupContactsTestWithUser(t)

		body := `{"name":"Test","email":"t@t.com"}`
		req := httptest.NewRequest(http.MethodPut, "/api/contacts/user@example.com/abc", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		h := setupContactsTestWithUser(t)

		req := httptest.NewRequest(http.MethodPut, "/api/contacts/user@example.com/1", bytes.NewBufferString("{invalid"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestContactsHandler_Delete(t *testing.T) {
	t.Run("deletes contact successfully", func(t *testing.T) {
		h := setupContactsTestWithUser(t)

		body := `{"name":"DeleteMe","email":"delete@test.com"}`
		req := httptest.NewRequest(http.MethodPost, "/api/contacts/user@example.com", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)

		var created domain.Contact
		err := json.NewDecoder(rec.Body).Decode(&created)
		require.NoError(t, err)

		path := fmt.Sprintf("/api/contacts/user@example.com/%d", created.ID)
		req = httptest.NewRequest(http.MethodDelete, path, nil)
		rec = httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]string
		err = json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "deleted", resp["status"])
	})

	t.Run("returns 404 for non-existent contact", func(t *testing.T) {
		h := setupContactsTestWithUser(t)

		req := httptest.NewRequest(http.MethodDelete, "/api/contacts/user@example.com/99999", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("returns 400 for invalid contact ID", func(t *testing.T) {
		h := setupContactsTestWithUser(t)

		req := httptest.NewRequest(http.MethodDelete, "/api/contacts/user@example.com/abc", nil)
		rec := httptest.NewRecorder()
		h.Router().ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
