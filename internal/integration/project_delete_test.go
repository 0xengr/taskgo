package integration

import (
	"encoding/json"
	"strings"
	"testing"
)

type projectDeletionJSON struct {
	Project    projectJSON `json:"project"`
	MovedTasks *int64      `json:"moved_tasks"`
}

func TestProjectDeleteRefusesWithTasks(t *testing.T) {
	db := newDB(t)
	run(t, db, "project", "add", "API")
	run(t, db, "add", "--project", "API", "task")

	got := run(t, db, "project", "delete", "API")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
	for _, want := range []string{"--to-inbox", "--force"} {
		if !strings.Contains(got.stderr, want) {
			t.Fatalf("stderr = %q, want %q", got.stderr, want)
		}
	}

	listed := run(t, db, "project", "list")
	if !strings.Contains(listed.stdout, "API") {
		t.Fatalf("project was deleted despite refusal: %q", listed.stdout)
	}
}

func TestProjectDeleteToInboxPreservesTasks(t *testing.T) {
	db := newDB(t)
	repo := t.TempDir()
	run(t, db, "project", "add", "--repo", repo, "API")
	runIn(t, repo, db, "add", "--priority", "high", "keep me")

	got := run(t, db, "project", "delete", "--to-inbox", "--json", "API")
	if got.exit != 0 {
		t.Fatalf("exit = %d, stderr = %q", got.exit, got.stderr)
	}

	var result projectDeletionJSON
	if err := json.Unmarshal([]byte(got.stdout), &result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if result.Project.Name != "API" {
		t.Fatalf("project = %+v", result.Project)
	}
	if result.MovedTasks == nil || *result.MovedTasks != 1 {
		t.Fatalf("moved_tasks = %v, want 1", result.MovedTasks)
	}

	projects := run(t, db, "project", "list", "--visibility", "all")
	if strings.Contains(projects.stdout, "API") {
		t.Fatalf("project list = %q, want API gone", projects.stdout)
	}

	listed := run(t, db, "list", "--json")
	var tasks []taskJSON
	if err := json.Unmarshal([]byte(listed.stdout), &tasks); err != nil {
		t.Fatalf("decode tasks: %v", err)
	}
	if len(tasks) != 1 || tasks[0].Project != nil || tasks[0].Title != "keep me" || tasks[0].Priority != "high" {
		t.Fatalf("tasks = %+v, want the preserved task in the Inbox", tasks)
	}
}

func TestProjectDeleteToInboxWorksForEmptyProject(t *testing.T) {
	db := newDB(t)
	run(t, db, "project", "add", "Empty")

	got := run(t, db, "project", "delete", "--to-inbox", "Empty")
	if got.exit != 0 {
		t.Fatalf("exit = %d, stderr = %q", got.exit, got.stderr)
	}
	listed := run(t, db, "project", "list", "--visibility", "all")
	if strings.Contains(listed.stdout, "Empty") {
		t.Fatalf("project list = %q, want Empty gone", listed.stdout)
	}
}

func TestProjectDeleteForceRequiresInteractiveConfirmation(t *testing.T) {
	db := newDB(t)
	run(t, db, "project", "add", "API")
	run(t, db, "add", "--project", "API", "task")

	got := run(t, db, "project", "delete", "--force", "API")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
	if !strings.Contains(got.stderr, "interactive") {
		t.Fatalf("stderr = %q, want an interactive-confirmation message", got.stderr)
	}

	listed := run(t, db, "project", "list")
	if !strings.Contains(listed.stdout, "API") {
		t.Fatalf("project was deleted without confirmation: %q", listed.stdout)
	}
}

func TestProjectDeleteEmptyRequiresInteractiveConfirmation(t *testing.T) {
	db := newDB(t)
	run(t, db, "project", "add", "Empty")

	got := run(t, db, "project", "delete", "Empty")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
	if !strings.Contains(got.stderr, "interactive") {
		t.Fatalf("stderr = %q", got.stderr)
	}
}

func TestProjectDeleteConflictingFlags(t *testing.T) {
	db := newDB(t)
	run(t, db, "project", "add", "API")

	got := run(t, db, "project", "delete", "--to-inbox", "--force", "API")
	if got.exit != 2 {
		t.Fatalf("exit = %d, want 2", got.exit)
	}
}

func TestProjectDeleteUnknown(t *testing.T) {
	db := newDB(t)

	got := run(t, db, "project", "delete", "--to-inbox", "Nope")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
	if !strings.Contains(got.stderr, "not found") {
		t.Fatalf("stderr = %q", got.stderr)
	}
}
