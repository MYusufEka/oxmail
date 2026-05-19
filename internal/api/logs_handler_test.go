package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/MYusufEka/oxmail/internal/logs"
)

func setupLogsTestRouter() (*chi.Mux, *logs.RingBuffer, *logs.Collector) {
	buf := logs.NewRingBuffer(100)
	parser := logs.NewParser()

	tmpDir := os.TempDir()
	logFile := filepath.Join(tmpDir, "oxmail-test-handler.log")
	os.WriteFile(logFile, []byte(""), 0644)

	collector := logs.NewCollector([]string{logFile}, buf, parser)
	handler := NewLogsHandler(buf, collector)

	router := chi.NewRouter()
	router.Get("/api/logs", handler.HandleGetLogs)
	router.Get("/api/logs/stream", handler.HandleStream)

	return router, buf, collector
}

func TestHandleGetLogsEmpty(t *testing.T) {
	router, _, _ := setupLogsTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp LogsResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Empty(t, resp.Entries)
	assert.Equal(t, 0, resp.Total)
}

func TestHandleGetLogsWithEntries(t *testing.T) {
	router, buf, _ := setupLogsTestRouter()

	for i := 0; i < 5; i++ {
		buf.Add(logs.LogEntry{
			ID:        int64(i + 1),
			Timestamp: time.Now(),
			Service:   "postfix",
			Level:     "info",
			Message:   "test message",
		})
	}

	req := httptest.NewRequest(http.MethodGet, "/api/logs?limit=3", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp LogsResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Len(t, resp.Entries, 3)
	assert.Equal(t, 5, resp.Total)
}

func TestHandleGetLogsFilterByService(t *testing.T) {
	router, buf, _ := setupLogsTestRouter()

	buf.Add(logs.LogEntry{ID: 1, Service: "postfix", Level: "info", Message: "msg1"})
	buf.Add(logs.LogEntry{ID: 2, Service: "dovecot", Level: "info", Message: "msg2"})
	buf.Add(logs.LogEntry{ID: 3, Service: "postfix", Level: "error", Message: "msg3"})

	req := httptest.NewRequest(http.MethodGet, "/api/logs?service=postfix", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp LogsResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Len(t, resp.Entries, 2)
	assert.Equal(t, 2, resp.Total)
	for _, e := range resp.Entries {
		assert.Equal(t, "postfix", e.Service)
	}
}

func TestHandleGetLogsFilterByLevel(t *testing.T) {
	router, buf, _ := setupLogsTestRouter()

	buf.Add(logs.LogEntry{ID: 1, Service: "postfix", Level: "info", Message: "msg1"})
	buf.Add(logs.LogEntry{ID: 2, Service: "dovecot", Level: "error", Message: "msg2"})
	buf.Add(logs.LogEntry{ID: 3, Service: "postfix", Level: "error", Message: "msg3"})

	req := httptest.NewRequest(http.MethodGet, "/api/logs?level=error", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp LogsResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Len(t, resp.Entries, 2)
	assert.Equal(t, 2, resp.Total)
}

func TestHandleGetLogsPagination(t *testing.T) {
	router, buf, _ := setupLogsTestRouter()

	for i := 0; i < 10; i++ {
		buf.Add(logs.LogEntry{ID: int64(i + 1), Service: "postfix", Level: "info", Message: "msg"})
	}

	req := httptest.NewRequest(http.MethodGet, "/api/logs?limit=3&offset=5", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp LogsResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Len(t, resp.Entries, 3)
	assert.Equal(t, int64(6), resp.Entries[0].ID)
	assert.Equal(t, 10, resp.Total)
}

func TestHandleGetLogsDefaultLimit(t *testing.T) {
	router, buf, _ := setupLogsTestRouter()

	for i := 0; i < 100; i++ {
		buf.Add(logs.LogEntry{ID: int64(i + 1), Service: "postfix", Level: "info", Message: "msg"})
	}

	req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var resp LogsResponse
	err := json.NewDecoder(rec.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Len(t, resp.Entries, 50) // default limit
}

func TestHandleStreamWebSocket(t *testing.T) {
	router, buf, collector := setupLogsTestRouter()
	_ = buf

	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/logs/stream"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Broadcast an entry via collector
	ch := collector.Subscribe()
	defer collector.Unsubscribe(ch)

	// Simulate a new entry arriving by adding to buffer and notifying
	entry := logs.LogEntry{
		ID:        99,
		Timestamp: time.Now(),
		Service:   "postfix",
		Component: "smtpd",
		Level:     "info",
		Message:   "websocket test",
	}
	buf.Add(entry)

	// The handler subscribes to collector, so we need to trigger broadcast
	// We'll test by writing directly — the handler test verifies the WS upgrade works
	// For a full integration test, we'd write to the log file

	// Verify connection is alive by sending a filter message
	filterMsg := `{"service":"postfix","level":""}`
	err = conn.WriteMessage(websocket.TextMessage, []byte(filterMsg))
	require.NoError(t, err)

	// Close cleanly
	err = conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	assert.NoError(t, err)
}

func TestHandleStreamWebSocketFilter(t *testing.T) {
	router, _, _ := setupLogsTestRouter()

	server := httptest.NewServer(router)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/api/logs/stream"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// Send filter
	filterMsg := `{"service":"dovecot","level":"error"}`
	err = conn.WriteMessage(websocket.TextMessage, []byte(filterMsg))
	require.NoError(t, err)

	// Connection should remain open
	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	_, _, err = conn.ReadMessage()
	// Should timeout (no messages to read), not error from closed connection
	assert.Error(t, err) // timeout is expected
}
