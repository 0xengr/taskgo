package integration

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEditMultipleFields(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "old title", "-d", "old description", "--priority", "low")

	edited := run(t, db, "edit", "--json", "1",
		"--title", "new title",
		"--description", "new description",
		"--priority", "high",
	)
	if edited.exit != 0 {
		t.Fatalf("edit exit = %d, stderr = %q", edited.exit, edited.stderr)
	}

	var task taskJSON
	if err := json.Unmarshal([]byte(edited.stdout), &task); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if task.Title != "new title" {
		t.Fatalf("title = %q", task.Title)
	}
	if task.Description == nil || *task.Description != "new description" {
		t.Fatalf("description = %v", task.Description)
	}
	if task.Priority != "high" {
		t.Fatalf("priority = %q", task.Priority)
	}
}

func TestEditClearsDescription(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "task", "-d", "remove me")

	edited := run(t, db, "edit", "--json", "1", "-d", "")
	if edited.exit != 0 {
		t.Fatalf("edit exit = %d, stderr = %q", edited.exit, edited.stderr)
	}
	var task taskJSON
	if err := json.Unmarshal([]byte(edited.stdout), &task); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if task.Description != nil {
		t.Fatalf("description = %v, want null", task.Description)
	}
}

func TestEditLeavesOmittedFieldsAlone(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "keep description", "-d", "keep me", "--priority", "low")

	edited := run(t, db, "edit", "--json", "1", "--title", "only the title")
	var task taskJSON
	if err := json.Unmarshal([]byte(edited.stdout), &task); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if task.Description == nil || *task.Description != "keep me" {
		t.Fatalf("description = %v, want unchanged", task.Description)
	}
	if task.Priority != "low" {
		t.Fatalf("priority = %q, want low", task.Priority)
	}
}

func TestEditMovesBetweenProjectAndInbox(t *testing.T) {
	db := newDB(t)
	run(t, db, "project", "add", "API")
	run(t, db, "add", "movable")

	moved := run(t, db, "edit", "--json", "1", "--project", "api")
	if moved.exit != 0 {
		t.Fatalf("move exit = %d, stderr = %q", moved.exit, moved.stderr)
	}
	var task taskJSON
	if err := json.Unmarshal([]byte(moved.stdout), &task); err != nil {
		t.Fatalf("decode moved: %v", err)
	}
	if task.Project == nil || task.Project.Name != "API" {
		t.Fatalf("project = %+v, want API", task.Project)
	}

	back := run(t, db, "edit", "--json", "1", "--inbox")
	if err := json.Unmarshal([]byte(back.stdout), &task); err != nil {
		t.Fatalf("decode inbox: %v", err)
	}
	if task.Project != nil {
		t.Fatalf("project = %+v, want null", task.Project)
	}
}

func TestEditIntoArchivedProjectFails(t *testing.T) {
	db := newDB(t)
	run(t, db, "project", "add", "API")
	run(t, db, "project", "archive", "API")
	run(t, db, "add", "task")

	got := run(t, db, "edit", "1", "--project", "API")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
	if !strings.Contains(got.stderr, "archived") {
		t.Fatalf("stderr = %q", got.stderr)
	}
}

func TestEditUsageErrors(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "task")

	if got := run(t, db, "edit", "1"); got.exit != 2 {
		t.Fatalf("no flags exit = %d, want 2", got.exit)
	}
	if got := run(t, db, "edit", "1", "--project", "x", "--inbox"); got.exit != 2 {
		t.Fatalf("both flags exit = %d, want 2", got.exit)
	}
	if got := run(t, db, "edit", "1", "--priority", ""); got.exit != 2 {
		t.Fatalf("empty priority exit = %d, want 2", got.exit)
	}
	if got := run(t, db, "edit", "1", "--title", ""); got.exit != 2 {
		t.Fatalf("empty title exit = %d, want 2", got.exit)
	}
}

func TestEditUnknownTaskAndProject(t *testing.T) {
	db := newDB(t)

	if got := run(t, db, "edit", "99", "--title", "x"); got.exit != 1 {
		t.Fatalf("unknown task exit = %d, want 1", got.exit)
	}

	run(t, db, "add", "task")
	if got := run(t, db, "edit", "1", "--project", "Nope"); got.exit != 1 {
		t.Fatalf("unknown project exit = %d, want 1", got.exit)
	}
}

func TestEditArchivedTask(t *testing.T) {
	db := newDB(t)
	run(t, db, "add", "archived editable")
	run(t, db, "archive", "1")

	edited := run(t, db, "edit", "--json", "1", "--title", "still editable")
	if edited.exit != 0 {
		t.Fatalf("exit = %d, stderr = %q", edited.exit, edited.stderr)
	}
	var task taskJSON
	if err := json.Unmarshal([]byte(edited.stdout), &task); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if task.Title != "still editable" || task.ArchivedAt == nil {
		t.Fatalf("task = %+v, want new title and archive state preserved", task)
	}
}
