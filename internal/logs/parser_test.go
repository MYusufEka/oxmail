package logs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePostfixLog(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected LogEntry
	}{
		{
			name: "postfix smtpd connection",
			line: "May 19 10:15:32 mail postfix/smtpd[1234]: connect from unknown[192.168.1.1]",
			expected: LogEntry{
				Timestamp: time.Date(time.Now().Year(), time.May, 19, 10, 15, 32, 0, time.UTC),
				Service:   "postfix",
				Component: "smtpd",
				Level:     "info",
				Message:   "connect from unknown[192.168.1.1]",
				PID:       1234,
			},
		},
		{
			name: "postfix smtp delivery with status",
			line: "May 19 10:15:33 mail postfix/smtp[5678]: ABC123: to=<user@example.com>, relay=mx.example.com[93.184.216.34]:25, delay=0.5, status=sent (250 OK)",
			expected: LogEntry{
				Timestamp: time.Date(time.Now().Year(), time.May, 19, 10, 15, 33, 0, time.UTC),
				Service:   "postfix",
				Component: "smtp",
				Level:     "info",
				Message:   "ABC123: to=<user@example.com>, relay=mx.example.com[93.184.216.34]:25, delay=0.5, status=sent (250 OK)",
				PID:       5678,
				QueueID:   "ABC123",
				To:        "user@example.com",
				Status:    "sent",
			},
		},
		{
			name: "postfix cleanup with from",
			line: "May 19 10:15:31 mail postfix/cleanup[9012]: DEF456: message-id=<msg001@example.com>",
			expected: LogEntry{
				Timestamp: time.Date(time.Now().Year(), time.May, 19, 10, 15, 31, 0, time.UTC),
				Service:   "postfix",
				Component: "cleanup",
				Level:     "info",
				Message:   "DEF456: message-id=<msg001@example.com>",
				PID:       9012,
				QueueID:   "DEF456",
			},
		},
		{
			name: "postfix qmgr with from",
			line: "May 19 10:15:31 mail postfix/qmgr[3456]: GHI789: from=<sender@local.test>, size=1024, nrcpt=1 (queue active)",
			expected: LogEntry{
				Timestamp: time.Date(time.Now().Year(), time.May, 19, 10, 15, 31, 0, time.UTC),
				Service:   "postfix",
				Component: "qmgr",
				Level:     "info",
				Message:   "GHI789: from=<sender@local.test>, size=1024, nrcpt=1 (queue active)",
				PID:       3456,
				QueueID:   "GHI789",
				From:      "sender@local.test",
			},
		},
		{
			name: "postfix bounce status",
			line: "May 19 10:16:00 mail postfix/smtp[7890]: JKL012: to=<bad@nowhere.invalid>, status=bounced (host not found)",
			expected: LogEntry{
				Timestamp: time.Date(time.Now().Year(), time.May, 19, 10, 16, 0, 0, time.UTC),
				Service:   "postfix",
				Component: "smtp",
				Level:     "error",
				Message:   "JKL012: to=<bad@nowhere.invalid>, status=bounced (host not found)",
				PID:       7890,
				QueueID:   "JKL012",
				To:        "bad@nowhere.invalid",
				Status:    "bounced",
			},
		},
	}

	parser := NewParser()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entry, err := parser.Parse(tc.line)
			require.NoError(t, err)

			assert.Equal(t, tc.expected.Timestamp, entry.Timestamp)
			assert.Equal(t, tc.expected.Service, entry.Service)
			assert.Equal(t, tc.expected.Component, entry.Component)
			assert.Equal(t, tc.expected.Level, entry.Level)
			assert.Equal(t, tc.expected.Message, entry.Message)
			assert.Equal(t, tc.expected.PID, entry.PID)
			assert.Equal(t, tc.expected.QueueID, entry.QueueID)
			assert.Equal(t, tc.expected.From, entry.From)
			assert.Equal(t, tc.expected.To, entry.To)
			assert.Equal(t, tc.expected.Status, entry.Status)
		})
	}
}

func TestParseDovecotLog(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected LogEntry
	}{
		{
			name: "dovecot imap login",
			line: "May 19 10:20:00 mail dovecot: imap-login: Login: user=<admin@local.test>, method=PLAIN, rip=172.18.0.1",
			expected: LogEntry{
				Timestamp: time.Date(time.Now().Year(), time.May, 19, 10, 20, 0, 0, time.UTC),
				Service:   "dovecot",
				Component: "imap-login",
				Level:     "info",
				Message:   "Login: user=<admin@local.test>, method=PLAIN, rip=172.18.0.1",
			},
		},
		{
			name: "dovecot lmtp delivery",
			line: "May 19 10:20:01 mail dovecot: lmtp(admin@local.test)<1234>: saved mail to INBOX",
			expected: LogEntry{
				Timestamp: time.Date(time.Now().Year(), time.May, 19, 10, 20, 1, 0, time.UTC),
				Service:   "dovecot",
				Component: "lmtp",
				Level:     "info",
				Message:   "saved mail to INBOX",
			},
		},
		{
			name: "dovecot auth failure",
			line: "May 19 10:20:02 mail dovecot: auth: Error: auth failed, user=<hacker@evil.com>, method=PLAIN",
			expected: LogEntry{
				Timestamp: time.Date(time.Now().Year(), time.May, 19, 10, 20, 2, 0, time.UTC),
				Service:   "dovecot",
				Component: "auth",
				Level:     "error",
				Message:   "auth failed, user=<hacker@evil.com>, method=PLAIN",
			},
		},
		{
			name: "dovecot imap disconnect",
			line: "May 19 10:20:03 mail dovecot: imap(admin@local.test)<5678>: Disconnected: Logged out bytes=123/456",
			expected: LogEntry{
				Timestamp: time.Date(time.Now().Year(), time.May, 19, 10, 20, 3, 0, time.UTC),
				Service:   "dovecot",
				Component: "imap",
				Level:     "info",
				Message:   "Disconnected: Logged out bytes=123/456",
			},
		},
	}

	parser := NewParser()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entry, err := parser.Parse(tc.line)
			require.NoError(t, err)

			assert.Equal(t, tc.expected.Timestamp, entry.Timestamp)
			assert.Equal(t, tc.expected.Service, entry.Service)
			assert.Equal(t, tc.expected.Component, entry.Component)
			assert.Equal(t, tc.expected.Level, entry.Level)
			assert.Equal(t, tc.expected.Message, entry.Message)
		})
	}
}

func TestParseInvalidLines(t *testing.T) {
	parser := NewParser()

	invalidLines := []string{
		"",
		"not a valid log line at all",
		"May 19",
	}

	for _, line := range invalidLines {
		t.Run(line, func(t *testing.T) {
			_, err := parser.Parse(line)
			assert.Error(t, err)
		})
	}
}

func TestParseLevelDetection(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name          string
		line          string
		expectedLevel string
	}{
		{
			name:          "warning in message",
			line:          "May 19 10:00:00 mail postfix/smtpd[100]: warning: hostname verification failed",
			expectedLevel: "warning",
		},
		{
			name:          "fatal in message",
			line:          "May 19 10:00:00 mail postfix/master[100]: fatal: no inet_interfaces configured",
			expectedLevel: "error",
		},
		{
			name:          "error prefix in dovecot",
			line:          "May 19 10:00:00 mail dovecot: auth: Error: password mismatch",
			expectedLevel: "error",
		},
		{
			name:          "normal info message",
			line:          "May 19 10:00:00 mail postfix/smtpd[100]: connect from localhost[127.0.0.1]",
			expectedLevel: "info",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entry, err := parser.Parse(tc.line)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedLevel, entry.Level)
		})
	}
}
