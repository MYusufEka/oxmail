package logs

import "sync"

// RingBuffer is a thread-safe fixed-capacity circular buffer for log entries.
type RingBuffer struct {
	mu       sync.RWMutex
	entries  []LogEntry
	capacity int
	head     int
	count    int
}

// NewRingBuffer creates a RingBuffer with the given capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		entries:  make([]LogEntry, capacity),
		capacity: capacity,
	}
}

// Add inserts a log entry, evicting the oldest if at capacity.
func (rb *RingBuffer) Add(entry LogEntry) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	idx := (rb.head + rb.count) % rb.capacity
	if rb.count == rb.capacity {
		// Overwrite oldest
		rb.entries[rb.head] = entry
		rb.head = (rb.head + 1) % rb.capacity
	} else {
		rb.entries[idx] = entry
		rb.count++
	}
}

// Len returns the current number of entries in the buffer.
func (rb *RingBuffer) Len() int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()
	return rb.count
}

// Entries returns a slice of entries with offset and limit (oldest first).
func (rb *RingBuffer) Entries(offset, limit int) []LogEntry {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if offset >= rb.count {
		return nil
	}

	end := offset + limit
	if end > rb.count {
		end = rb.count
	}

	result := make([]LogEntry, 0, end-offset)
	for i := offset; i < end; i++ {
		idx := (rb.head + i) % rb.capacity
		result = append(result, rb.entries[idx])
	}
	return result
}

// Filter returns entries matching service and level with pagination.
// Empty service or level means no filter on that field.
func (rb *RingBuffer) Filter(service, level string, offset, limit int) []LogEntry {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	var result []LogEntry
	skipped := 0

	for i := 0; i < rb.count; i++ {
		idx := (rb.head + i) % rb.capacity
		entry := rb.entries[idx]

		if !matchesFilter(entry, service, level) {
			continue
		}

		if skipped < offset {
			skipped++
			continue
		}

		result = append(result, entry)
		if len(result) >= limit {
			break
		}
	}

	return result
}

// FilterCount returns the total count of entries matching the filter.
func (rb *RingBuffer) FilterCount(service, level string) int {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	count := 0
	for i := 0; i < rb.count; i++ {
		idx := (rb.head + i) % rb.capacity
		if matchesFilter(rb.entries[idx], service, level) {
			count++
		}
	}
	return count
}

func matchesFilter(entry LogEntry, service, level string) bool {
	if service != "" && entry.Service != service {
		return false
	}
	if level != "" && entry.Level != level {
		return false
	}
	return true
}
