package app

import (
	"context"
	"time"

	"github.com/0xengr/taskgo/internal/domain"
)

// Store is the persistence port. Infrastructure implements it; the
// application layer depends only on this interface.
type Store interface {
	InsertTask(ctx context.Context, task domain.NewTask, createdAt time.Time) (domain.Task, error)
	ListOpenActive(ctx context.Context) ([]domain.Task, error)

	InsertProject(ctx context.Context, project domain.NewProject, createdAt time.Time) (domain.Project, error)
	FindProjectByKey(ctx context.Context, key string) (domain.Project, error)
	ListProjects(ctx context.Context, visibility domain.Visibility) ([]domain.Project, error)
	UpdateProjectName(ctx context.Context, id int64, name, nameKey string) (domain.Project, error)
	UpdateProjectRepo(ctx context.Context, id int64, repoPath *string) (domain.Project, error)
	UpdateProjectArchived(ctx context.Context, id int64, archivedAt *time.Time) (domain.Project, error)
}
