package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"

	"github.com/MYusufEka/oxmail/internal/logs"
)

// LogsResponse is the JSON envelope for paginated log entries.
type LogsResponse struct {
	Entries []logs.LogEntry `json:"entries"`
	Total   int            `json:"total"`
	Limit   int            `json:"limit"`
	Offset  int            `json:"offset"`
}

// StreamFilter is the filter message sent by WebSocket clients.
type StreamFilter struct {
	Service string `json:"service"`
	Level   string `json:"level"`
}

// LogsHandler handles log-related HTTP and WebSocket endpoints.
type LogsHandler struct {
	buffer    *logs.RingBuffer
	collector *logs.Collector
	upgrader  websocket.Upgrader
}

// NewLogsHandler creates a LogsHandler.
func NewLogsHandler(buffer *logs.RingBuffer, collector *logs.Collector) *LogsHandler {
	return &LogsHandler{
		buffer:    buffer,
		collector: collector,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool { return true },
		},
	}
}

// HandleGetLogs serves GET /api/logs with optional filtering and pagination.
func (h *LogsHandler) HandleGetLogs(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Query().Get("service")
	level := r.URL.Query().Get("level")
	limit := parseIntParam(r, "limit", 50)
	offset := parseIntParam(r, "offset", 0)

	var entries []logs.LogEntry
	var total int

	if service == "" && level == "" {
		entries = h.buffer.Entries(offset, limit)
		total = h.buffer.Len()
	} else {
		entries = h.buffer.Filter(service, level, offset, limit)
		total = h.buffer.FilterCount(service, level)
	}

	if entries == nil {
		entries = []logs.LogEntry{}
	}

	resp := LogsResponse{
		Entries: entries,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// HandleStream upgrades to WebSocket and streams log entries to the client.
func (h *LogsHandler) HandleStream(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	ch := h.collector.Subscribe()
	defer h.collector.Unsubscribe(ch)

	var filter StreamFilter
	var filterMu sync.RWMutex

	// Read filter messages from client
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}

			var newFilter StreamFilter
			if json.Unmarshal(msg, &newFilter) == nil {
				filterMu.Lock()
				filter = newFilter
				filterMu.Unlock()
			}
		}
	}()

	// Stream entries to client
	for entry := range ch {
		filterMu.RLock()
		currentFilter := filter
		filterMu.RUnlock()

		if !matchesStreamFilter(entry, currentFilter) {
			continue
		}

		if err := conn.WriteJSON(entry); err != nil {
			return
		}
	}
}

func matchesStreamFilter(entry logs.LogEntry, filter StreamFilter) bool {
	if filter.Service != "" && entry.Service != filter.Service {
		return false
	}
	if filter.Level != "" && entry.Level != filter.Level {
		return false
	}
	return true
}

func parseIntParam(r *http.Request, key string, defaultVal int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(raw)
	if err != nil || val < 0 {
		return defaultVal
	}
	return val
}
