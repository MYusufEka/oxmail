package domain

import (
	"context"
	"database/sql"
	"fmt"
)

type StatsService struct {
	conn *sql.DB
}

func NewStatsService(conn *sql.DB) *StatsService {
	return &StatsService{conn: conn}
}

func (s *StatsService) IncrementSent(ctx context.Context) error {
	_, err := s.conn.ExecContext(ctx,
		`INSERT INTO mail_stats (stat_date, sent) VALUES (date('now'), 1)
		 ON CONFLICT (stat_date) DO UPDATE SET sent = sent + 1`,
	)
	if err != nil {
		return fmt.Errorf("increment sent: %w", err)
	}
	return nil
}

func (s *StatsService) IncrementReceived(ctx context.Context) error {
	_, err := s.conn.ExecContext(ctx,
		`INSERT INTO mail_stats (stat_date, received) VALUES (date('now'), 1)
		 ON CONFLICT (stat_date) DO UPDATE SET received = received + 1`,
	)
	if err != nil {
		return fmt.Errorf("increment received: %w", err)
	}
	return nil
}

func (s *StatsService) IncrementBounced(ctx context.Context) error {
	_, err := s.conn.ExecContext(ctx,
		`INSERT INTO mail_stats (stat_date, bounced) VALUES (date('now'), 1)
		 ON CONFLICT (stat_date) DO UPDATE SET bounced = bounced + 1`,
	)
	if err != nil {
		return fmt.Errorf("increment bounced: %w", err)
	}
	return nil
}

func (s *StatsService) IncrementSpamCaught(ctx context.Context) error {
	_, err := s.conn.ExecContext(ctx,
		`INSERT INTO mail_stats (stat_date, spam_caught) VALUES (date('now'), 1)
		 ON CONFLICT (stat_date) DO UPDATE SET spam_caught = spam_caught + 1`,
	)
	if err != nil {
		return fmt.Errorf("increment spam_caught: %w", err)
	}
	return nil
}

func (s *StatsService) GetSummary(ctx context.Context) (*StatSummary, error) {
	var summary StatSummary
	err := s.conn.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(sent), 0), COALESCE(SUM(received), 0),
		        COALESCE(SUM(bounced), 0), COALESCE(SUM(spam_caught), 0)
		 FROM mail_stats`,
	).Scan(&summary.TotalSent, &summary.TotalReceived, &summary.TotalBounced, &summary.TotalSpamCaught)
	if err != nil {
		return nil, fmt.Errorf("query summary: %w", err)
	}
	return &summary, nil
}

func (s *StatsService) GetStats(ctx context.Context, days int) ([]DailyStat, error) {
	rows, err := s.conn.QueryContext(ctx,
		`SELECT stat_date, sent, received, bounced, spam_caught
		 FROM mail_stats
		 WHERE stat_date >= date('now', ? || ' days')
		 ORDER BY stat_date DESC`,
		fmt.Sprintf("-%d", days),
	)
	if err != nil {
		return nil, fmt.Errorf("query stats: %w", err)
	}
	defer rows.Close()

	stats := make([]DailyStat, 0)
	for rows.Next() {
		var stat DailyStat
		if err := rows.Scan(&stat.Date, &stat.Sent, &stat.Received, &stat.Bounced, &stat.SpamCaught); err != nil {
			return nil, fmt.Errorf("scan stat: %w", err)
		}
		stats = append(stats, stat)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return stats, nil
}
