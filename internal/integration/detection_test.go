package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
}

func TestCaptureAssignsDetectedProject(t *testing.T) {
	db := newDB(t)
	repo := t.TempDir()
	if got := run(t, db, "project", "add", "--repo", repo, "API"); got.exit != 0 {
		t.Fatalf("project add exit = %d, stderr = %q", got.exit, got.stderr)
	}

	added := runIn(t, repo, db, "add", "detected")
	if added.exit != 0 {
		t.Fatalf("add exit = %d, stderr = %q", added.exit, added.stderr)
	}
	if !strings.Contains(added.stdout, "Added #1 to API") {
		t.Fatalf("add stdout = %q, want the detected destination", added.stdout)
	}

	listed := run(t, db, "list", "--json")
	var tasks []taskJSON
	if err := json.Unmarshal([]byte(listed.stdout), &tasks); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(tasks) != 1 || tasks[0].Project == nil || tasks[0].Project.Name != "API" {
		t.Fatalf("tasks = %+v, want one task assigned to API", tasks)
	}
}

func TestCaptureUsesDeepestNestedProject(t *testing.T) {
	db := newDB(t)
	outer := t.TempDir()
	inner := filepath.Join(outer, "inner")
	deep := filepath.Join(inner, "pkg", "sub")
	mustMkdir(t, deep)

	run(t, db, "project", "add", "--repo", outer, "Outer")
	run(t, db, "project", "add", "--repo", inner, "Inner")

	nested := runIn(t, deep, db, "add", "nested work")
	if !strings.Contains(nested.stdout, "to Inner") {
		t.Fatalf("nested add stdout = %q, want Inner", nested.stdout)
	}

	atOuter := runIn(t, outer, db, "add", "outer work")
	if !strings.Contains(atOuter.stdout, "to Outer") {
		t.Fatalf("outer add stdout = %q, want Outer", atOuter.stdout)
	}
}

func TestArchivedProjectBlocksParentFallback(t *testing.T) {
	db := newDB(t)
	outer := t.TempDir()
	inner := filepath.Join(outer, "inner")
	mustMkdir(t, inner)

	run(t, db, "project", "add", "--repo", outer, "Outer")
	run(t, db, "project", "add", "--repo", inner, "Inner")
	run(t, db, "project", "archive", "Inner")

	blocked := runIn(t, inner, db, "add", "should not inherit Outer")
	if !strings.Contains(blocked.stdout, "to Inbox") {
		t.Fatalf("add stdout = %q, want Inbox at the archived barrier", blocked.stdout)
	}

	explicit := runIn(t, inner, db, "add", "--project", "Inner", "no")
	if explicit.exit != 1 {
		t.Fatalf("explicit archived project exit = %d, want 1", explicit.exit)
	}
	if !strings.Contains(explicit.stderr, "archived") {
		t.Fatalf("stderr = %q, want an archived message", explicit.stderr)
	}

	atOuter := runIn(t, outer, db, "add", "outer still works")
	if !strings.Contains(atOuter.stdout, "to Outer") {
		t.Fatalf("outer add stdout = %q, want Outer", atOuter.stdout)
	}
}

func TestInboxOverridesDetection(t *testing.T) {
	db := newDB(t)
	repo := t.TempDir()
	run(t, db, "project", "add", "--repo", repo, "API")

	got := runIn(t, repo, db, "add", "--inbox", "personal errand")
	if got.exit != 0 {
		t.Fatalf("add exit = %d, stderr = %q", got.exit, got.stderr)
	}
	if !strings.Contains(got.stdout, "to Inbox") {
		t.Fatalf("stdout = %q, want Inbox", got.stdout)
	}
}

func TestProjectFlagOverridesDetection(t *testing.T) {
	db := newDB(t)
	run(t, db, "project", "add", "API")

	got := run(t, db, "add", "--project", "api", "explicit")
	if got.exit != 0 {
		t.Fatalf("add exit = %d, stderr = %q", got.exit, got.stderr)
	}
	if !strings.Contains(got.stdout, "to API") {
		t.Fatalf("stdout = %q, want API", got.stdout)
	}
}

func TestProjectAndInboxConflict(t *testing.T) {
	db := newDB(t)
	run(t, db, "project", "add", "API")

	got := run(t, db, "add", "--project", "API", "--inbox", "nope")
	if got.exit != 2 {
		t.Fatalf("exit = %d, want 2", got.exit)
	}
}

func TestUnknownProjectOnCapture(t *testing.T) {
	db := newDB(t)

	got := run(t, db, "add", "--project", "Nope", "nope")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
	if !strings.Contains(got.stderr, "not found") {
		t.Fatalf("stderr = %q", got.stderr)
	}
}

func TestSymlinkedWorkingDirectoryMatches(t *testing.T) {
	db := newDB(t)
	repo := t.TempDir()
	run(t, db, "project", "add", "--repo", repo, "API")

	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(repo, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	got := runIn(t, link, db, "add", "through the looking glass")
	if !strings.Contains(got.stdout, "to API") {
		t.Fatalf("stdout = %q, want API through the symlink", got.stdout)
	}
}

func TestMovedRepositoryStopsMatching(t *testing.T) {
	db := newDB(t)
	parent := t.TempDir()
	repo := filepath.Join(parent, "repo")
	mustMkdir(t, repo)
	run(t, db, "project", "add", "--repo", repo, "API")

	moved := filepath.Join(parent, "moved")
	if err := os.Rename(repo, moved); err != nil {
		t.Fatalf("rename repo: %v", err)
	}

	got := runIn(t, moved, db, "add", "after the move")
	if !strings.Contains(got.stdout, "to Inbox") {
		t.Fatalf("stdout = %q, want Inbox after the repository moved", got.stdout)
	}
}

func TestSiblingPathPrefixDoesNotMatch(t *testing.T) {
	db := newDB(t)
	root := t.TempDir()
	api := filepath.Join(root, "api")
	api2 := filepath.Join(root, "api2")
	mustMkdir(t, api)
	mustMkdir(t, api2)
	run(t, db, "project", "add", "--repo", api, "API")

	got := runIn(t, api2, db, "add", "not in api")
	if !strings.Contains(got.stdout, "to Inbox") {
		t.Fatalf("stdout = %q, want Inbox for the sibling path", got.stdout)
	}
}

func TestCaptureJSONIncludesProject(t *testing.T) {
	db := newDB(t)
	repo := t.TempDir()
	run(t, db, "project", "add", "--repo", repo, "API")

	got := runIn(t, repo, db, "add", "--json", "json task")
	var task taskJSON
	if err := json.Unmarshal([]byte(got.stdout), &task); err != nil {
		t.Fatalf("decode JSON %q: %v", got.stdout, err)
	}
	if task.Project == nil || task.Project.ID == 0 || task.Project.Name != "API" {
		t.Fatalf("project = %+v, want the API reference", task.Project)
	}
}

func TestArchivedProjectHidesItsTasks(t *testing.T) {
	db := newDB(t)
	repo := t.TempDir()
	run(t, db, "project", "add", "--repo", repo, "API")
	runIn(t, repo, db, "add", "hidden task")

	before := run(t, db, "list")
	if !strings.Contains(before.stdout, "hidden task") {
		t.Fatalf("list before archive = %q, want the task", before.stdout)
	}

	run(t, db, "project", "archive", "API")

	after := run(t, db, "list")
	if strings.Contains(after.stdout, "hidden task") {
		t.Fatalf("list after archive = %q, want the task hidden", after.stdout)
	}
}
