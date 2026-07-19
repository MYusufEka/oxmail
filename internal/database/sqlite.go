package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

const sqliteDriverName = "oxmail_sqlite"

var registerSQLiteDriverOnce sync.Once

// DB wraps the sql.DB connection.
type DB struct {
	Conn   *sql.DB
	logger *slog.Logger
}

func registerSQLiteDriver() {
	registerSQLiteDriverOnce.Do(func() {
		driver := &sqlite.Driver{}
		driver.RegisterConnectionHook(func(conn sqlite.ExecQuerierContext, dsn string) error {
			_, err := conn.ExecContext(context.Background(), "PRAGMA foreign_keys=ON", nil)
			return err
		})
		sql.Register(sqliteDriverName, driver)
	})
}

// Open creates a new SQLite connection and runs migrations.
func Open(dsn string, logger *slog.Logger) (*DB, error) {
	registerSQLiteDriver()

	conn, err := sql.Open(sqliteDriverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	db := &DB{
		Conn:   conn,
		logger: logger,
	}

	if err := db.runMigrations(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return db, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.Conn.Close()
}

// runMigrations applies all embedded SQL migrations in order.
func (db *DB) runMigrations() error {
	if err := db.ensureSchemaMigrationsTable(); err != nil {
		return err
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		version := strings.TrimSuffix(entry.Name(), ".sql")
		applied, err := db.isMigrationApplied(version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		content, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		if _, err := db.Conn.Exec(string(content)); err != nil {
			return fmt.Errorf("execute migration %s: %w", entry.Name(), err)
		}

		if _, err := db.Conn.Exec("INSERT OR IGNORE INTO schema_migrations(version) VALUES(?)", version); err != nil {
			return fmt.Errorf("record migration %s: %w", entry.Name(), err)
		}

		db.logMigrationApplied(entry.Name())
	}

	return nil
}

func (db *DB) isMigrationApplied(version string) (bool, error) {
	var appliedVersion string
	err := db.Conn.QueryRow("SELECT version FROM schema_migrations WHERE version = ?", version).Scan(&appliedVersion)
	if err == nil {
		return true, nil
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return false, fmt.Errorf("check migration %s: %w", version, err)
}

func (db *DB) ensureSchemaMigrationsTable() error {
	compatible, err := db.hasCompatibleSchemaMigrationsTable()
	if err != nil {
		return err
	}
	if !compatible {
		if _, err := db.Conn.Exec("DROP TABLE IF EXISTS schema_migrations"); err != nil {
			return fmt.Errorf("drop incompatible schema_migrations table: %w", err)
		}
	}

	_, err = db.Conn.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
)`)
	if err != nil {
		return fmt.Errorf("ensure schema_migrations table: %w", err)
	}
	return nil
}

func (db *DB) hasCompatibleSchemaMigrationsTable() (bool, error) {
	rows, err := db.Conn.Query("PRAGMA table_info(schema_migrations)")
	if err != nil {
		return false, fmt.Errorf("inspect schema_migrations table: %w", err)
	}
	defer rows.Close()

	columns := make(map[string]schemaMigrationColumn)
	for rows.Next() {
		var column schemaMigrationColumn
		var cid int
		if err := rows.Scan(&cid, &column.name, &column.columnType, &column.notNull, &column.defaultValue, &column.primaryKey); err != nil {
			return false, fmt.Errorf("scan schema_migrations column: %w", err)
		}
		columns[column.name] = column
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterate schema_migrations columns: %w", err)
	}
	if len(columns) == 0 {
		return false, nil
	}

	versionColumn, hasVersion := columns["version"]
	appliedAtColumn, hasAppliedAt := columns["applied_at"]
	return hasVersion && hasAppliedAt &&
		strings.EqualFold(versionColumn.columnType, "TEXT") &&
		versionColumn.primaryKey > 0 &&
		appliedAtColumn.notNull == 1, nil
}

type schemaMigrationColumn struct {
	name         string
	columnType   string
	notNull      int
	defaultValue sql.NullString
	primaryKey   int
}

func (db *DB) logMigrationApplied(filename string) {
	if db.logger == nil {
		return
	}
	db.logger.Info("migration applied", "file", filename)
}
