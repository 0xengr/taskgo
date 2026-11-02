package integration

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

type taskJSON struct {
	ID          int64   `json:"id"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Priority    string  `json:"priority"`
	Project     any     `json:"project"`
	CreatedAt   string  `json:"created_at"`
	CompletedAt *string `json:"completed_at"`
	ArchivedAt  *string `json:"archived_at"`
}

func TestAddCreatesTaskInInbox(t *testing.T) {
	db := newDB(t)

	added := run(t, db, "add", "write", "tests")
	if added.exit != 0 {
		t.Fatalf("add exit = %d, stderr = %q", added.exit, added.stderr)
	}
	if !strings.Contains(added.stdout, "#1") || !strings.Contains(added.stdout, "Inbox") {
		t.Fatalf("add stdout = %q, want the id and destination", added.stdout)
	}

	listed := run(t, db, "list")
	if listed.exit != 0 {
		t.Fatalf("list exit = %d, stderr = %q", listed.exit, listed.stderr)
	}
	if !strings.Contains(listed.stdout, "write tests") {
		t.Fatalf("list stdout = %q, want the captured title", listed.stdout)
	}
}

func TestAddNormalizesTitle(t *testing.T) {
	db := newDB(t)

	if got := run(t, db, "add", "  fix \t  it "); got.exit != 0 {
		t.Fatalf("add exit = %d, stderr = %q", got.exit, got.stderr)
	}

	listed := run(t, db, "list")
	if !strings.Contains(listed.stdout, "fix it") {
		t.Fatalf("list stdout = %q, want normalized title", listed.stdout)
	}
	if strings.Contains(listed.stdout, "\t") {
		t.Fatalf("list stdout still contains a tab: %q", listed.stdout)
	}
}

func TestAddJSONReturnsCreatedTask(t *testing.T) {
	db := newDB(t)

	got := run(t, db, "add", "--json", "-d", "line one\nline two", "--priority", "HIGH", "ship it")
	if got.exit != 0 {
		t.Fatalf("add exit = %d, stderr = %q", got.exit, got.stderr)
	}

	var task taskJSON
	if err := json.Unmarshal([]byte(got.stdout), &task); err != nil {
		t.Fatalf("decode JSON %q: %v", got.stdout, err)
	}
	if task.ID != 1 {
		t.Fatalf("id = %d, want 1", task.ID)
	}
	if task.Title != "ship it" {
		t.Fatalf("title = %q, want %q", task.Title, "ship it")
	}
	if task.Priority != "high" {
		t.Fatalf("priority = %q, want high", task.Priority)
	}
	if task.Description == nil || *task.Description != "line one\nline two" {
		t.Fatalf("description = %v, want multiline description", task.Description)
	}
	if task.Project != nil {
		t.Fatalf("project = %v, want null in this slice", task.Project)
	}
	if task.CompletedAt != nil || task.ArchivedAt != nil {
		t.Fatalf("completed_at/archived_at = %v/%v, want null", task.CompletedAt, task.ArchivedAt)
	}
	if _, err := time.Parse(time.RFC3339, task.CreatedAt); err != nil {
		t.Fatalf("created_at %q is not RFC3339: %v", task.CreatedAt, err)
	}
}

func TestJSONListEmptyIsArray(t *testing.T) {
	db := newDB(t)

	got := run(t, db, "list", "--json")
	if got.exit != 0 {
		t.Fatalf("list exit = %d, stderr = %q", got.exit, got.stderr)
	}
	if strings.TrimSpace(got.stdout) != "[]" {
		t.Fatalf("empty list JSON = %q, want []", got.stdout)
	}
}

func TestListOrdersByPriority(t *testing.T) {
	db := newDB(t)

	for _, priority := range []string{"low", "medium", "high"} {
		if got := run(t, db, "add", "--priority", priority, priority+" task"); got.exit != 0 {
			t.Fatalf("add %s exit = %d, stderr = %q", priority, got.exit, got.stderr)
		}
	}

	listed := run(t, db, "list")
	if listed.exit != 0 {
		t.Fatalf("list exit = %d, stderr = %q", listed.exit, listed.stderr)
	}

	var priorities []string
	for _, line := range strings.Split(strings.TrimSpace(listed.stdout), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			priorities = append(priorities, fields[1])
		}
	}
	want := []string{"high", "medium", "low"}
	if strings.Join(priorities, ",") != strings.Join(want, ",") {
		t.Fatalf("priorities = %v, want %v (output %q)", priorities, want, listed.stdout)
	}
}

func TestInvalidPriorityIsUsageError(t *testing.T) {
	db := newDB(t)

	got := run(t, db, "add", "--priority", "urgent", "nope")
	if got.exit != 2 {
		t.Fatalf("exit = %d, want 2", got.exit)
	}
	if !strings.Contains(got.stderr, "priority") {
		t.Fatalf("stderr = %q, want a priority message", got.stderr)
	}
	if got.stdout != "" {
		t.Fatalf("stdout = %q, want empty", got.stdout)
	}
}

func TestEmptyTitleIsUsageError(t *testing.T) {
	db := newDB(t)

	got := run(t, db, "add", "")
	if got.exit != 2 {
		t.Fatalf("exit = %d, want 2", got.exit)
	}
	if !strings.Contains(got.stderr, "title") {
		t.Fatalf("stderr = %q, want a title message", got.stderr)
	}
}

func TestUnknownCommandIsUsageError(t *testing.T) {
	got := run(t, newDB(t), "frobnicate")
	if got.exit != 2 {
		t.Fatalf("exit = %d, want 2 (stderr %q)", got.exit, got.stderr)
	}
}

func TestUnknownFlagIsUsageError(t *testing.T) {
	got := run(t, newDB(t), "list", "--nope")
	if got.exit != 2 {
		t.Fatalf("exit = %d, want 2 (stderr %q)", got.exit, got.stderr)
	}
}

func TestDatabaseFileIsCreated(t *testing.T) {
	db := newDB(t)

	if got := run(t, db, "add", "persisted"); got.exit != 0 {
		t.Fatalf("add exit = %d, stderr = %q", got.exit, got.stderr)
	}
	info, err := os.Stat(db)
	if err != nil {
		t.Fatalf("stat database: %v", err)
	}
	if info.IsDir() {
		t.Fatal("database path is a directory")
	}
}

func TestHelpDoesNotCreateDatabase(t *testing.T) {
	db := newDB(t)

	got := run(t, db, "--help")
	if got.exit != 0 {
		t.Fatalf("help exit = %d, stderr = %q", got.exit, got.stderr)
	}
	if _, err := os.Stat(db); !os.IsNotExist(err) {
		t.Fatalf("help created a database (%v)", err)
	}
}

func TestPipedOutputHasNoANSI(t *testing.T) {
	db := newDB(t)

	if got := run(t, db, "add", "--priority", "high", "colorless"); got.exit != 0 {
		t.Fatalf("add exit = %d, stderr = %q", got.exit, got.stderr)
	}
	listed := run(t, db, "list")
	if strings.Contains(listed.stdout, "\x1b[") {
		t.Fatalf("piped output contains ANSI escapes: %q", listed.stdout)
	}
}
