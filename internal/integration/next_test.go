package integration

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

func TestNextReturnsHighestPriority(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "--priority", "medium", "medium task")
	run(t, db, "add", "--priority", "high", "high task")
	run(t, db, "add", "--priority", "low", "low task")

	got := run(t, db, "next")
	if got.exit != 0 {
		t.Fatalf("next exit = %d, stderr = %q", got.exit, got.stderr)
	}
	if !strings.Contains(got.stdout, "high task") {
		t.Fatalf("next stdout = %q, want the high task", got.stdout)
	}
	if strings.Count(strings.TrimSpace(got.stdout), "\n") != 0 {
		t.Fatalf("next returned more than one line: %q", got.stdout)
	}
}

func TestNextCount(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "--priority", "high", "first high")
	run(t, db, "add", "--priority", "high", "second high")
	run(t, db, "add", "--priority", "medium", "medium")

	got := run(t, db, "next", "--count", "2")
	lines := strings.Split(strings.TrimSpace(got.stdout), "\n")
	if len(lines) != 2 {
		t.Fatalf("next --count 2 returned %d lines: %q", len(lines), got.stdout)
	}
	if !strings.Contains(lines[0], "first high") || !strings.Contains(lines[1], "second high") {
		t.Fatalf("order = %q, want the two high tasks oldest first", got.stdout)
	}
}

func TestNextEmpty(t *testing.T) {
	db := newDB(t)

	human := run(t, db, "next")
	if human.exit != 0 {
		t.Fatalf("human exit = %d, want 0", human.exit)
	}
	if !strings.Contains(human.stdout, "Nothing open") {
		t.Fatalf("human stdout = %q", human.stdout)
	}

	asJSON := run(t, db, "next", "--json")
	if asJSON.exit != 0 {
		t.Fatalf("json exit = %d, want 0", asJSON.exit)
	}
	if strings.TrimSpace(asJSON.stdout) != "null" {
		t.Fatalf("json stdout = %q, want null", asJSON.stdout)
	}
}

func TestNextJSONShape(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "one")
	run(t, db, "add", "two")

	single := run(t, db, "next", "--json")
	var task taskJSON
	if err := json.Unmarshal([]byte(single.stdout), &task); err != nil {
		t.Fatalf("count=1 should be a single object: %v (%q)", err, single.stdout)
	}
	if task.ID == 0 {
		t.Fatalf("task = %+v", task)
	}

	multiple := run(t, db, "next", "--count", "2", "--json")
	var tasks []taskJSON
	if err := json.Unmarshal([]byte(multiple.stdout), &tasks); err != nil {
		t.Fatalf("count=2 should be an array: %v (%q)", err, multiple.stdout)
	}
	if len(tasks) != 2 {
		t.Fatalf("count=2 returned %d tasks", len(tasks))
	}
}

func TestNextScopes(t *testing.T) {
	db := newDB(t)
	repo := t.TempDir()
	run(t, db, "project", "add", "--repo", repo, "API")
	run(t, db, "add", "inbox task")
	runIn(t, repo, db, "add", "project task")

	inProject := run(t, db, "next", "--project", "api")
	if !strings.Contains(inProject.stdout, "project task") {
		t.Fatalf("next --project = %q", inProject.stdout)
	}

	here := runIn(t, repo, db, "next", "--here")
	if !strings.Contains(here.stdout, "project task") {
		t.Fatalf("next --here = %q", here.stdout)
	}

	outside := run(t, db, "next", "--here")
	if outside.exit != 1 {
		t.Fatalf("next --here outside exit = %d, want 1", outside.exit)
	}
}

func TestNextHereStopsAtArchivedProject(t *testing.T) {
	db := newDB(t)
	outer := t.TempDir()
	inner := filepath.Join(outer, "inner")
	mustMkdir(t, inner)
	run(t, db, "project", "add", "--repo", outer, "Outer")
	run(t, db, "project", "add", "--repo", inner, "Inner")
	run(t, db, "project", "archive", "Inner")

	got := runIn(t, inner, db, "next", "--here")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1 at the archived barrier", got.exit)
	}
	if !strings.Contains(got.stderr, "active project") {
		t.Fatalf("stderr = %q", got.stderr)
	}
}

func TestNextExcludesDoneAndArchived(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "--priority", "high", "done high")
	run(t, db, "done", "1")
	run(t, db, "add", "--priority", "high", "archived high")
	run(t, db, "archive", "2")
	run(t, db, "add", "--priority", "medium", "available medium")

	got := run(t, db, "next")
	if !strings.Contains(got.stdout, "available medium") {
		t.Fatalf("next stdout = %q, want the available task", got.stdout)
	}
}

func TestNextCountValidation(t *testing.T) {
	db := newDB(t)

	for _, count := range []string{"0", "-1"} {
		got := run(t, db, "next", "--count", count)
		if got.exit != 2 {
			t.Fatalf("--count %s exit = %d, want 2", count, got.exit)
		}
	}
}

func TestNextConflictingScopes(t *testing.T) {
	db := newDB(t)

	got := run(t, db, "next", "--project", "API", "--here")
	if got.exit != 2 {
		t.Fatalf("exit = %d, want 2", got.exit)
	}
}

func TestNextUnknownProject(t *testing.T) {
	db := newDB(t)

	got := run(t, db, "next", "--project", "Nope")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
}

func TestNextDoesNotMutate(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "stable")

	first := run(t, db, "next", "--json")
	second := run(t, db, "next", "--json")
	if first.stdout != second.stdout {
		t.Fatalf("next changed its answer: %q -> %q", first.stdout, second.stdout)
	}
	if listed := run(t, db, "list"); !strings.Contains(listed.stdout, "stable") {
		t.Fatalf("list after next = %q", listed.stdout)
	}
}
