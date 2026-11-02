package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// openRetryWindow bounds retries for the creation race: several processes
// starting together can collide while the driver applies journal_mode. It is
// deliberately short; a genuinely stuck lock still fails with a clear error.
const openRetryWindow = 5 * time.Second

// DB is the SQLite-backed store. It implements the application's Store port
// and owns connection setup and migrations.
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

	deadline := time.Now().Add(openRetryWindow)
	delay := 10 * time.Millisecond
	for {
		db, err := openOnce(ctx, path)
		if err == nil {
			return db, nil
		}
		if !isBusyError(err) || time.Now().After(deadline) {
			return nil, describeDBError(path, err)
		}

		select {
		case <-ctx.Done():
			return nil, describeDBError(path, ctx.Err())
		case <-time.After(delay):
		}
		if delay < 100*time.Millisecond {
			delay *= 2
		}
	}
}

func openOnce(ctx context.Context, path string) (*DB, error) {
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

	if err := migrate(ctx, sqlDB, migrationsFS); err != nil {
		sqlDB.Close()
		return nil, err
	}
	return &DB{db: sqlDB}, nil
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

func isBusyError(err error) bool {
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "database is locked") || strings.Contains(text, "sqlite_busy")
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
