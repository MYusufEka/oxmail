package domain

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Signature stores one user's webmail signature settings.
type Signature struct {
	Email   string `json:"email"`
	Content string `json:"content"`
	Enabled bool   `json:"enabled"`
}

// SignatureService handles signature persistence.
type SignatureService struct {
	conn *sql.DB
}

// NewSignatureService creates a SignatureService with the given DB connection.
func NewSignatureService(conn *sql.DB) *SignatureService {
	return &SignatureService{conn: conn}
}

// Get returns saved signature settings, or an empty disabled signature when absent.
func (s *SignatureService) Get(ctx context.Context, email string) (*Signature, error) {
	var signature Signature
	var enabled int
	err := s.conn.QueryRowContext(ctx,
		`SELECT user_email, content, enabled FROM signatures WHERE user_email = ?`, email,
	).Scan(&signature.Email, &signature.Content, &enabled)
	if errors.Is(err, sql.ErrNoRows) {
		return emptySignature(email), nil
	}
	if err != nil {
		return nil, fmt.Errorf("query signature: %w", err)
	}

	signature.Enabled = enabled != 0
	return &signature, nil
}

// Upsert saves signature settings for an email and returns the persisted signature.
func (s *SignatureService) Upsert(ctx context.Context, email, content string, enabled bool) (*Signature, error) {
	enabledValue := 0
	if enabled {
		enabledValue = 1
	}

	_, err := s.conn.ExecContext(ctx,
		`INSERT INTO signatures (user_email, content, enabled, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(user_email) DO UPDATE SET
			content = excluded.content,
			enabled = excluded.enabled,
			updated_at = CURRENT_TIMESTAMP`,
		email, content, enabledValue,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert signature: %w", err)
	}

	return s.Get(ctx, email)
}

// Delete removes signature settings for an email. Missing signatures are ignored.
func (s *SignatureService) Delete(ctx context.Context, email string) error {
	_, err := s.conn.ExecContext(ctx, `DELETE FROM signatures WHERE user_email = ?`, email)
	if err != nil {
		return fmt.Errorf("delete signature: %w", err)
	}
	return nil
}

func emptySignature(email string) *Signature {
	return &Signature{
		Email:   email,
		Content: "",
		Enabled: false,
	}
}
