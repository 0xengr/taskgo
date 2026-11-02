package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type projectJSON struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	RepositoryPath *string `json:"repository_path"`
	CreatedAt      string  `json:"created_at"`
	ArchivedAt     *string `json:"archived_at"`
}

func TestProjectAddAndList(t *testing.T) {
	db := newDB(t)

	added := run(t, db, "project", "add", "API")
	if added.exit != 0 {
		t.Fatalf("add exit = %d, stderr = %q", added.exit, added.stderr)
	}
	if !strings.Contains(added.stdout, `Created project "API"`) {
		t.Fatalf("add stdout = %q", added.stdout)
	}

	listed := run(t, db, "project", "list")
	if listed.exit != 0 {
		t.Fatalf("list exit = %d, stderr = %q", listed.exit, listed.stderr)
	}
	if !strings.Contains(listed.stdout, "API") {
		t.Fatalf("list stdout = %q, want API", listed.stdout)
	}
}

func TestProjectAddJSONIncludesCanonicalRepository(t *testing.T) {
	db := newDB(t)
	repo := t.TempDir()
	resolved, err := filepath.EvalSymlinks(repo)
	if err != nil {
		t.Fatalf("resolve temp dir: %v", err)
	}

	got := run(t, db, "project", "add", "--json", "--repo", repo, "API")
	if got.exit != 0 {
		t.Fatalf("add exit = %d, stderr = %q", got.exit, got.stderr)
	}

	var project projectJSON
	if err := json.Unmarshal([]byte(got.stdout), &project); err != nil {
		t.Fatalf("decode JSON %q: %v", got.stdout, err)
	}
	if project.Name != "API" {
		t.Fatalf("name = %q", project.Name)
	}
	if project.RepositoryPath == nil || *project.RepositoryPath != resolved {
		t.Fatalf("repository_path = %v, want %q", project.RepositoryPath, resolved)
	}
	if project.ArchivedAt != nil {
		t.Fatalf("archived_at = %v, want null", project.ArchivedAt)
	}
	if _, err := time.Parse(time.RFC3339, project.CreatedAt); err != nil {
		t.Fatalf("created_at %q is not RFC3339: %v", project.CreatedAt, err)
	}
}

func TestProjectNameUniquenessIsCaseInsensitive(t *testing.T) {
	db := newDB(t)

	if got := run(t, db, "project", "add", "API"); got.exit != 0 {
		t.Fatalf("first add exit = %d, stderr = %q", got.exit, got.stderr)
	}
	got := run(t, db, "project", "add", "api")
	if got.exit != 1 {
		t.Fatalf("duplicate add exit = %d, want 1", got.exit)
	}
	if !strings.Contains(got.stderr, "already exists") {
		t.Fatalf("stderr = %q, want an already-exists message", got.stderr)
	}
}

func TestProjectRename(t *testing.T) {
	db := newDB(t)
	if got := run(t, db, "project", "add", "API"); got.exit != 0 {
		t.Fatalf("add exit = %d, stderr = %q", got.exit, got.stderr)
	}

	got := run(t, db, "project", "rename", "API", "Backend")
	if got.exit != 0 {
		t.Fatalf("rename exit = %d, stderr = %q", got.exit, got.stderr)
	}
	if !strings.Contains(got.stdout, "Backend") {
		t.Fatalf("rename stdout = %q", got.stdout)
	}

	listed := run(t, db, "project", "list")
	if strings.Contains(listed.stdout, "API") || !strings.Contains(listed.stdout, "Backend") {
		t.Fatalf("list stdout = %q, want the new name only", listed.stdout)
	}
}

func TestProjectRenameConflict(t *testing.T) {
	db := newDB(t)
	run(t, db, "project", "add", "API")
	run(t, db, "project", "add", "Web")

	got := run(t, db, "project", "rename", "API", "web")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
	if !strings.Contains(got.stderr, "already exists") {
		t.Fatalf("stderr = %q", got.stderr)
	}
}

func TestProjectEditRepository(t *testing.T) {
	db := newDB(t)
	run(t, db, "project", "add", "API")
	repo := t.TempDir()

	edited := run(t, db, "project", "edit", "API", "--repo", repo)
	if edited.exit != 0 {
		t.Fatalf("edit exit = %d, stderr = %q", edited.exit, edited.stderr)
	}
	if !strings.Contains(edited.stdout, repo) && !strings.Contains(edited.stdout, filepath.Base(repo)) {
		t.Fatalf("edit stdout = %q, want the repository", edited.stdout)
	}

	cleared := run(t, db, "project", "edit", "API", "--clear-repo")
	if cleared.exit != 0 {
		t.Fatalf("clear exit = %d, stderr = %q", cleared.exit, cleared.stderr)
	}
	if !strings.Contains(cleared.stdout, "cleared") {
		t.Fatalf("clear stdout = %q", cleared.stdout)
	}
}

func TestProjectEditRequiresExactlyOneChange(t *testing.T) {
	db := newDB(t)
	run(t, db, "project", "add", "API")

	if got := run(t, db, "project", "edit", "API"); got.exit != 2 {
		t.Fatalf("no flags exit = %d, want 2", got.exit)
	}
	if got := run(t, db, "project", "edit", "API", "--repo", t.TempDir(), "--clear-repo"); got.exit != 2 {
		t.Fatalf("both flags exit = %d, want 2", got.exit)
	}
}

func TestProjectRepositoryPathMustExist(t *testing.T) {
	db := newDB(t)

	got := run(t, db, "project", "add", "--repo", filepath.Join(t.TempDir(), "missing"), "API")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
	if !strings.Contains(got.stderr, "does not exist") {
		t.Fatalf("stderr = %q", got.stderr)
	}
}

func TestProjectRepositoryPathMustBeDirectory(t *testing.T) {
	db := newDB(t)
	path := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	got := run(t, db, "project", "add", "--repo", path, "API")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
	if !strings.Contains(got.stderr, "not a directory") {
		t.Fatalf("stderr = %q", got.stderr)
	}
}

func TestProjectRepositoryPathIsUnique(t *testing.T) {
	db := newDB(t)
	repo := t.TempDir()
	run(t, db, "project", "add", "--repo", repo, "API")

	got := run(t, db, "project", "add", "--repo", repo, "Web")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
	if !strings.Contains(got.stderr, "already bound") {
		t.Fatalf("stderr = %q", got.stderr)
	}
}

func TestProjectArchiveAndVisibility(t *testing.T) {
	db := newDB(t)
	run(t, db, "project", "add", "API")
	run(t, db, "project", "add", "Web")

	archived := run(t, db, "project", "archive", "API")
	if archived.exit != 0 {
		t.Fatalf("archive exit = %d, stderr = %q", archived.exit, archived.stderr)
	}

	active := run(t, db, "project", "list")
	if strings.Contains(active.stdout, "API") || !strings.Contains(active.stdout, "Web") {
		t.Fatalf("active list = %q, want only Web", active.stdout)
	}

	hidden := run(t, db, "project", "list", "--visibility", "archived")
	if !strings.Contains(hidden.stdout, "API") || strings.Contains(hidden.stdout, "Web") {
		t.Fatalf("archived list = %q, want only API", hidden.stdout)
	}
	if !strings.Contains(hidden.stdout, "(archived)") {
		t.Fatalf("archived list = %q, want an archived marker", hidden.stdout)
	}

	all := run(t, db, "project", "list", "--visibility", "all")
	if !strings.Contains(all.stdout, "API") || !strings.Contains(all.stdout, "Web") {
		t.Fatalf("all list = %q, want both projects", all.stdout)
	}

	restored := run(t, db, "project", "unarchive", "API")
	if restored.exit != 0 {
		t.Fatalf("unarchive exit = %d, stderr = %q", restored.exit, restored.stderr)
	}
	active = run(t, db, "project", "list")
	if !strings.Contains(active.stdout, "API") {
		t.Fatalf("active list after unarchive = %q", active.stdout)
	}
}

func TestProjectArchiveIsIdempotent(t *testing.T) {
	db := newDB(t)
	run(t, db, "project", "add", "API")

	first := run(t, db, "project", "archive", "--json", "API")
	if first.exit != 0 {
		t.Fatalf("first archive exit = %d, stderr = %q", first.exit, first.stderr)
	}
	var firstProject projectJSON
	if err := json.Unmarshal([]byte(first.stdout), &firstProject); err != nil {
		t.Fatalf("decode first: %v", err)
	}
	if firstProject.ArchivedAt == nil {
		t.Fatal("archived_at is null after archive")
	}

	second := run(t, db, "project", "archive", "--json", "API")
	if second.exit != 0 {
		t.Fatalf("second archive exit = %d, stderr = %q", second.exit, second.stderr)
	}
	var secondProject projectJSON
	if err := json.Unmarshal([]byte(second.stdout), &secondProject); err != nil {
		t.Fatalf("decode second: %v", err)
	}
	if secondProject.ArchivedAt == nil || *secondProject.ArchivedAt != *firstProject.ArchivedAt {
		t.Fatalf("archived_at changed: %v -> %v", firstProject.ArchivedAt, secondProject.ArchivedAt)
	}
}

func TestProjectUnknownName(t *testing.T) {
	db := newDB(t)

	got := run(t, db, "project", "archive", "Nope")
	if got.exit != 1 {
		t.Fatalf("exit = %d, want 1", got.exit)
	}
	if !strings.Contains(got.stderr, "not found") {
		t.Fatalf("stderr = %q", got.stderr)
	}
}

func TestProjectInvalidVisibility(t *testing.T) {
	db := newDB(t)

	got := run(t, db, "project", "list", "--visibility", "hidden")
	if got.exit != 2 {
		t.Fatalf("exit = %d, want 2", got.exit)
	}
}

func TestProjectListSortsByName(t *testing.T) {
	db := newDB(t)
	run(t, db, "project", "add", "Zebra")
	run(t, db, "project", "add", "apple")

	listed := run(t, db, "project", "list", "--json")
	var projects []projectJSON
	if err := json.Unmarshal([]byte(listed.stdout), &projects); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(projects) != 2 || projects[0].Name != "apple" || projects[1].Name != "Zebra" {
		t.Fatalf("projects = %+v, want apple then Zebra", projects)
	}
}
