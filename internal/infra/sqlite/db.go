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
		return nil, fmt.Errorf("open database: %w", err)
	}

	db := &DB{db: sqlDB}
	if err := migrate(ctx, sqlDB, migrationsFS); err != nil {
		sqlDB.Close()
		return nil, err
	}
	return db, nil
}

// dsn applies pragmas on every connection the driver opens, rather than only
// on whichever connection happens to run first.
func dsn(path string) string {
	pragmas := []string{
		"_pragma=busy_timeout(5000)",
		"_pragma=journal_mode(WAL)",
		"_pragma=synchronous(NORMAL)",
		"_pragma=foreign_keys(ON)",
	}
	return "file:" + filepath.ToSlash(path) + "?" + strings.Join(pragmas, "&")
}

func (d *DB) Close() error { return d.db.Close() }
