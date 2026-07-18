package mail

import (
	"testing"
)

func TestParseBounceType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"hard 5xx permanent failure", "550 5.1.1 User unknown", "hard"},
		{"hard 5xx no detail", "5.0.0", "hard"},
		{"hard 5xx numeric only", "5", "hard"},
		{"hard 5xx with spaces", "  550 mailbox unavailable", "hard"},
		{"soft 4xx temporary", "450 4.1.0 Mailbox busy", "soft"},
		{"soft 4xx retry", "451 4.3.0 Temporary lookup failure", "soft"},
		{"soft 2xx success", "250 2.1.5 OK", "soft"},
		{"soft 3xx", "354 Start mail input", "soft"},
		{"soft empty string", "", "soft"},
		{"soft random text", "Connection timed out", "soft"},
		{"soft leading whitespace then 4xx", "  450 mailbox full", "soft"},
		{"hard 5xx with leading/trailing", "\t550 5.1.1 <user@domain>: User unknown\r\n", "hard"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseBounceType(tt.input)
			if got != tt.expected {
				t.Errorf("ParseBounceType(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
