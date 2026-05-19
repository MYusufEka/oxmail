package logs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRingBufferAdd(t *testing.T) {
	buf := NewRingBuffer(5)

	for i := 0; i < 5; i++ {
		buf.Add(LogEntry{
			ID:      int64(i + 1),
			Message: "msg",
			Service: "postfix",
		})
	}

	entries := buf.Entries(0, 10)
	require.Len(t, entries, 5)
	assert.Equal(t, int64(1), entries[0].ID)
	assert.Equal(t, int64(5), entries[4].ID)
}

func TestRingBufferEviction(t *testing.T) {
	buf := NewRingBuffer(3)

	for i := 0; i < 5; i++ {
		buf.Add(LogEntry{ID: int64(i + 1), Message: "msg"})
	}

	entries := buf.Entries(0, 10)
	require.Len(t, entries, 3)
	// Oldest (1, 2) evicted; remaining: 3, 4, 5
	assert.Equal(t, int64(3), entries[0].ID)
	assert.Equal(t, int64(5), entries[2].ID)
}

func TestRingBufferPagination(t *testing.T) {
	buf := NewRingBuffer(10)

	for i := 0; i < 10; i++ {
		buf.Add(LogEntry{ID: int64(i + 1), Message: "msg"})
	}

	// First page
	page1 := buf.Entries(0, 3)
	require.Len(t, page1, 3)
	assert.Equal(t, int64(1), page1[0].ID)
	assert.Equal(t, int64(3), page1[2].ID)

	// Second page
	page2 := buf.Entries(3, 3)
	require.Len(t, page2, 3)
	assert.Equal(t, int64(4), page2[0].ID)
	assert.Equal(t, int64(6), page2[2].ID)

	// Beyond end
	page4 := buf.Entries(9, 5)
	require.Len(t, page4, 1)
	assert.Equal(t, int64(10), page4[0].ID)

	// Way beyond
	empty := buf.Entries(20, 5)
	assert.Empty(t, empty)
}

func TestRingBufferFilter(t *testing.T) {
	buf := NewRingBuffer(10)

	buf.Add(LogEntry{ID: 1, Service: "postfix", Level: "info"})
	buf.Add(LogEntry{ID: 2, Service: "dovecot", Level: "error"})
	buf.Add(LogEntry{ID: 3, Service: "postfix", Level: "error"})
	buf.Add(LogEntry{ID: 4, Service: "dovecot", Level: "info"})
	buf.Add(LogEntry{ID: 5, Service: "postfix", Level: "warning"})

	// Filter by service
	postfix := buf.Filter("postfix", "", 0, 10)
	require.Len(t, postfix, 3)
	assert.Equal(t, int64(1), postfix[0].ID)

	// Filter by level
	errors := buf.Filter("", "error", 0, 10)
	require.Len(t, errors, 2)
	assert.Equal(t, int64(2), errors[0].ID)
	assert.Equal(t, int64(3), errors[1].ID)

	// Filter by both
	postfixErrors := buf.Filter("postfix", "error", 0, 10)
	require.Len(t, postfixErrors, 1)
	assert.Equal(t, int64(3), postfixErrors[0].ID)

	// Filter with pagination
	postfixPaged := buf.Filter("postfix", "", 1, 2)
	require.Len(t, postfixPaged, 2)
	assert.Equal(t, int64(3), postfixPaged[0].ID)
	assert.Equal(t, int64(5), postfixPaged[1].ID)
}

func TestRingBufferLen(t *testing.T) {
	buf := NewRingBuffer(5)
	assert.Equal(t, 0, buf.Len())

	buf.Add(LogEntry{ID: 1})
	assert.Equal(t, 1, buf.Len())

	for i := 2; i <= 7; i++ {
		buf.Add(LogEntry{ID: int64(i)})
	}
	assert.Equal(t, 5, buf.Len())
}

func TestRingBufferConcurrency(t *testing.T) {
	buf := NewRingBuffer(1000)
	done := make(chan struct{})

	// Writer goroutine
	go func() {
		for i := 0; i < 5000; i++ {
			buf.Add(LogEntry{
				ID:        int64(i),
				Timestamp: time.Now(),
				Service:   "postfix",
				Message:   "concurrent write",
			})
		}
		close(done)
	}()

	// Reader goroutine — should not panic
	for i := 0; i < 100; i++ {
		_ = buf.Entries(0, 50)
		_ = buf.Filter("postfix", "", 0, 10)
	}

	<-done

	// Buffer should have exactly 1000 entries (capacity)
	assert.Equal(t, 1000, buf.Len())
}

func TestRingBufferFilterCount(t *testing.T) {
	buf := NewRingBuffer(10)

	buf.Add(LogEntry{ID: 1, Service: "postfix", Level: "info"})
	buf.Add(LogEntry{ID: 2, Service: "dovecot", Level: "error"})
	buf.Add(LogEntry{ID: 3, Service: "postfix", Level: "error"})

	assert.Equal(t, 3, buf.FilterCount("", ""))
	assert.Equal(t, 2, buf.FilterCount("postfix", ""))
	assert.Equal(t, 2, buf.FilterCount("", "error"))
	assert.Equal(t, 1, buf.FilterCount("postfix", "error"))
}
