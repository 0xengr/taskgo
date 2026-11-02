package app

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/0xengr/taskgo/internal/domain"
)

func testProject(id int64, name, repo string, archived bool) domain.Project {
	project := domain.Project{ID: id, Name: name}
	if repo != "" {
		project.RepoPath = &repo
	}
	if archived {
		archivedAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		project.ArchivedAt = &archivedAt
	}
	return project
}

func TestDeepestProjectForPath(t *testing.T) {
	root := t.TempDir()
	api := filepath.Join(root, "api")
	nested := filepath.Join(api, "nested")
	api2 := filepath.Join(root, "api2")

	projects := []domain.Project{
		testProject(1, "API", api, false),
		testProject(2, "Nested", nested, false),
		testProject(3, "API2", api2, false),
		testProject(4, "NoRepo", "", false),
	}

	tests := []struct {
		path string
		want string
	}{
		{path: api, want: "API"},
		{path: filepath.Join(api, "src"), want: "API"},
		{path: nested, want: "Nested"},
		{path: filepath.Join(nested, "deep", "dir"), want: "Nested"},
		{path: api2, want: "API2"},
		{path: root, want: ""},
		{path: filepath.Join(root, "unrelated"), want: ""},
	}

	for _, test := range tests {
		got := DeepestProjectForPath(projects, test.path)
		if test.want == "" {
			if got != nil {
				t.Fatalf("path %q matched %q, want no match", test.path, got.Name)
			}
			continue
		}
		if got == nil {
			t.Fatalf("path %q matched nothing, want %q", test.path, test.want)
		}
		if got.Name != test.want {
			t.Fatalf("path %q matched %q, want %q", test.path, got.Name, test.want)
		}
	}
}

func TestDeepestProjectForPathIgnoresArchiveState(t *testing.T) {
	root := t.TempDir()
	outer := filepath.Join(root, "outer")
	inner := filepath.Join(outer, "inner")

	projects := []domain.Project{
		testProject(1, "Outer", outer, false),
		testProject(2, "Inner", inner, true),
	}

	got := DeepestProjectForPath(projects, inner)
	if got == nil || got.Name != "Inner" {
		t.Fatalf("matched %v, want the archived Inner project", got)
	}
}

func TestPathContains(t *testing.T) {
	root := t.TempDir()
	api := filepath.Join(root, "api")
	api2 := filepath.Join(root, "api2")

	tests := []struct {
		name   string
		parent string
		child  string
		want   bool
	}{
		{name: "same", parent: api, child: api, want: true},
		{name: "descendant", parent: api, child: filepath.Join(api, "pkg"), want: true},
		{name: "sibling prefix", parent: api, child: api2, want: false},
		{name: "parent", parent: api, child: root, want: false},
		{name: "dot dot escape", parent: api, child: filepath.Join(api, "..", "api2"), want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := pathContains(test.parent, test.child); got != test.want {
				t.Fatalf("pathContains(%q, %q) = %v, want %v", test.parent, test.child, got, test.want)
			}
		})
	}
}
