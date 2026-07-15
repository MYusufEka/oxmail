package mail

import "strings"

// ParseBounceType returns "hard" for 5xx SMTP errors, "soft" for 4xx or unknown.
func ParseBounceType(smtpErr string) string {
	trimmed := strings.TrimSpace(smtpErr)
	if len(trimmed) >= 1 && trimmed[0] == '5' {
		return "hard"
	}
	return "soft"
}
