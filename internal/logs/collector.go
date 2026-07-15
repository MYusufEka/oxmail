package logs

import (
	"bufio"
	"context"
	"io"
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// Collector tails log files and feeds parsed entries into a RingBuffer.
type Collector struct {
	paths       []string
	buffer      *RingBuffer
	parser      *Parser
	subscribers map[chan LogEntry]struct{}
	subMu       sync.RWMutex
	nextID      atomic.Int64
}

// NewCollector creates a Collector that watches the given file paths.
func NewCollector(paths []string, buffer *RingBuffer, parser *Parser) *Collector {
	return &Collector{
		paths:       paths,
		buffer:      buffer,
		parser:      parser,
		subscribers: make(map[chan LogEntry]struct{}),
	}
}

// Subscribe returns a channel that receives new log entries as they arrive.
func (c *Collector) Subscribe() chan LogEntry {
	ch := make(chan LogEntry, 64)
	c.subMu.Lock()
	c.subscribers[ch] = struct{}{}
	c.subMu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber channel.
func (c *Collector) Unsubscribe(ch chan LogEntry) {
	c.subMu.Lock()
	delete(c.subscribers, ch)
	c.subMu.Unlock()
	close(ch)
}

func (c *Collector) Emit(entry LogEntry) {
	entry.ID = c.nextID.Add(1)
	c.buffer.Add(entry)
	c.broadcast(entry)
}
func (c *Collector) Start(ctx context.Context) {
	for _, path := range c.paths {
		go c.tailFile(ctx, path)
	}
}

func (c *Collector) tailFile(ctx context.Context, path string) {
	f, err := os.Open(path)
	if err != nil {
		slog.Error("failed to open log file", "path", path, "error", err)
		return
	}
	defer f.Close()

	reader := bufio.NewReader(f)

	// Read existing content first
	c.readLines(reader)

	// Poll for new content
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.readLines(reader)
		}
	}
}

func (c *Collector) readLines(reader *bufio.Reader) {
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			// Trim trailing newline
			trimmed := line
			if len(trimmed) > 0 && trimmed[len(trimmed)-1] == '\n' {
				trimmed = trimmed[:len(trimmed)-1]
			}
			if len(trimmed) > 0 && trimmed[len(trimmed)-1] == '\r' {
				trimmed = trimmed[:len(trimmed)-1]
			}

			if len(trimmed) == 0 {
				continue
			}

			entry, parseErr := c.parser.Parse(trimmed)
			if parseErr != nil {
				// Skip unparseable lines
				continue
			}

			entry.ID = c.nextID.Add(1)
			c.buffer.Add(entry)
			c.broadcast(entry)
		}

		if err == io.EOF {
			return
		}
		if err != nil {
			slog.Error("error reading log file", "error", err)
			return
		}
	}
}

func (c *Collector) broadcast(entry LogEntry) {
	c.subMu.RLock()
	defer c.subMu.RUnlock()

	for ch := range c.subscribers {
		select {
		case ch <- entry:
		default:
			// Drop if subscriber is slow
		}
	}
}
