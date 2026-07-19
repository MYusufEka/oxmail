package database

import (
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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

func TestForeignKeysEnabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	dbPath := filepath.Join(t.TempDir(), "foreign-keys.sqlite")
	db, err := Open(dbPath, logger)
	if err != nil {
		t.Fatalf("Open(temp db) = _, %v, want nil", err)
	}
	defer db.Close()

	for connectionIndex := 0; connectionIndex < 3; connectionIndex++ {
		conn, err := db.Conn.Conn(t.Context())
		if err != nil {
			t.Fatalf("Conn(%d): %v", connectionIndex, err)
		}

		var enabled int
		err = conn.QueryRowContext(t.Context(), "PRAGMA foreign_keys").Scan(&enabled)
		closeErr := conn.Close()
		if err != nil {
			t.Fatalf("PRAGMA foreign_keys on connection %d: %v", connectionIndex, err)
		}
		if closeErr != nil {
			t.Fatalf("close connection %d: %v", connectionIndex, closeErr)
		}
		if enabled != 1 {
			t.Fatalf("PRAGMA foreign_keys on connection %d = %d, want 1", connectionIndex, enabled)
		}
	}
}

func TestSchemaMigrationsTracked(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	db, err := Open(":memory:", logger)
	if err != nil {
		t.Fatalf("Open(':memory:') = _, %v, want nil", err)
	}
	defer db.Close()

	assertSchemaMigrationsTracked(t, db.Conn)
}

func TestOpen_SkipsAlreadyTrackedMigrations(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	dbPath := filepath.Join(t.TempDir(), "repeat-migrations.sqlite")

	db, err := Open(dbPath, logger)
	if err != nil {
		t.Fatalf("first Open(temp db) = _, %v, want nil", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close first db: %v", err)
	}

	db, err = Open(dbPath, logger)
	if err != nil {
		t.Fatalf("second Open(temp db) = _, %v, want nil", err)
	}
	defer db.Close()

	assertSchemaMigrationsTracked(t, db.Conn)
}

func TestSchemaMigrationsTrackedRepairsIntegerVersionTable(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	dbPath := filepath.Join(t.TempDir(), "schema-migrations.sqlite")

	registerSQLiteDriver()
	legacyConn, err := sql.Open(sqliteDriverName, dbPath)
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	if _, err := legacyConn.Exec(`CREATE TABLE schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		legacyConn.Close()
		t.Fatalf("create legacy schema_migrations: %v", err)
	}
	if err := legacyConn.Close(); err != nil {
		t.Fatalf("close legacy db: %v", err)
	}

	db, err := Open(dbPath, logger)
	if err != nil {
		t.Fatalf("Open(legacy db) = _, %v, want nil", err)
	}
	defer db.Close()

	assertSchemaMigrationsTracked(t, db.Conn)
}

func assertSchemaMigrationsTracked(t *testing.T, conn *sql.DB) {
	t.Helper()

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		t.Fatalf("read migrations dir: %v", err)
	}

	expectedVersions := make(map[string]bool)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		expectedVersions[strings.TrimSuffix(entry.Name(), ".sql")] = true
	}

	rows, err := conn.Query("SELECT version, typeof(version) FROM schema_migrations ORDER BY version")
	if err != nil {
		t.Fatalf("query schema_migrations: %v", err)
	}
	defer rows.Close()

	actualVersions := make(map[string]bool)
	for rows.Next() {
		var version string
		var storageType string
		if err := rows.Scan(&version, &storageType); err != nil {
			t.Fatalf("scan schema_migrations version: %v", err)
		}
		if storageType != "text" {
			t.Fatalf("schema_migrations version %q storage type = %q, want text", version, storageType)
		}
		actualVersions[version] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate schema_migrations: %v", err)
	}

	if len(actualVersions) != len(expectedVersions) {
		t.Fatalf("schema_migrations rows = %d, want %d; got %#v", len(actualVersions), len(expectedVersions), actualVersions)
	}
	for version := range expectedVersions {
		if !actualVersions[version] {
			t.Errorf("schema_migrations missing version %q", version)
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
