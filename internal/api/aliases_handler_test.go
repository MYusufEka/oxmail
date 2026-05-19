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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAliasTestServer(t *testing.T) (*api.AliasHandler, *database.DB) {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := database.Open(":memory:", logger)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.Conn.Exec(`INSERT INTO domains (name, active) VALUES ('local.test', 1)`)
	require.NoError(t, err)

	svc := domain.NewAliasService(db.Conn)
	handler := api.NewAliasHandler(svc, nil)
	return handler, db
}

func TestAliasHandler_Create_Success(t *testing.T) {
	handler, _ := setupAliasTestServer(t)
	router := handler.Router()

	body := `{"sourceAddress":"info@local.test","destinationAddress":"alice@local.test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/aliases", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var alias domain.Alias
	err := json.NewDecoder(rec.Body).Decode(&alias)
	require.NoError(t, err)
	assert.Equal(t, "info@local.test", alias.SourceAddress)
	assert.Equal(t, "alice@local.test", alias.DestinationAddress)
	assert.NotZero(t, alias.ID)
}

func TestAliasHandler_Create_InvalidJSON(t *testing.T) {
	handler, _ := setupAliasTestServer(t)
	router := handler.Router()

	req := httptest.NewRequest(http.MethodPost, "/api/aliases", bytes.NewBufferString(`{invalid`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestAliasHandler_Create_CircularAlias(t *testing.T) {
	handler, _ := setupAliasTestServer(t)
	router := handler.Router()

	// Create first alias
	body := `{"sourceAddress":"alice@local.test","destinationAddress":"info@local.test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/aliases", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	// Create circular alias
	body = `{"sourceAddress":"info@local.test","destinationAddress":"alice@local.test"}`
	req = httptest.NewRequest(http.MethodPost, "/api/aliases", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errResp domain.ErrorResponse
	err := json.NewDecoder(rec.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Contains(t, errResp.Error.Message, "circular")
}

func TestAliasHandler_Create_DomainNotFound(t *testing.T) {
	handler, _ := setupAliasTestServer(t)
	router := handler.Router()

	body := `{"sourceAddress":"info@unknown.test","destinationAddress":"alice@local.test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/aliases", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAliasHandler_Create_Duplicate(t *testing.T) {
	handler, _ := setupAliasTestServer(t)
	router := handler.Router()

	body := `{"sourceAddress":"info@local.test","destinationAddress":"alice@local.test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/aliases", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	// Duplicate
	req = httptest.NewRequest(http.MethodPost, "/api/aliases", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestAliasHandler_List(t *testing.T) {
	handler, _ := setupAliasTestServer(t)
	router := handler.Router()

	// Create two aliases
	for _, dest := range []string{"alice@local.test", "bob@local.test"} {
		body := fmt.Sprintf(`{"sourceAddress":"info@local.test","destinationAddress":"%s"}`, dest)
		req := httptest.NewRequest(http.MethodPost, "/api/aliases", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusCreated, rec.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/aliases", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var aliases []domain.Alias
	err := json.NewDecoder(rec.Body).Decode(&aliases)
	require.NoError(t, err)
	assert.Len(t, aliases, 2)
}

func TestAliasHandler_Get(t *testing.T) {
	handler, _ := setupAliasTestServer(t)
	router := handler.Router()

	body := `{"sourceAddress":"info@local.test","destinationAddress":"alice@local.test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/aliases", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created domain.Alias
	json.NewDecoder(rec.Body).Decode(&created)

	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/aliases/%d", created.ID), nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var alias domain.Alias
	err := json.NewDecoder(rec.Body).Decode(&alias)
	require.NoError(t, err)
	assert.Equal(t, created.ID, alias.ID)
}

func TestAliasHandler_Get_NotFound(t *testing.T) {
	handler, _ := setupAliasTestServer(t)
	router := handler.Router()

	req := httptest.NewRequest(http.MethodGet, "/api/aliases/9999", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAliasHandler_Delete(t *testing.T) {
	handler, _ := setupAliasTestServer(t)
	router := handler.Router()

	body := `{"sourceAddress":"info@local.test","destinationAddress":"alice@local.test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/aliases", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)

	var created domain.Alias
	json.NewDecoder(rec.Body).Decode(&created)

	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/aliases/%d", created.ID), nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify deleted
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/aliases/%d", created.ID), nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestAliasHandler_Delete_NotFound(t *testing.T) {
	handler, _ := setupAliasTestServer(t)
	router := handler.Router()

	req := httptest.NewRequest(http.MethodDelete, "/api/aliases/9999", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}
