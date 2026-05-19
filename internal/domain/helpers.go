package domain

import "strings"

// validateEmail checks basic email format: non-empty local@domain with no whitespace.
func validateEmail(email string) error {
	if email == "" {
		return ErrInvalidEmail
	}
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return ErrInvalidEmail
	}
	if strings.ContainsAny(email, " \t\n\r") {
		return ErrInvalidEmail
	}
	return nil
}

// extractDomain returns the domain part of an email address.
func extractDomain(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}
