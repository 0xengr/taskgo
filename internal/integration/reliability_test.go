package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestCorruptDatabaseIsLeftUntouched(t *testing.T) {
	db := newDB(t)
	if err := os.MkdirAll(filepath.Dir(db), 0o700); err != nil {
		t.Fatalf("create db dir: %v", err)
	}
	garbage := []byte("this is definitely not a sqlite database")
	if err := os.WriteFile(db, garbage, 0o600); err != nil {
		t.Fatalf("seed corrupt db: %v", err)
	}

	got := run(t, db, "list")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
	if !strings.Contains(got.stderr, "corrupt") {
		t.Fatalf("stderr = %q, want a corruption message", got.stderr)
	}
	if !strings.Contains(got.stderr, db) {
		t.Fatalf("stderr = %q, want the database path", got.stderr)
	}

	after, err := os.ReadFile(db)
	if err != nil {
		t.Fatalf("read database after failure: %v", err)
	}
	if string(after) != string(garbage) {
		t.Fatal("corrupt database was modified")
	}
}

func TestNewerSchemaIsRejectedAtTheProcessBoundary(t *testing.T) {
	db := newDB(t)
	if got := run(t, db, "add", "seed"); got.exit != 0 {
		t.Fatalf("seed add exit = %d, stderr = %q", got.exit, got.stderr)
	}

	raw, err := sql.Open("sqlite", db)
	if err != nil {
		t.Fatalf("open raw: %v", err)
	}
	if _, err := raw.Exec("PRAGMA user_version = 99"); err != nil {
		t.Fatalf("bump version: %v", err)
	}
	raw.Close()

	got := run(t, db, "list")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
	if !strings.Contains(got.stderr, "newer") {
		t.Fatalf("stderr = %q, want a newer-schema message", got.stderr)
	}
}

func TestLockTimeoutProducesActionableError(t *testing.T) {
	db := newDB(t)
	if got := run(t, db, "add", "seed"); got.exit != 0 {
		t.Fatalf("seed add exit = %d, stderr = %q", got.exit, got.stderr)
	}

	raw, err := sql.Open("sqlite", db)
	if err != nil {
		t.Fatalf("open raw: %v", err)
	}
	defer raw.Close()
	conn, err := raw.Conn(context.Background())
	if err != nil {
		t.Fatalf("get connection: %v", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(context.Background(), "BEGIN IMMEDIATE"); err != nil {
		t.Fatalf("acquire write lock: %v", err)
	}
	defer conn.ExecContext(context.Background(), "ROLLBACK")

	start := time.Now()
	got := run(t, db, "add", "blocked")
	elapsed := time.Since(start)

	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
	if !strings.Contains(got.stderr, "busy") {
		t.Fatalf("stderr = %q, want a busy message", got.stderr)
	}
	if elapsed < 3*time.Second {
		t.Fatalf("returned after %v, want it to wait for the busy timeout", elapsed)
	}
	if elapsed > 15*time.Second {
		t.Fatalf("waited %v, want a bounded wait", elapsed)
	}

	if _, err := conn.ExecContext(context.Background(), "ROLLBACK"); err != nil {
		t.Fatalf("release lock: %v", err)
	}
	if after := run(t, db, "add", "after unlock"); after.exit != 0 {
		t.Fatalf("add after unlock exit = %d, stderr = %q", after.exit, after.stderr)
	}
}

func TestConcurrentFirstRun(t *testing.T) {
	db := newDB(t)
	dir := t.TempDir()

	const starters = 8
	failures := make(chan error, starters)
	var wait sync.WaitGroup

	for i := 0; i < starters; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			got, err := execute(dir, db, "list")
			if err != nil {
				failures <- err
				return
			}
			if got.exit != 0 {
				failures <- fmt.Errorf("list exit %d: %s", got.exit, got.stderr)
			}
		}()
	}

	wait.Wait()
	close(failures)
	for err := range failures {
		t.Error(err)
	}
	if t.Failed() {
		t.FailNow()
	}

	if got := run(t, db, "add", "after the race"); got.exit != 0 {
		t.Fatalf("add after concurrent migration exit = %d, stderr = %q", got.exit, got.stderr)
	}
}

func TestConcurrentCaptures(t *testing.T) {
	db := newDB(t)
	dir := t.TempDir()
	if got := runIn(t, dir, db, "list"); got.exit != 0 {
		t.Fatalf("initial list exit = %d, stderr = %q", got.exit, got.stderr)
	}

	const writers = 8
	failures := make(chan error, writers)
	var wait sync.WaitGroup

	for i := 0; i < writers; i++ {
		wait.Add(1)
		go func(i int) {
			defer wait.Done()
			got, err := execute(dir, db, "add", fmt.Sprintf("concurrent %d", i))
			if err != nil {
				failures <- err
				return
			}
			if got.exit != 0 {
				failures <- fmt.Errorf("add %d exit %d: %s", i, got.exit, got.stderr)
			}
		}(i)
	}

	wait.Wait()
	close(failures)
	for err := range failures {
		t.Error(err)
	}
	if t.Failed() {
		t.FailNow()
	}

	listed := run(t, db, "list", "--json")
	var tasks []taskJSON
	if err := json.Unmarshal([]byte(listed.stdout), &tasks); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(tasks) != writers {
		t.Fatalf("tasks = %d, want %d", len(tasks), writers)
	}

	seen := make(map[int64]bool, writers)
	for _, task := range tasks {
		if seen[task.ID] {
			t.Fatalf("duplicate task id %d", task.ID)
		}
		seen[task.ID] = true
	}
}
