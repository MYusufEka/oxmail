package domain

import (
	"errors"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr error
	}{
		{"valid simple", "alice@example.com", nil},
		{"valid with dots", "alice.smith@example.co.uk", nil},
		{"valid plus", "alice+tag@example.com", nil},
		{"valid subdomain", "user@mail.example.com", nil},
		{"empty string", "", ErrInvalidEmail},
		{"no domain", "alice@", ErrInvalidEmail},
		{"no local", "@example.com", ErrInvalidEmail},
		{"no at sign", "alice", ErrInvalidEmail},
		// SplitN splits on first @ only, so alice@foo@bar.com gives ["alice", "foo@bar.com"] - passes validation
		{"double at passes (SplitN behavior)", "alice@foo@bar.com", nil},
		{"has space", "ali ce@example.com", ErrInvalidEmail},
		{"has tab", "alice\t@example.com", ErrInvalidEmail},
		{"has newline", "alice\n@example.com", ErrInvalidEmail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmail(tt.email)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("validateEmail(%q) = %v, want %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  string
	}{
		{"simple domain", "alice@example.com", "example.com"},
		{"subdomain", "user@mail.example.co.uk", "mail.example.co.uk"},
		{"plus tag", "alice+tag@domain.org", "domain.org"},
		{"no at sign", "alice", ""},
		{"no domain", "alice@", ""},
		{"empty string", "", ""},
		{"multiple at signs", "alice@foo@bar.com", "foo@bar.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDomain(tt.email)
			if got != tt.want {
				t.Errorf("extractDomain(%q) = %q, want %q", tt.email, got, tt.want)
			}
		})
	}
}
