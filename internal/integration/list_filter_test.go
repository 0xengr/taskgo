package integration

import (
	"strings"
	"testing"
)

func TestListFiltersByProject(t *testing.T) {
	db := newDB(t)
	repo := t.TempDir()
	run(t, db, "project", "add", "--repo", repo, "API")
	run(t, db, "add", "inbox task")
	runIn(t, repo, db, "add", "project task")

	got := run(t, db, "list", "--project", "api")
	if !strings.Contains(got.stdout, "project task") || strings.Contains(got.stdout, "inbox task") {
		t.Fatalf("list --project = %q", got.stdout)
	}
}

func TestListFiltersByPriority(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "--priority", "low", "low task")
	run(t, db, "add", "--priority", "medium", "medium task")
	run(t, db, "add", "--priority", "high", "high task")

	got := run(t, db, "list", "--priority", "high")
	if !strings.Contains(got.stdout, "high task") ||
		strings.Contains(got.stdout, "low task") ||
		strings.Contains(got.stdout, "medium task") {
		t.Fatalf("list --priority = %q", got.stdout)
	}
}

func TestListHere(t *testing.T) {
	db := newDB(t)
	repo := t.TempDir()
	run(t, db, "project", "add", "--repo", repo, "API")
	run(t, db, "add", "inbox task")
	runIn(t, repo, db, "add", "project task")

	inside := runIn(t, repo, db, "list", "--here")
	if !strings.Contains(inside.stdout, "project task") || strings.Contains(inside.stdout, "inbox task") {
		t.Fatalf("list --here = %q", inside.stdout)
	}

	outside := run(t, db, "list", "--here")
	if outside.exit != 1 {
		t.Fatalf("list --here outside exit = %d, want 1", outside.exit)
	}
}

func TestListScopeConflictsAndInvalidPriority(t *testing.T) {
	db := newDB(t)

	if got := run(t, db, "list", "--project", "API", "--here"); got.exit != 2 {
		t.Fatalf("scope conflict exit = %d, want 2", got.exit)
	}
	if got := run(t, db, "list", "--priority", "urgent"); got.exit != 2 {
		t.Fatalf("invalid priority exit = %d, want 2", got.exit)
	}
	if got := run(t, db, "list", "--priority", ""); got.exit != 2 {
		t.Fatalf("empty priority exit = %d, want 2", got.exit)
	}
}
