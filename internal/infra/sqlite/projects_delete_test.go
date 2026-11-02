package sqlite

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/0xengr/taskgo/internal/domain"
)

func seedProjectWithTask(t *testing.T, ctx context.Context, db *DB) (domain.Project, domain.Task) {
	t.Helper()

	project, err := db.InsertProject(ctx, domain.NewProject{
		Name:    "API",
		NameKey: "api",
	}, time.Now().UTC())
	if err != nil {
		t.Fatalf("insert project: %v", err)
	}

	task, err := db.InsertTask(ctx, domain.NewTask{
		Title:     "project task",
		Priority:  domain.PriorityMedium,
		ProjectID: &project.ID,
	}, time.Now().UTC())
	if err != nil {
		t.Fatalf("insert task: %v", err)
	}
	return project, task
}

func TestDeleteProjectToInbox(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "taskgo.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	project, task := seedProjectWithTask(t, ctx, db)

	moved, err := db.DeleteProjectToInbox(ctx, project.ID)
	if err != nil {
		t.Fatalf("DeleteProjectToInbox: %v", err)
	}
	if moved != 1 {
		t.Fatalf("moved = %d, want 1", moved)
	}

	if _, err := db.FindProjectByKey(ctx, "api"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("project still exists (err = %v)", err)
	}

	kept, err := db.FindTaskByID(ctx, task.ID)
	if err != nil {
		t.Fatalf("task was not preserved: %v", err)
	}
	if kept.Project != nil {
		t.Fatalf("task project = %+v, want nil", kept.Project)
	}
	if kept.Title != "project task" {
		t.Fatalf("task title = %q", kept.Title)
	}
}

func TestDeleteProjectWithTasks(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(t.TempDir(), "taskgo.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer db.Close()

	project, task := seedProjectWithTask(t, ctx, db)

	if err := db.DeleteProjectWithTasks(ctx, project.ID); err != nil {
		t.Fatalf("DeleteProjectWithTasks: %v", err)
	}

	if _, err := db.FindProjectByKey(ctx, "api"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("project still exists (err = %v)", err)
	}
	if _, err := db.FindTaskByID(ctx, task.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("task still exists (err = %v)", err)
	}

	next, err := db.InsertProject(ctx, domain.NewProject{Name: "Next", NameKey: "next"}, time.Now().UTC())
	if err != nil {
		t.Fatalf("insert next project: %v", err)
	}
	if next.ID <= project.ID {
		t.Fatalf("project id reused: %d after %d", next.ID, project.ID)
	}
}
