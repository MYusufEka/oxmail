package logs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectorTailsNewLines(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "mail.log")

	// Create file with initial content
	err := os.WriteFile(logFile, []byte("May 19 10:00:00 mail postfix/smtpd[100]: connect from localhost[127.0.0.1]\n"), 0644)
	require.NoError(t, err)

	buf := NewRingBuffer(100)
	collector := NewCollector([]string{logFile}, buf, NewParser())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	collector.Start(ctx)

	// Wait for initial line to be processed
	assert.Eventually(t, func() bool {
		return buf.Len() >= 1
	}, 2*time.Second, 50*time.Millisecond)

	// Append a new line
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
	require.NoError(t, err)
	_, err = f.WriteString("May 19 10:00:01 mail postfix/smtp[200]: ABC123: to=<user@test.com>, status=sent (250 OK)\n")
	require.NoError(t, err)
	f.Close()

	// Wait for new line
	assert.Eventually(t, func() bool {
		return buf.Len() >= 2
	}, 2*time.Second, 50*time.Millisecond)

	entries := buf.Entries(0, 10)
	assert.Equal(t, "postfix", entries[0].Service)
	assert.Equal(t, "smtpd", entries[0].Component)
	assert.Equal(t, "postfix", entries[1].Service)
	assert.Equal(t, "sent", entries[1].Status)
}

func TestCollectorMultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "postfix.log")
	file2 := filepath.Join(tmpDir, "dovecot.log")

	err := os.WriteFile(file1, []byte("May 19 10:00:00 mail postfix/smtpd[100]: connect from localhost[127.0.0.1]\n"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(file2, []byte("May 19 10:00:00 mail dovecot: imap-login: Login: user=<test@local.test>, method=PLAIN, rip=127.0.0.1\n"), 0644)
	require.NoError(t, err)

	buf := NewRingBuffer(100)
	collector := NewCollector([]string{file1, file2}, buf, NewParser())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	collector.Start(ctx)

	assert.Eventually(t, func() bool {
		return buf.Len() >= 2
	}, 2*time.Second, 50*time.Millisecond)

	entries := buf.Entries(0, 10)
	services := map[string]bool{}
	for _, e := range entries {
		services[e.Service] = true
	}
	assert.True(t, services["postfix"])
	assert.True(t, services["dovecot"])
}

func TestCollectorSkipsInvalidLines(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "mail.log")

	content := "not a valid log line\nMay 19 10:00:00 mail postfix/smtpd[100]: valid line\nalso invalid\n"
	err := os.WriteFile(logFile, []byte(content), 0644)
	require.NoError(t, err)

	buf := NewRingBuffer(100)
	collector := NewCollector([]string{logFile}, buf, NewParser())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	collector.Start(ctx)

	assert.Eventually(t, func() bool {
		return buf.Len() >= 1
	}, 2*time.Second, 50*time.Millisecond)

	// Only the valid line should be in the buffer
	time.Sleep(200 * time.Millisecond)
	assert.Equal(t, 1, buf.Len())
}

func TestCollectorContextCancellation(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "mail.log")

	err := os.WriteFile(logFile, []byte("May 19 10:00:00 mail postfix/smtpd[100]: test\n"), 0644)
	require.NoError(t, err)

	buf := NewRingBuffer(100)
	collector := NewCollector([]string{logFile}, buf, NewParser())

	ctx, cancel := context.WithCancel(context.Background())
	collector.Start(ctx)

	assert.Eventually(t, func() bool {
		return buf.Len() >= 1
	}, 2*time.Second, 50*time.Millisecond)

	cancel()

	// Give goroutines time to stop
	time.Sleep(200 * time.Millisecond)

	// Append after cancel — should NOT be collected
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
	require.NoError(t, err)
	_, err = f.WriteString("May 19 10:00:01 mail postfix/smtpd[100]: after cancel\n")
	require.NoError(t, err)
	f.Close()

	time.Sleep(300 * time.Millisecond)
	assert.Equal(t, 1, buf.Len())
}

func TestCollectorSubscribe(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "mail.log")

	err := os.WriteFile(logFile, []byte(""), 0644)
	require.NoError(t, err)

	buf := NewRingBuffer(100)
	collector := NewCollector([]string{logFile}, buf, NewParser())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ch := collector.Subscribe()
	defer collector.Unsubscribe(ch)

	collector.Start(ctx)

	// Write a line
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
	require.NoError(t, err)
	_, err = f.WriteString("May 19 10:00:00 mail postfix/smtpd[100]: subscriber test\n")
	require.NoError(t, err)
	f.Close()

	select {
	case entry := <-ch:
		assert.Equal(t, "postfix", entry.Service)
		assert.Contains(t, entry.Message, "subscriber test")
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for subscriber notification")
	}
}
