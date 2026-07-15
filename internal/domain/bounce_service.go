package domain

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

var ErrBounceNotFound = errors.New("bounce not found")

type BounceService struct {
	conn *sql.DB
}

func NewBounceService(conn *sql.DB) *BounceService {
	return &BounceService{conn: conn}
}

func (s *BounceService) RecordBounce(ctx context.Context, recipient, sender, subject, bounceType, errorMessage string) (*Bounce, error) {
	if recipient == "" {
		return nil, fmt.Errorf("recipient is required")
	}
	if bounceType != "hard" && bounceType != "soft" {
		return nil, fmt.Errorf("bounceType must be hard or soft")
	}

	now := time.Now().UTC()
	result, err := s.conn.ExecContext(ctx,
		`INSERT INTO bounces (recipient, sender, subject, bounce_type, error_message, bounced_at) VALUES (?, ?, ?, ?, ?, ?)`,
		recipient, sender, subject, bounceType, errorMessage, now,
	)
	if err != nil {
		return nil, fmt.Errorf("insert bounce: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	return &Bounce{
		ID:           id,
		Recipient:    recipient,
		Sender:       sender,
		Subject:      subject,
		BounceType:   bounceType,
		ErrorMessage: errorMessage,
		BouncedAt:    now,
	}, nil
}

func (s *BounceService) GetBounce(ctx context.Context, id int64) (*Bounce, error) {
	var b Bounce
	err := s.conn.QueryRowContext(ctx,
		`SELECT id, recipient, sender, subject, bounce_type, error_message, bounced_at FROM bounces WHERE id = ?`,
		id,
	).Scan(&b.ID, &b.Recipient, &b.Sender, &b.Subject, &b.BounceType, &b.ErrorMessage, &b.BouncedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrBounceNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get bounce: %w", err)
	}
	return &b, nil
}

func (s *BounceService) DeleteBounce(ctx context.Context, id int64) error {
	result, err := s.conn.ExecContext(ctx, `DELETE FROM bounces WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete bounce: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if rows == 0 {
		return ErrBounceNotFound
	}
	return nil
}

func (s *BounceService) ListBounces(ctx context.Context, filter BounceFilter) ([]Bounce, error) {
	query := `SELECT id, recipient, sender, subject, bounce_type, error_message, bounced_at FROM bounces WHERE 1=1`
	args := []interface{}{}

	if filter.Recipient != "" {
		query += ` AND recipient = ?`
		args = append(args, filter.Recipient)
	}
	if filter.BounceType != "" {
		query += ` AND bounce_type = ?`
		args = append(args, filter.BounceType)
	}

	query += ` ORDER BY bounced_at DESC`

	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += ` OFFSET ?`
		args = append(args, filter.Offset)
	}

	rows, err := s.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query bounces: %w", err)
	}
	defer rows.Close()

	bounces := []Bounce{}
	for rows.Next() {
		var b Bounce
		if err := rows.Scan(&b.ID, &b.Recipient, &b.Sender, &b.Subject, &b.BounceType, &b.ErrorMessage, &b.BouncedAt); err != nil {
			return nil, fmt.Errorf("scan bounce: %w", err)
		}
		bounces = append(bounces, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bounces: %w", err)
	}

	return bounces, nil
}
