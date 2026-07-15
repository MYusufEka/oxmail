package mail

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// QueueStatus represents the Postfix mail queue state.
type QueueStatus struct {
	Total     int    `json:"total"`
	Deferred  int    `json:"deferred"`
	Active    int    `json:"active"`
	OldestAge string `json:"oldest_age"`
}

// GetQueueStatus runs mailq in the postfix container and returns parsed status.
func GetQueueStatus(executor CommandExecutor) (*QueueStatus, error) {
	output, err := executor.RunWithOutput("mailq")
	if err != nil {
		return nil, fmt.Errorf("mailq failed: %w", err)
	}
	return parseMailqOutput(output), nil
}

// parseMailqOutput parses raw output of mailq / postqueue -p.
func parseMailqOutput(output string) *QueueStatus {
	if strings.Contains(output, "Mail queue is empty") {
		return &QueueStatus{}
	}

	lines := strings.Split(output, "\n")

	type queueEntry struct {
		isActive   bool
		isDeferred bool
		arrTime   time.Time
	}

	var entries []queueEntry
	var cur *queueEntry

	// Matches: queue-id[*]  size  rest...
	entryRe := regexp.MustCompile(`^(\S+)\s+(\d+)\s+(.+)$`)

	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if line == "" || strings.HasPrefix(line, "-Queue ID-") || strings.HasPrefix(line, "--") {
			continue
		}

		if line[0] != ' ' && line[0] != '\t' {
			if cur != nil {
				entries = append(entries, *cur)
			}
			cur = &queueEntry{}

			matches := entryRe.FindStringSubmatch(line)
			if matches == nil {
				continue
			}

			qid := matches[1]
			cur.isActive = strings.HasSuffix(qid, "*")

			rest := strings.TrimSpace(matches[3])
			dateParts := spacedRe.Split(rest, 2)
			if len(dateParts) > 0 {
				cur.arrTime = tryParseMailqDate(strings.TrimSpace(dateParts[0]))
			}
		} else if cur != nil {
			if strings.Contains(line, "(deferred)") || strings.Contains(line, "(deferred,") {
				cur.isDeferred = true
			}
		}
	}
	if cur != nil {
		entries = append(entries, *cur)
	}

	status := &QueueStatus{
		Total: len(entries),
	}

	now := time.Now()
	var oldest time.Time
	for _, e := range entries {
		if e.isDeferred {
			status.Deferred++
		}
		if e.isActive {
			status.Active++
		}
		if !e.arrTime.IsZero() {
			if oldest.IsZero() || e.arrTime.Before(oldest) {
				oldest = e.arrTime
			}
		}
	}

	// If no * flags found (all entries listed as plain), derive active
	if status.Active == 0 && status.Total > 0 {
		status.Active = status.Total - status.Deferred
	}

	if !oldest.IsZero() {
		age := now.Sub(oldest)
		status.OldestAge = formatAge(age)
	}

	return status
}

var spacedRe = regexp.MustCompile(`\s{2,}`)

// tryParseMailqDate attempts to parse the date portion from mailq output.
// mailq omits the year, so we assume the current year and adjust if the
// result would be in the future (e.g. December message in January).
func tryParseMailqDate(s string) time.Time {
	layouts := []string{
		"Mon Jan 2 15:04:05",
		"Mon Jan 2",
		"Jan 2 15:04:05",
		"Jan 2",
	}
	now := time.Now()
	for _, layout := range layouts {
		t, err := time.Parse(layout, s)
		if err != nil {
			continue
		}
		t = t.AddDate(now.Year(), 0, 0)
		if t.After(now) {
			t = t.AddDate(-1, 0, 0)
		}
		return t
	}
	return time.Time{}
}

// formatAge formats a duration into a short human-readable string.
func formatAge(d time.Duration) string {
	switch {
	case d >= 24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d >= time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d >= time.Minute:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	default:
		return "<1m"
	}
}
