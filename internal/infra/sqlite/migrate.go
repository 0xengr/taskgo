package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type migration struct {
	version int
	name    string
	sql     string
}

func loadMigrations(fsys fs.FS) ([]migration, error) {
	entries, err := fs.ReadDir(fsys, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}

	var migrations []migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		prefix, _, ok := strings.Cut(entry.Name(), "_")
		if !ok {
			return nil, fmt.Errorf("migration %q has no version prefix", entry.Name())
		}
		version, err := strconv.Atoi(prefix)
		if err != nil {
			return nil, fmt.Errorf("migration %q has an invalid version prefix", entry.Name())
		}
		body, err := fs.ReadFile(fsys, "migrations/"+entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read migration %q: %w", entry.Name(), err)
		}
		migrations = append(migrations, migration{version: version, name: entry.Name(), sql: string(body)})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})
	for i := 1; i < len(migrations); i++ {
		if migrations[i].version == migrations[i-1].version {
			return nil, fmt.Errorf("duplicate migration version %d", migrations[i].version)
		}
	}
	return migrations, nil
}

// migrate applies pending migrations. The common path (already current) only
// reads user_version. An upgrade takes an immediate transaction and re-reads
// the version under that lock, so two processes starting together cannot both
// apply the same migration.
func migrate(ctx context.Context, db *sql.DB, fsys fs.FS) error {
	migrations, err := loadMigrations(fsys)
	if err != nil {
		return err
	}
	if len(migrations) == 0 {
		return errors.New("no migrations embedded in this build")
	}
	latest := migrations[len(migrations)-1].version

	current, err := readSchemaVersion(ctx, db)
	if err != nil {
		return err
	}
	if current > latest {
		return errSchemaNewer(current, latest)
	}
	if current == latest {
		return nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("start migration transaction: %w", err)
	}
	defer tx.Rollback()

	current, err = readSchemaVersion(ctx, tx)
	if err != nil {
		return err
	}
	if current > latest {
		return errSchemaNewer(current, latest)
	}

	for _, m := range migrations {
		if m.version <= current {
			continue
		}
		if _, err := tx.ExecContext(ctx, m.sql); err != nil {
			return fmt.Errorf("migration %04d (%s): %w", m.version, m.name, err)
		}
		if _, err := tx.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", m.version)); err != nil {
			return fmt.Errorf("record schema version %d: %w", m.version, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migrations: %w", err)
	}
	return nil
}

type versionReader interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func readSchemaVersion(ctx context.Context, reader versionReader) (int, error) {
	var version int
	if err := reader.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return 0, fmt.Errorf("read schema version: %w", err)
	}
	return version, nil
}

func errSchemaNewer(current, supported int) error {
	return fmt.Errorf("database schema v%d is newer than this taskgo supports (v%d)", current, supported)
}
