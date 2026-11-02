package integration

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestDeleteWithForce(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "delete me")

	deleted := run(t, db, "delete", "--force", "1")
	if deleted.exit != 0 {
		t.Fatalf("delete exit = %d, stderr = %q", deleted.exit, deleted.stderr)
	}
	if !strings.Contains(deleted.stdout, "Deleted #1") {
		t.Fatalf("delete stdout = %q", deleted.stdout)
	}

	if got := run(t, db, "show", "1"); got.exit != 1 {
		t.Fatalf("show exit = %d, want 1 after deletion", got.exit)
	}
	if listed := run(t, db, "list"); strings.Contains(listed.stdout, "delete me") {
		t.Fatalf("list stdout = %q, want the task gone", listed.stdout)
	}
}

func TestDeleteJSONReturnsDeletedTask(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "--priority", "high", "-d", "snapshot", "delete with json")

	got := run(t, db, "delete", "--force", "--json", "1")
	if got.exit != 0 {
		t.Fatalf("delete exit = %d, stderr = %q", got.exit, got.stderr)
	}

	var task taskJSON
	if err := json.Unmarshal([]byte(got.stdout), &task); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if task.ID != 1 || task.Title != "delete with json" || task.Priority != "high" {
		t.Fatalf("task = %+v, want the pre-deletion resource", task)
	}
	if task.Description == nil || *task.Description != "snapshot" {
		t.Fatalf("description = %v", task.Description)
	}
}

func TestDeleteUnknownTask(t *testing.T) {
	db := newDB(t)

	got := run(t, db, "delete", "--force", "99")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
	if !strings.Contains(got.stderr, "not found") {
		t.Fatalf("stderr = %q", got.stderr)
	}
}

func TestDeleteWithoutForceRefusesNonInteractive(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "protected")

	got := run(t, db, "delete", "1")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
	if !strings.Contains(got.stderr, "--force") {
		t.Fatalf("stderr = %q, want guidance to use --force", got.stderr)
	}
	if shown := run(t, db, "show", "1"); shown.exit != 0 {
		t.Fatalf("task was removed despite refusal (show exit %d)", shown.exit)
	}
}

func TestDeletedTaskIDIsNotReused(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "first")
	run(t, db, "delete", "--force", "1")

	added := run(t, db, "add", "second")
	if !strings.Contains(added.stdout, "#2") {
		t.Fatalf("add stdout = %q, want id 2 after deleting id 1", added.stdout)
	}
}
