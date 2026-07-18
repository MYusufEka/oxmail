package database

import (
	"log/slog"
	"os"
	"testing"
)

func TestOpen_InMemory(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := Open(":memory:", logger)
	if err != nil {
		t.Fatalf("Open(':memory:') = _, %v, want nil", err)
	}
	defer db.Close()

	if db.Conn == nil {
		t.Fatal("db.Conn is nil after Open")
	}

	// Verify migrations ran by checking for expected tables
	rows, err := db.Conn.Query("SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
	if err != nil {
		t.Fatalf("query tables: %v", err)
	}
	defer rows.Close()

	tables := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan table name: %v", err)
		}
		tables[name] = true
	}

	expectedTables := []string{"domains", "users", "aliases", "dkim_keys", "contacts", "bounces", "mail_stats", "audit_log", "schema_migrations"}
	for _, table := range expectedTables {
		if !tables[table] {
			t.Errorf("expected table %q not found after migrations", table)
		}
	}
}

func TestOpen_InvalidPath(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	// Try to create DB in a non-existent directory
	_, err := Open("/nonexistent/path/db.sqlite", logger)
	if err == nil {
		t.Fatal("Open('/nonexistent/...') = nil, want error")
	}
}

func TestOpen_DoubleClose(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := Open(":memory:", logger)
	if err != nil {
		t.Fatalf("Open(':memory:') = _, %v, want nil", err)
	}

	// First close should succeed
	err = db.Close()
	if err != nil {
		t.Fatalf("first Close() = %v, want nil", err)
	}

	// Second close should also not panic (likely errors but shouldn't crash)
	_ = db.Close()
}

func TestDB_Close(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := Open(":memory:", logger)
	if err != nil {
		t.Fatalf("Open(':memory:') = _, %v, want nil", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
}
