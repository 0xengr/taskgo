package app

import (
	"context"
	"errors"
	"time"

	"github.com/0xengr/taskgo/internal/domain"
)

// App coordinates domain rules over the persistence port.
type App struct {
	store Store
	now   func() time.Time
}

func New(store Store) *App {
	return &App{
		store: store,
		now:   func() time.Time { return time.Now().UTC() },
	}
}

// AddTaskInput carries raw capture input from the CLI.
type AddTaskInput struct {
	Title       string
	Description string
	Priority    domain.Priority
}

// AddTask validates capture input and persists a new Task.
func (a *App) AddTask(ctx context.Context, in AddTaskInput) (domain.Task, error) {
	validated, err := domain.BuildTask(in.Title, in.Description, in.Priority)
	if err != nil {
		return domain.Task{}, err
	}
	return a.store.InsertTask(ctx, validated, a.now())
}

// ListTasks returns open, unarchived Tasks in canonical order.
func (a *App) ListTasks(ctx context.Context) ([]domain.Task, error) {
	return a.store.ListOpenActive(ctx)
}

// AddProjectInput carries raw Project creation input from the CLI.
type AddProjectInput struct {
	Name     string
	RepoPath string
	HasRepo  bool
}

// AddProject validates a Project and persists it.
func (a *App) AddProject(ctx context.Context, in AddProjectInput) (domain.Project, error) {
	name, key, err := domain.NormalizeProjectName(in.Name)
	if err != nil {
		return domain.Project{}, err
	}

	var repoPath *string
	if in.HasRepo {
		canonical, err := CanonicalRepoPath(in.RepoPath)
		if err != nil {
			return domain.Project{}, err
		}
		repoPath = &canonical
	}

	return a.store.InsertProject(ctx, domain.NewProject{
		Name:     name,
		NameKey:  key,
		RepoPath: repoPath,
	}, a.now())
}

// ListProjects returns Projects filtered by visibility, ordered by name.
func (a *App) ListProjects(ctx context.Context, visibility domain.Visibility) ([]domain.Project, error) {
	return a.store.ListProjects(ctx, visibility)
}

// RenameProject changes a Project's name without touching its Tasks.
func (a *App) RenameProject(ctx context.Context, current, next string) (domain.Project, error) {
	project, err := a.findProject(ctx, current)
	if err != nil {
		return domain.Project{}, err
	}

	name, key, err := domain.NormalizeProjectName(next)
	if err != nil {
		return domain.Project{}, err
	}
	return a.store.UpdateProjectName(ctx, project.ID, name, key)
}

// EditProject rebinds or clears a Project's repository path.
func (a *App) EditProject(ctx context.Context, current, repoPath string, repoSet, clearRepo bool) (domain.Project, error) {
	if repoSet && clearRepo {
		return domain.Project{}, domain.NewValidationError("--repo and --clear-repo cannot be used together")
	}
	if !repoSet && !clearRepo {
		return domain.Project{}, domain.NewValidationError("nothing to change")
	}

	project, err := a.findProject(ctx, current)
	if err != nil {
		return domain.Project{}, err
	}

	var target *string
	if repoSet {
		canonical, err := CanonicalRepoPath(repoPath)
		if err != nil {
			return domain.Project{}, err
		}
		target = &canonical
	}
	return a.store.UpdateProjectRepo(ctx, project.ID, target)
}

// ArchiveProject hides a Project without changing its Tasks. It is
// idempotent.
func (a *App) ArchiveProject(ctx context.Context, name string) (domain.Project, error) {
	project, err := a.findProject(ctx, name)
	if err != nil {
		return domain.Project{}, err
	}
	if project.Archived() {
		return project, nil
	}

	now := a.now()
	return a.store.UpdateProjectArchived(ctx, project.ID, &now)
}

// UnarchiveProject restores a Project's visibility. It is idempotent.
func (a *App) UnarchiveProject(ctx context.Context, name string) (domain.Project, error) {
	project, err := a.findProject(ctx, name)
	if err != nil {
		return domain.Project{}, err
	}
	if !project.Archived() {
		return project, nil
	}
	return a.store.UpdateProjectArchived(ctx, project.ID, nil)
}

func (a *App) findProject(ctx context.Context, name string) (domain.Project, error) {
	_, key, err := domain.NormalizeProjectName(name)
	if err != nil {
		return domain.Project{}, err
	}

	project, err := a.store.FindProjectByKey(ctx, key)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.Project{}, &domain.NotFoundError{Kind: "project", Ref: name}
	}
	if err != nil {
		return domain.Project{}, err
	}
	return project, nil
}
