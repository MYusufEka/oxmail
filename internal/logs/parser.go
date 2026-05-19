package logs

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// LogEntry represents a parsed log line from a mail service.
type LogEntry struct {
	ID        int64     `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Service   string    `json:"service"`
	Component string    `json:"component"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	PID       int       `json:"pid,omitempty"`
	QueueID   string    `json:"queueId,omitempty"`
	From      string    `json:"from,omitempty"`
	To        string    `json:"to,omitempty"`
	Status    string    `json:"status,omitempty"`
}

// Parser parses syslog-formatted mail log lines into structured LogEntry values.
type Parser struct {
	postfixRe *regexp.Regexp
	dovecotRe *regexp.Regexp
	queueIDRe *regexp.Regexp
	fromRe    *regexp.Regexp
	toRe      *regexp.Regexp
	statusRe  *regexp.Regexp
}

// NewParser creates a Parser with precompiled regular expressions.
func NewParser() *Parser {
	return &Parser{
		postfixRe: regexp.MustCompile(`^(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+\S+\s+postfix/(\w+)\[(\d+)\]:\s+(.+)$`),
		dovecotRe: regexp.MustCompile(`^(\w{3}\s+\d{1,2}\s+\d{2}:\d{2}:\d{2})\s+\S+\s+dovecot:\s+(.+)$`),
		queueIDRe: regexp.MustCompile(`^([A-Z0-9]+):\s+`),
		fromRe:    regexp.MustCompile(`from=<([^>]*)>`),
		toRe:      regexp.MustCompile(`to=<([^>]*)>`),
		statusRe:  regexp.MustCompile(`status=(\w+)`),
	}
}

// Parse parses a single log line and returns a LogEntry.
func (p *Parser) Parse(line string) (LogEntry, error) {
	if len(line) == 0 {
		return LogEntry{}, fmt.Errorf("empty log line")
	}

	if entry, err := p.parsePostfix(line); err == nil {
		return entry, nil
	}

	if entry, err := p.parseDovecot(line); err == nil {
		return entry, nil
	}

	return LogEntry{}, fmt.Errorf("unrecognized log format: %s", line)
}

func (p *Parser) parsePostfix(line string) (LogEntry, error) {
	matches := p.postfixRe.FindStringSubmatch(line)
	if matches == nil {
		return LogEntry{}, fmt.Errorf("not a postfix log line")
	}

	timestamp, err := parseTimestamp(matches[1])
	if err != nil {
		return LogEntry{}, err
	}

	pid, _ := strconv.Atoi(matches[3])
	message := matches[4]

	entry := LogEntry{
		Timestamp: timestamp,
		Service:   "postfix",
		Component: matches[2],
		Level:     detectLevel(message),
		Message:   message,
		PID:       pid,
	}

	p.extractPostfixFields(&entry, message)

	return entry, nil
}

func (p *Parser) parseDovecot(line string) (LogEntry, error) {
	matches := p.dovecotRe.FindStringSubmatch(line)
	if matches == nil {
		return LogEntry{}, fmt.Errorf("not a dovecot log line")
	}

	timestamp, err := parseTimestamp(matches[1])
	if err != nil {
		return LogEntry{}, err
	}

	remainder := matches[2]
	component, message := parseDovecotComponent(remainder)
	level := detectLevel(remainder)

	return LogEntry{
		Timestamp: timestamp,
		Service:   "dovecot",
		Component: component,
		Level:     level,
		Message:   message,
	}, nil
}

func (p *Parser) extractPostfixFields(entry *LogEntry, message string) {
	if queueMatch := p.queueIDRe.FindStringSubmatch(message); queueMatch != nil {
		entry.QueueID = queueMatch[1]
	}

	if fromMatch := p.fromRe.FindStringSubmatch(message); fromMatch != nil {
		entry.From = fromMatch[1]
	}

	if toMatch := p.toRe.FindStringSubmatch(message); toMatch != nil {
		entry.To = toMatch[1]
	}

	if statusMatch := p.statusRe.FindStringSubmatch(message); statusMatch != nil {
		entry.Status = statusMatch[1]
		if entry.Status == "bounced" || entry.Status == "deferred" {
			entry.Level = "error"
		}
	}
}

// parseDovecotComponent extracts the component and message from dovecot log remainder.
// Formats:
//   - "imap-login: Login: user=..."         -> component="imap-login", message="Login: user=..."
//   - "lmtp(user@dom)<pid>: saved mail..."  -> component="lmtp", message="saved mail..."
//   - "auth: Error: password mismatch"      -> component="auth", message="password mismatch" (level=error)
//   - "imap(user@dom)<pid>: Disconnected"   -> component="imap", message="Disconnected..."
func parseDovecotComponent(remainder string) (string, string) {
	// Format: component(extra)<extra>: message
	parenIdx := strings.IndexByte(remainder, '(')
	colonIdx := strings.IndexByte(remainder, ':')

	if parenIdx > 0 && (colonIdx < 0 || parenIdx < colonIdx) {
		component := remainder[:parenIdx]
		// Find the colon after the paren/angle bracket section
		afterParen := strings.Index(remainder[parenIdx:], ": ")
		if afterParen >= 0 {
			message := remainder[parenIdx+afterParen+2:]
			return component, message
		}
	}

	// Format: component: message or component: Level: message
	if colonIdx > 0 {
		component := remainder[:colonIdx]
		rest := strings.TrimPrefix(remainder[colonIdx+1:], " ")

		// Check for sub-level like "Error: actual message"
		if strings.HasPrefix(rest, "Error: ") {
			return component, rest[7:]
		}
		if strings.HasPrefix(rest, "Warning: ") {
			return component, rest[9:]
		}

		return component, rest
	}

	return "unknown", remainder
}

func parseTimestamp(raw string) (time.Time, error) {
	currentYear := time.Now().Year()
	withYear := fmt.Sprintf("%s %d", raw, currentYear)
	return time.Parse("Jan  2 15:04:05 2006", withYear)
}

func detectLevel(message string) string {
	lower := strings.ToLower(message)

	if strings.Contains(lower, "fatal") || strings.Contains(lower, "panic") {
		return "error"
	}
	if strings.Contains(lower, "error:") || strings.HasPrefix(lower, "error ") {
		return "error"
	}
	if strings.Contains(lower, "warning:") || strings.HasPrefix(lower, "warning") {
		return "warning"
	}

	return "info"
}
