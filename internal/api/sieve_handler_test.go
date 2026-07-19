package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MYusufEka/oxmail/internal/api"
	"github.com/MYusufEka/oxmail/internal/mail"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCmdExecutor implements mail.CommandExecutor for testing sieve operations.
type mockCmdExecutor struct {
	mail.CommandExecutor
	onRun           func(name string, args ...string) error
	onRunWithOutput func(name string, args ...string) (string, error)
}

func (m *mockCmdExecutor) Run(name string, args ...string) error {
	if m.onRun != nil {
		return m.onRun(name, args...)
	}
	return nil
}

func (m *mockCmdExecutor) RunWithOutput(name string, args ...string) (string, error) {
	if m.onRunWithOutput != nil {
		return m.onRunWithOutput(name, args...)
	}
	return "", nil
}

func setupSieveHandler(t *testing.T, exec mail.CommandExecutor) (*api.SieveHandler, *chi.Mux) {
	t.Helper()
	manager := mail.NewSieveManager("/tmp/sieve", "/tmp/global", exec)
	handler := api.NewSieveHandler(manager)
	r := chi.NewRouter()
	handler.RegisterRoutes(r)
	return handler, r
}

func TestSieveHandler_GetFilter(t *testing.T) {
	t.Run("returns existing script", func(t *testing.T) {
		exec := &mockCmdExecutor{
			onRun: func(name string, args ...string) error {
				if name == "test" {
					return nil // file exists
				}
				return nil
			},
			onRunWithOutput: func(name string, args ...string) (string, error) {
				return "require [\"fileinto\"];\nfileinto \"Test\";", nil
			},
		}
		_, router := setupSieveHandler(t, exec)

		req := httptest.NewRequest(http.MethodGet, "/api/mail/filters/test@example.com", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Email  string `json:"email"`
			Script string `json:"script"`
			Active bool   `json:"active"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", resp.Email)
		assert.Contains(t, resp.Script, "fileinto")
		assert.True(t, resp.Active)
	})

	t.Run("returns empty when no script exists", func(t *testing.T) {
		exec := &mockCmdExecutor{
			onRun: func(name string, args ...string) error {
				if name == "test" {
					return fmt.Errorf("file not found")
				}
				return nil
			},
		}
		_, router := setupSieveHandler(t, exec)

		req := httptest.NewRequest(http.MethodGet, "/api/mail/filters/empty@example.com", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Email  string `json:"email"`
			Script string `json:"script"`
			Active bool   `json:"active"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "empty@example.com", resp.Email)
		assert.Empty(t, resp.Script)
		assert.False(t, resp.Active)
	})

	t.Run("returns 404 for missing email segment", func(t *testing.T) {
		exec := &mockCmdExecutor{}
		_, router := setupSieveHandler(t, exec)

		req := httptest.NewRequest(http.MethodGet, "/api/mail/filters/", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestSieveHandler_SetFilter(t *testing.T) {
	t.Run("sets script successfully", func(t *testing.T) {
		exec := &mockCmdExecutor{}
		_, router := setupSieveHandler(t, exec)

		body := `{"script":"require [\"fileinto\"];\nfileinto \"Inbox\";"}`
		req := httptest.NewRequest(http.MethodPost, "/api/mail/filters/test@example.com", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Email  string `json:"email"`
			Status string `json:"status"`
			Active bool   `json:"active"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", resp.Email)
		assert.Equal(t, "saved", resp.Status)
		assert.True(t, resp.Active)
	})

	t.Run("returns 400 for empty script body", func(t *testing.T) {
		exec := &mockCmdExecutor{}
		_, router := setupSieveHandler(t, exec)

		body := `{"script":""}`
		req := httptest.NewRequest(http.MethodPost, "/api/mail/filters/test@example.com", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		exec := &mockCmdExecutor{}
		_, router := setupSieveHandler(t, exec)

		req := httptest.NewRequest(http.MethodPost, "/api/mail/filters/test@example.com", bytes.NewBufferString("{invalid"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 404 for missing email segment", func(t *testing.T) {
		exec := &mockCmdExecutor{}
		_, router := setupSieveHandler(t, exec)

		body := `{"script":"require [\"fileinto\"];"}`
		req := httptest.NewRequest(http.MethodPost, "/api/mail/filters/", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestSieveHandler_DeleteFilter(t *testing.T) {
	t.Run("deletes script successfully", func(t *testing.T) {
		removedPaths := make([]string, 0, 2)
		exec := &mockCmdExecutor{
			onRun: func(name string, args ...string) error {
				if name == "rm" {
					removedPaths = append(removedPaths, args[len(args)-1])
				}
				return nil
			},
		}
		_, router := setupSieveHandler(t, exec)

		req := httptest.NewRequest(http.MethodDelete, "/api/mail/filters/test@example.com", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.ElementsMatch(t, []string{
			"/tmp/sieve/example.com/test.sieve",
			"/tmp/sieve/example.com/test.sieve.svbin",
		}, removedPaths)

		var resp struct {
			Email  string `json:"email"`
			Status string `json:"status"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", resp.Email)
		assert.Equal(t, "deleted", resp.Status)
	})

	t.Run("decodes URL-encoded email path", func(t *testing.T) {
		removedPaths := make([]string, 0, 2)
		exec := &mockCmdExecutor{
			onRun: func(name string, args ...string) error {
				if name == "rm" {
					removedPaths = append(removedPaths, args[len(args)-1])
				}
				return nil
			},
		}
		_, router := setupSieveHandler(t, exec)

		req := httptest.NewRequest(http.MethodDelete, "/api/mail/filters/user%2Btag@example.com", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.ElementsMatch(t, []string{
			"/tmp/sieve/example.com/user+tag.sieve",
			"/tmp/sieve/example.com/user+tag.sieve.svbin",
		}, removedPaths)

		var resp struct {
			Email  string `json:"email"`
			Status string `json:"status"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "user+tag@example.com", resp.Email)
		assert.Equal(t, "deleted", resp.Status)
	})

	t.Run("returns 404 for missing email segment", func(t *testing.T) {
		exec := &mockCmdExecutor{}
		_, router := setupSieveHandler(t, exec)

		req := httptest.NewRequest(http.MethodDelete, "/api/mail/filters/", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestSieveHandler_GetVacation(t *testing.T) {
	t.Run("returns existing vacation", func(t *testing.T) {
		exec := &mockCmdExecutor{
			onRun: func(name string, args ...string) error {
				if name == "test" {
					return nil // file exists
				}
				return nil
			},
			onRunWithOutput: func(name string, args ...string) (string, error) {
				return "require [\"vacation\"];\nvacation :days 1 :subject \"OOO\" \"Away\";", nil
			},
		}
		_, router := setupSieveHandler(t, exec)

		req := httptest.NewRequest(http.MethodGet, "/api/mail/vacation/user@example.com", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Email   string `json:"email"`
			Subject string `json:"subject"`
			Body    string `json:"body"`
			Enabled bool   `json:"enabled"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "user@example.com", resp.Email)
		assert.True(t, resp.Enabled)
	})

	t.Run("returns 500 when existing vacation read fails", func(t *testing.T) {
		exec := &mockCmdExecutor{
			onRun: func(name string, args ...string) error {
				if name == "test" {
					return nil
				}
				return nil
			},
			onRunWithOutput: func(name string, args ...string) (string, error) {
				return "", fmt.Errorf("read failed")
			},
		}
		_, router := setupSieveHandler(t, exec)

		req := httptest.NewRequest(http.MethodGet, "/api/mail/vacation/user@example.com", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "internal_error")
	})

	t.Run("returns disabled when no vacation exists", func(t *testing.T) {
		exec := &mockCmdExecutor{
			onRun: func(name string, args ...string) error {
				if name == "test" {
					return fmt.Errorf("file not found")
				}
				return nil
			},
		}
		_, router := setupSieveHandler(t, exec)

		req := httptest.NewRequest(http.MethodGet, "/api/mail/vacation/user@example.com", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Email   string `json:"email"`
			Enabled bool   `json:"enabled"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.False(t, resp.Enabled)
	})
}

func TestSieveHandler_SetVacation(t *testing.T) {
	t.Run("enables vacation", func(t *testing.T) {
		commands := make([]string, 0, 3)
		exec := &mockCmdExecutor{
			onRun: func(name string, args ...string) error {
				commands = append(commands, name)
				return nil
			},
		}
		_, router := setupSieveHandler(t, exec)

		body := `{"subject":"Out of Office","body":"I am away","enabled":true}`
		req := httptest.NewRequest(http.MethodPost, "/api/mail/vacation/user@example.com", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, []string{"mkdir", "sh", "sievec"}, commands)

		var resp struct {
			Email   string `json:"email"`
			Enabled bool   `json:"enabled"`
			Status  string `json:"status"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.True(t, resp.Enabled)
		assert.Equal(t, "enabled", resp.Status)
	})

	t.Run("returns 500 when compile fails", func(t *testing.T) {
		exec := &mockCmdExecutor{
			onRun: func(name string, args ...string) error {
				if name == "sievec" {
					return fmt.Errorf("compile failed")
				}
				return nil
			},
		}
		_, router := setupSieveHandler(t, exec)

		body := `{"subject":"Out of Office","body":"I am away","enabled":true}`
		req := httptest.NewRequest(http.MethodPost, "/api/mail/vacation/user@example.com", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Contains(t, rec.Body.String(), "internal_error")
	})

	t.Run("disables vacation", func(t *testing.T) {
		exec := &mockCmdExecutor{}
		_, router := setupSieveHandler(t, exec)

		body := `{"enabled":false}`
		req := httptest.NewRequest(http.MethodPost, "/api/mail/vacation/user@example.com", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var resp struct {
			Email   string `json:"email"`
			Enabled bool   `json:"enabled"`
			Status  string `json:"status"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.False(t, resp.Enabled)
		assert.Equal(t, "disabled", resp.Status)
	})

	t.Run("returns 400 when enabled without subject", func(t *testing.T) {
		exec := &mockCmdExecutor{}
		_, router := setupSieveHandler(t, exec)

		body := `{"enabled":true}`
		req := httptest.NewRequest(http.MethodPost, "/api/mail/vacation/user@example.com", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("returns 400 for invalid JSON", func(t *testing.T) {
		exec := &mockCmdExecutor{}
		_, router := setupSieveHandler(t, exec)

		req := httptest.NewRequest(http.MethodPost, "/api/mail/vacation/user@example.com", bytes.NewBufferString("{invalid"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestSieveHandler_DeleteVacation(t *testing.T) {
	t.Run("deletes vacation successfully", func(t *testing.T) {
		removedPaths := make([]string, 0, 2)
		exec := &mockCmdExecutor{
			onRun: func(name string, args ...string) error {
				if name == "rm" {
					removedPaths = append(removedPaths, args[len(args)-1])
				}
				return nil
			},
		}
		_, router := setupSieveHandler(t, exec)

		req := httptest.NewRequest(http.MethodDelete, "/api/mail/vacation/user@example.com", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.ElementsMatch(t, []string{
			"/tmp/sieve/example.com/user.vacation.sieve",
			"/tmp/sieve/example.com/user.vacation.sieve.svbin",
		}, removedPaths)

		var resp struct {
			Email   string `json:"email"`
			Enabled bool   `json:"enabled"`
			Status  string `json:"status"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "user@example.com", resp.Email)
		assert.False(t, resp.Enabled)
		assert.Equal(t, "deleted", resp.Status)
	})

	t.Run("decodes URL-encoded email path", func(t *testing.T) {
		removedPaths := make([]string, 0, 2)
		exec := &mockCmdExecutor{
			onRun: func(name string, args ...string) error {
				if name == "rm" {
					removedPaths = append(removedPaths, args[len(args)-1])
				}
				return nil
			},
		}
		_, router := setupSieveHandler(t, exec)

		req := httptest.NewRequest(http.MethodDelete, "/api/mail/vacation/user%2Btag@example.com", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.ElementsMatch(t, []string{
			"/tmp/sieve/example.com/user+tag.vacation.sieve",
			"/tmp/sieve/example.com/user+tag.vacation.sieve.svbin",
		}, removedPaths)

		var resp struct {
			Email   string `json:"email"`
			Enabled bool   `json:"enabled"`
			Status  string `json:"status"`
		}
		err := json.NewDecoder(rec.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "user+tag@example.com", resp.Email)
		assert.False(t, resp.Enabled)
		assert.Equal(t, "deleted", resp.Status)
	})

	t.Run("returns 404 for missing email segment", func(t *testing.T) {
		exec := &mockCmdExecutor{}
		_, router := setupSieveHandler(t, exec)

		req := httptest.NewRequest(http.MethodDelete, "/api/mail/vacation/", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestSieveHandler_ErrorPaths(t *testing.T) {
	t.Run("get filter - 404 for missing email segment", func(t *testing.T) {
		exec := &mockCmdExecutor{}
		_, router := setupSieveHandler(t, exec)

		req := httptest.NewRequest(http.MethodGet, "/api/mail/filters/", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("set vacation - 404 for missing email segment", func(t *testing.T) {
		exec := &mockCmdExecutor{}
		_, router := setupSieveHandler(t, exec)

		body := `{"subject":"OOO","body":"away","enabled":true}`
		req := httptest.NewRequest(http.MethodPost, "/api/mail/vacation/", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}
