package sqlite

import (
	"context"
	"database/sql"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/0xengr/taskgo/internal/domain"
)

func TestOpenCreatesSchema(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "nested", "taskgo.db")

	db, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer db.Close()

	var version int
	if err := db.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		t.Fatalf("read user_version: %v", err)
	}
	if version != 2 {
		t.Fatalf("user_version = %d, want 2", version)
	}

	var name string
	err = db.db.QueryRowContext(ctx,
		"SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'tasks'").Scan(&name)
	if err != nil {
		t.Fatalf("tasks table missing: %v", err)
	}
}

func TestOpenReusesExistingDatabase(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "taskgo.db")

	db, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("first Open returned error: %v", err)
	}
	if _, err := db.InsertTask(ctx, domain.NewTask{
		Title: "survives", Priority: domain.PriorityMedium,
	}, time.Now().UTC()); err != nil {
		t.Fatalf("insert: %v", err)
	}
	db.Close()

	db, err = Open(ctx, path)
	if err != nil {
		t.Fatalf("second Open returned error: %v", err)
	}
	defer db.Close()

	tasks, err := db.ListOpenActive(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(tasks) != 1 || tasks[0].Title != "survives" {
		t.Fatalf("tasks = %+v, want one task titled %q", tasks, "survives")
	}
}

func TestPragmas(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "taskgo.db"))
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer db.Close()

	var mode string
	if err := db.db.QueryRowContext(ctx, "PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatalf("read journal_mode: %v", err)
	}
	if strings.ToLower(mode) != "wal" {
		t.Fatalf("journal_mode = %q, want wal", mode)
	}

	var foreignKeys int
	if err := db.db.QueryRowContext(ctx, "PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		t.Fatalf("read foreign_keys: %v", err)
	}
	if foreignKeys != 1 {
		t.Fatalf("foreign_keys = %d, want 1", foreignKeys)
	}

	var busyTimeout int
	if err := db.db.QueryRowContext(ctx, "PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
		t.Fatalf("read busy_timeout: %v", err)
	}
	if busyTimeout != 5000 {
		t.Fatalf("busy_timeout = %d, want 5000", busyTimeout)
	}
}

func TestNewerSchemaIsRejected(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "taskgo.db")

	db, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	db.Close()

	raw, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		t.Fatalf("open raw: %v", err)
	}
	if _, err := raw.ExecContext(ctx, "PRAGMA user_version = 99"); err != nil {
		t.Fatalf("bump version: %v", err)
	}
	raw.Close()

	_, err = Open(ctx, path)
	if err == nil {
		t.Fatal("Open should reject a newer schema")
	}
	if !strings.Contains(err.Error(), "newer") {
		t.Fatalf("error = %q, want it to mention a newer schema", err)
	}
}

func TestFailedMigrationRollsBack(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "taskgo.db")

	raw, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		t.Fatalf("open raw: %v", err)
	}
	defer raw.Close()

	fsys := fstest.MapFS{
		"migrations/0001_broken.sql": &fstest.MapFile{
			Data: []byte("CREATE TABLE broken (id INTEGER);\nTHIS IS NOT SQL;"),
		},
	}
	if err := migrate(ctx, raw, fsys); err == nil {
		t.Fatal("migrate should fail on invalid SQL")
	}

	var version int
	if err := raw.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		t.Fatalf("read user_version: %v", err)
	}
	if version != 0 {
		t.Fatalf("user_version = %d, want 0 after rollback", version)
	}

	var name string
	err = raw.QueryRowContext(ctx,
		"SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'broken'").Scan(&name)
	if err == nil {
		t.Fatal("partially created table survived a failed migration")
	}
}

func TestMigration0002PreservesTasks(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "taskgo.db")

	initSQL, err := fs.ReadFile(migrationsFS, "migrations/0001_init.sql")
	if err != nil {
		t.Fatalf("read embedded 0001 migration: %v", err)
	}
	raw, err := sql.Open("sqlite", dsn(path))
	if err != nil {
		t.Fatalf("open raw: %v", err)
	}
	v1 := fstest.MapFS{
		"migrations/0001_init.sql": &fstest.MapFile{Data: initSQL},
	}
	if err := migrate(ctx, raw, v1); err != nil {
		t.Fatalf("apply schema 0001: %v", err)
	}

	const created = "2026-01-01T00:00:00.000000000Z"
	if _, err := raw.ExecContext(ctx,
		`INSERT INTO tasks (title, description, priority, created_at) VALUES ('legacy', NULL, 'medium', ?)`,
		created,
	); err != nil {
		t.Fatalf("seed a version 0001 task: %v", err)
	}
	raw.Close()

	db, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("upgrade to version 0002: %v", err)
	}
	defer db.Close()

	var version int
	if err := db.db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		t.Fatalf("read user_version: %v", err)
	}
	if version != 2 {
		t.Fatalf("user_version = %d, want 2", version)
	}

	var projectID sql.NullInt64
	if err := db.db.QueryRowContext(ctx,
		`SELECT project_id FROM tasks WHERE title = 'legacy'`,
	).Scan(&projectID); err != nil {
		t.Fatalf("read upgraded task: %v", err)
	}
	if projectID.Valid {
		t.Fatalf("legacy task gained project_id %d", projectID.Int64)
	}

	var projects int
	if err := db.db.QueryRowContext(ctx, `SELECT count(*) FROM projects`).Scan(&projects); err != nil {
		t.Fatalf("count projects: %v", err)
	}
	if projects != 0 {
		t.Fatalf("projects = %d, want 0 after an upgrade", projects)
	}
}

func TestListOrdersByPriorityThenAge(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "taskgo.db"))
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	defer db.Close()

	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	seed := []struct {
		title    string
		priority domain.Priority
		offset   time.Duration
	}{
		{"low", domain.PriorityLow, 0},
		{"high newer", domain.PriorityHigh, 2 * time.Second},
		{"medium", domain.PriorityMedium, 1 * time.Second},
		{"high older", domain.PriorityHigh, 0},
	}

	for _, item := range seed {
		if _, err := db.InsertTask(ctx, domain.NewTask{
			Title: item.title, Priority: item.priority,
		}, base.Add(item.offset)); err != nil {
			t.Fatalf("insert %q: %v", item.title, err)
		}
	}

	tasks, err := db.ListOpenActive(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	got := make([]string, 0, len(tasks))
	for _, task := range tasks {
		got = append(got, task.Title)
	}
	want := []string{"high older", "high newer", "medium", "low"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("order = %v, want %v", got, want)
	}
}
