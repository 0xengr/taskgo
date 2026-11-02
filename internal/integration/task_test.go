package integration

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestShowTask(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "--priority", "high", "-d", "details here", "show me")

	got := run(t, db, "show", "1")
	if got.exit != 0 {
		t.Fatalf("show exit = %d, stderr = %q", got.exit, got.stderr)
	}
	for _, want := range []string{"show me", "high", "Inbox", "open", "details here"} {
		if !strings.Contains(got.stdout, want) {
			t.Fatalf("show stdout = %q, want it to contain %q", got.stdout, want)
		}
	}
}

func TestShowAcceptsHashPrefix(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "hash prefixed")

	got := run(t, db, "show", "#1")
	if got.exit != 0 {
		t.Fatalf("show exit = %d, stderr = %q", got.exit, got.stderr)
	}
	if !strings.Contains(got.stdout, "hash prefixed") {
		t.Fatalf("show stdout = %q", got.stdout)
	}
}

func TestShowUnknownTask(t *testing.T) {
	db := newDB(t)

	got := run(t, db, "show", "99")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
	if !strings.Contains(got.stderr, "not found") {
		t.Fatalf("stderr = %q", got.stderr)
	}
}

func TestInvalidTaskID(t *testing.T) {
	db := newDB(t)

	for _, raw := range []string{"abc", "0", "-3"} {
		got := run(t, db, "show", raw)
		if got.exit != 2 {
			t.Fatalf("show %q exit = %d, want 2", raw, got.exit)
		}
	}
}

func TestDoneIsIdempotent(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "finish me")

	first := run(t, db, "done", "--json", "1")
	if first.exit != 0 {
		t.Fatalf("first done exit = %d, stderr = %q", first.exit, first.stderr)
	}
	var firstTask taskJSON
	if err := json.Unmarshal([]byte(first.stdout), &firstTask); err != nil {
		t.Fatalf("decode first done: %v", err)
	}
	if firstTask.CompletedAt == nil {
		t.Fatal("completed_at is null after done")
	}
	if _, err := time.Parse(time.RFC3339, *firstTask.CompletedAt); err != nil {
		t.Fatalf("completed_at %q is not RFC3339: %v", *firstTask.CompletedAt, err)
	}

	second := run(t, db, "done", "--json", "1")
	if second.exit != 0 {
		t.Fatalf("second done exit = %d, stderr = %q", second.exit, second.stderr)
	}
	var secondTask taskJSON
	if err := json.Unmarshal([]byte(second.stdout), &secondTask); err != nil {
		t.Fatalf("decode second done: %v", err)
	}
	if secondTask.CompletedAt == nil || *secondTask.CompletedAt != *firstTask.CompletedAt {
		t.Fatalf("completed_at changed: %v -> %v", firstTask.CompletedAt, secondTask.CompletedAt)
	}

	listed := run(t, db, "list")
	if strings.Contains(listed.stdout, "finish me") {
		t.Fatalf("list stdout = %q, want the done task hidden", listed.stdout)
	}
}

func TestReopen(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "reopen me")
	run(t, db, "done", "1")

	first := run(t, db, "reopen", "--json", "1")
	if first.exit != 0 {
		t.Fatalf("first reopen exit = %d, stderr = %q", first.exit, first.stderr)
	}
	var firstTask taskJSON
	if err := json.Unmarshal([]byte(first.stdout), &firstTask); err != nil {
		t.Fatalf("decode first reopen: %v", err)
	}
	if firstTask.CompletedAt != nil {
		t.Fatalf("completed_at = %v, want null", firstTask.CompletedAt)
	}

	second := run(t, db, "reopen", "--json", "1")
	if second.exit != 0 {
		t.Fatalf("second reopen exit = %d, stderr = %q", second.exit, second.stderr)
	}

	listed := run(t, db, "list")
	if !strings.Contains(listed.stdout, "reopen me") {
		t.Fatalf("list stdout = %q, want the task back", listed.stdout)
	}
}

func TestDoneUnknownTask(t *testing.T) {
	db := newDB(t)

	got := run(t, db, "done", "42")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
}

func TestShowIncludesProject(t *testing.T) {
	db := newDB(t)
	repo := t.TempDir()
	run(t, db, "project", "add", "--repo", repo, "API")
	runIn(t, repo, db, "add", "project task")

	got := run(t, db, "show", "1")
	if !strings.Contains(got.stdout, "API") {
		t.Fatalf("show stdout = %q, want the project", got.stdout)
	}
}
