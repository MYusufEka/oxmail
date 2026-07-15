package domain

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"
)

type AuditService struct {
	conn *sql.DB
}

func NewAuditService(conn *sql.DB) *AuditService {
	return &AuditService{conn: conn}
}

func resolveActor(ctx context.Context) string {
	if os.Getenv("OXMAIL_MODE") == "dev" {
		return "admin"
	}
	if actor, ok := ctx.Value(contextKeyActor).(string); ok && actor != "" {
		return actor
	}
	return "admin"
}

type contextKey string

const contextKeyActor contextKey = "actor"

func (s *AuditService) Log(ctx context.Context, actor, action, targetType, targetID, detail string) error {
	if actor == "" {
		actor = resolveActor(ctx)
	}
	if detail == "" {
		detail = "{}"
	}
	_, err := s.conn.ExecContext(ctx,
		`INSERT INTO audit_log (actor, action, target_type, target_id, detail, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		actor, action, targetType, targetID, detail, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("audit log insert: %w", err)
	}
	return nil
}

func (s *AuditService) List(ctx context.Context, limit, offset int) ([]AuditEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.conn.QueryContext(ctx,
		`SELECT id, actor, action, target_type, target_id, detail, created_at
		 FROM audit_log ORDER BY created_at DESC LIMIT ? OFFSET ?`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("audit log query: %w", err)
	}
	defer rows.Close()

	var entries []AuditEntry
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.ID, &e.Actor, &e.Action, &e.TargetType, &e.TargetID, &e.Detail, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("audit log scan: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("audit log rows: %w", err)
	}
	if entries == nil {
		entries = []AuditEntry{}
	}
	return entries, nil
}

func (s *AuditService) Count(ctx context.Context) (int, error) {
	var total int
	err := s.conn.QueryRowContext(ctx, `SELECT COUNT(*) FROM audit_log`).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("audit log count: %w", err)
	}
	return total, nil
}

func (s *AuditService) ListFiltered(ctx context.Context, actor, action string, limit, offset int) ([]AuditEntry, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `SELECT id, actor, action, target_type, target_id, detail, created_at FROM audit_log WHERE 1=1`
	args := []any{}

	if actor != "" {
		query += ` AND actor = ?`
		args = append(args, actor)
	}
	if action != "" {
		query += ` AND action = ?`
		args = append(args, action)
	}

	query += ` ORDER BY created_at DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := s.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("audit log filtered query: %w", err)
	}
	defer rows.Close()

	var entries []AuditEntry
	for rows.Next() {
		var e AuditEntry
		if err := rows.Scan(&e.ID, &e.Actor, &e.Action, &e.TargetType, &e.TargetID, &e.Detail, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("audit log scan: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("audit log rows: %w", err)
	}
	if entries == nil {
		entries = []AuditEntry{}
	}
	return entries, nil
}

func (s *AuditService) CountFiltered(ctx context.Context, actor, action string) (int, error) {
	query := `SELECT COUNT(*) FROM audit_log WHERE 1=1`
	args := []any{}

	if actor != "" {
		query += ` AND actor = ?`
		args = append(args, actor)
	}
	if action != "" {
		query += ` AND action = ?`
		args = append(args, action)
	}

	var total int
	err := s.conn.QueryRowContext(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("audit log count filtered: %w", err)
	}
	return total, nil
}
