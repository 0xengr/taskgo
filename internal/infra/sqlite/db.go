package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// DB is the SQLite-backed store. It implements the application's TaskStore
// port and owns connection setup and migrations.
type DB struct {
	db *sql.DB
}

// Open creates the database file and its parent directory if needed, applies
// connection pragmas, and runs pending migrations.
func Open(ctx context.Context, path string) (*DB, error) {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	sqlDB, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	if err := sqlDB.PingContext(ctx); err != nil {
		sqlDB.Close()
		return nil, describeDBError(path, fmt.Errorf("open database: %w", err))
	}

	db := &DB{db: sqlDB}
	if err := migrate(ctx, sqlDB, migrationsFS); err != nil {
		sqlDB.Close()
		return nil, describeDBError(path, err)
	}
	return db, nil
}

// dsn applies pragmas on every connection the driver opens, rather than only
// on whichever connection happens to run first. _txlock=immediate makes
// migration transactions take the write lock up front.
func dsn(path string) string {
	pragmas := []string{
		"_pragma=busy_timeout(5000)",
		"_pragma=journal_mode(WAL)",
		"_pragma=synchronous(NORMAL)",
		"_pragma=foreign_keys(ON)",
		"_txlock=immediate",
	}
	return "file:" + filepath.ToSlash(path) + "?" + strings.Join(pragmas, "&")
}

// describeDBError adds the database path and a hint for the failures users
// actually hit: corruption and lock contention.
func describeDBError(path string, err error) error {
	text := strings.ToLower(err.Error())
	switch {
	case strings.Contains(text, "not a database"), strings.Contains(text, "malformed"):
		return fmt.Errorf("database %s is corrupt or unreadable; taskgo will not recreate or delete it: %w", path, err)
	case strings.Contains(text, "database is locked"), strings.Contains(text, "sqlite_busy"):
		return fmt.Errorf("database %s is busy; another taskgo process may be running: %w", path, err)
	default:
		return fmt.Errorf("database %s: %w", path, err)
	}
}

func (d *DB) Close() error { return d.db.Close() }
