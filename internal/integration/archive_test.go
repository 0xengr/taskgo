package integration

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestArchiveTaskHidesAndUnarchiveRestores(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "archive me")

	archived := run(t, db, "archive", "1")
	if archived.exit != 0 {
		t.Fatalf("archive exit = %d, stderr = %q", archived.exit, archived.stderr)
	}

	hidden := run(t, db, "list")
	if strings.Contains(hidden.stdout, "archive me") {
		t.Fatalf("default list = %q, want the archived task hidden", hidden.stdout)
	}

	shown := run(t, db, "list", "--visibility", "archived")
	if !strings.Contains(shown.stdout, "archive me") || !strings.Contains(shown.stdout, "(archived)") {
		t.Fatalf("archived list = %q, want the task marked archived", shown.stdout)
	}

	restored := run(t, db, "unarchive", "1")
	if restored.exit != 0 {
		t.Fatalf("unarchive exit = %d, stderr = %q", restored.exit, restored.stderr)
	}
	back := run(t, db, "list")
	if !strings.Contains(back.stdout, "archive me") {
		t.Fatalf("list after unarchive = %q", back.stdout)
	}
}

func TestArchiveDoesNotComplete(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "still open")
	run(t, db, "archive", "1")

	shown := run(t, db, "show", "1")
	for _, want := range []string{"Status:     open", "Archived:   yes"} {
		if !strings.Contains(shown.stdout, want) {
			t.Fatalf("show stdout = %q, want %q", shown.stdout, want)
		}
	}
}

func TestArchiveIsIdempotent(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "archive twice")

	first := run(t, db, "archive", "--json", "1")
	var firstTask taskJSON
	if err := json.Unmarshal([]byte(first.stdout), &firstTask); err != nil {
		t.Fatalf("decode first: %v", err)
	}
	if firstTask.ArchivedAt == nil {
		t.Fatal("archived_at is null after archive")
	}

	second := run(t, db, "archive", "--json", "1")
	var secondTask taskJSON
	if err := json.Unmarshal([]byte(second.stdout), &secondTask); err != nil {
		t.Fatalf("decode second: %v", err)
	}
	if secondTask.ArchivedAt == nil || *secondTask.ArchivedAt != *firstTask.ArchivedAt {
		t.Fatalf("archived_at changed: %v -> %v", firstTask.ArchivedAt, secondTask.ArchivedAt)
	}
}

func TestUnarchiveIsIdempotent(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "not archived")

	got := run(t, db, "unarchive", "--json", "1")
	if got.exit != 0 {
		t.Fatalf("exit = %d, stderr = %q", got.exit, got.stderr)
	}
	var task taskJSON
	if err := json.Unmarshal([]byte(got.stdout), &task); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if task.ArchivedAt != nil {
		t.Fatalf("archived_at = %v, want null", task.ArchivedAt)
	}
}

func TestArchivedTaskRemainsAddressable(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "archived but reachable")
	run(t, db, "archive", "1")

	shown := run(t, db, "show", "1")
	if shown.exit != 0 || !strings.Contains(shown.stdout, "archived but reachable") {
		t.Fatalf("show exit = %d, stdout = %q", shown.exit, shown.stdout)
	}

	done := run(t, db, "done", "1")
	if done.exit != 0 {
		t.Fatalf("done on archived task exit = %d, stderr = %q", done.exit, done.stderr)
	}
}

func TestListStateDimension(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "open task")
	run(t, db, "add", "done task")
	run(t, db, "done", "2")

	open := run(t, db, "list", "--state", "open")
	if !strings.Contains(open.stdout, "open task") || strings.Contains(open.stdout, "done task") {
		t.Fatalf("open list = %q", open.stdout)
	}

	done := run(t, db, "list", "--state", "done")
	if !strings.Contains(done.stdout, "done task") || strings.Contains(done.stdout, "open task") {
		t.Fatalf("done list = %q", done.stdout)
	}
}

func TestVisibilityArchivedIncludesArchivedProjectTasks(t *testing.T) {
	db := newDB(t)
	repo := t.TempDir()
	run(t, db, "project", "add", "--repo", repo, "API")
	runIn(t, repo, db, "add", "inside archived project")
	run(t, db, "project", "archive", "API")

	active := run(t, db, "list")
	if strings.Contains(active.stdout, "inside archived project") {
		t.Fatalf("active list = %q, want the project to hide its task", active.stdout)
	}

	archived := run(t, db, "list", "--visibility", "archived")
	if !strings.Contains(archived.stdout, "inside archived project") {
		t.Fatalf("archived list = %q, want the archived project's task", archived.stdout)
	}

	all := run(t, db, "list", "--all")
	if !strings.Contains(all.stdout, "inside archived project") {
		t.Fatalf("all list = %q, want the archived project's task", all.stdout)
	}
}

func TestAllIncludesDoneAndArchived(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "normal")
	run(t, db, "add", "done")
	run(t, db, "done", "2")
	run(t, db, "add", "archived")
	run(t, db, "archive", "3")

	all := run(t, db, "list", "--all")
	for _, want := range []string{"normal", "done", "archived"} {
		if !strings.Contains(all.stdout, want) {
			t.Fatalf("all list = %q, want %q", all.stdout, want)
		}
	}

	def := run(t, db, "list")
	if !strings.Contains(def.stdout, "normal") ||
		strings.Contains(def.stdout, "done") ||
		strings.Contains(def.stdout, "archived") {
		t.Fatalf("default list = %q, want only the normal task", def.stdout)
	}
}

func TestAllConflictsWithExplicitDimension(t *testing.T) {
	db := newDB(t)

	if got := run(t, db, "list", "--all", "--state", "open"); got.exit != 2 {
		t.Fatalf("--all --state exit = %d, want 2", got.exit)
	}
	if got := run(t, db, "list", "--all", "--visibility", "active"); got.exit != 2 {
		t.Fatalf("--all --visibility exit = %d, want 2", got.exit)
	}
}

func TestInvalidStateIsUsageError(t *testing.T) {
	db := newDB(t)

	got := run(t, db, "list", "--state", "weird")
	if got.exit != 2 {
		t.Fatalf("exit = %d, want 2", got.exit)
	}
}
