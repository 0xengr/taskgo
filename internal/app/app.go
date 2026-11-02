package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
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
	Target      ProjectTarget
	ProjectName string
}

// AddTask validates capture input, resolves its Project, and persists it.
func (a *App) AddTask(ctx context.Context, in AddTaskInput) (domain.Task, error) {
	validated, err := domain.BuildTask(in.Title, in.Description, in.Priority)
	if err != nil {
		return domain.Task{}, err
	}

	project, err := a.resolveTarget(ctx, in)
	if err != nil {
		return domain.Task{}, err
	}
	if project != nil {
		validated.ProjectID = &project.ID
	}

	task, err := a.store.InsertTask(ctx, validated, a.now())
	if err != nil {
		return domain.Task{}, err
	}
	if project != nil {
		task.Project = &domain.ProjectRef{ID: project.ID, Name: project.Name}
	}
	return task, nil
}

func (a *App) resolveTarget(ctx context.Context, in AddTaskInput) (*domain.Project, error) {
	switch in.Target {
	case TargetInbox:
		return nil, nil

	case TargetProject:
		project, err := a.findProject(ctx, in.ProjectName)
		if err != nil {
			return nil, err
		}
		if project.Archived() {
			return nil, &domain.ConflictError{
				Message: fmt.Sprintf("project %q is archived; unarchive it first", project.Name),
			}
		}
		return &project, nil

	default:
		cwd, err := resolveCwd()
		if err != nil {
			return nil, fmt.Errorf("resolve current directory: %w", err)
		}
		projects, err := a.store.ListProjects(ctx, domain.VisibilityAll)
		if err != nil {
			return nil, err
		}
		project := DeepestProjectForPath(projects, cwd)
		if project == nil || project.Archived() {
			return nil, nil
		}
		return project, nil
	}
}

// ListTasks returns Tasks matching the query in canonical order.
func (a *App) ListTasks(ctx context.Context, query domain.TaskQuery) ([]domain.Task, error) {
	return a.store.ListTasks(ctx, query)
}

// ArchiveTask hides a Task without changing its completion state. It is
// idempotent.
func (a *App) ArchiveTask(ctx context.Context, id int64) (domain.Task, error) {
	task, err := a.GetTask(ctx, id)
	if err != nil {
		return domain.Task{}, err
	}
	if task.Archived() {
		return task, nil
	}

	now := a.now()
	return a.store.UpdateTaskArchived(ctx, id, &now)
}

// UnarchiveTask restores a Task's visibility. It is idempotent.
func (a *App) UnarchiveTask(ctx context.Context, id int64) (domain.Task, error) {
	task, err := a.GetTask(ctx, id)
	if err != nil {
		return domain.Task{}, err
	}
	if !task.Archived() {
		return task, nil
	}
	return a.store.UpdateTaskArchived(ctx, id, nil)
}

// GetTask returns one Task by ID, including archived Tasks.
func (a *App) GetTask(ctx context.Context, id int64) (domain.Task, error) {
	task, err := a.store.FindTaskByID(ctx, id)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.Task{}, &domain.NotFoundError{Kind: "task", Ref: "#" + strconv.FormatInt(id, 10)}
	}
	if err != nil {
		return domain.Task{}, err
	}
	return task, nil
}

// CompleteTask marks a Task Done. Repeating it preserves the original
// completion time.
func (a *App) CompleteTask(ctx context.Context, id int64) (domain.Task, error) {
	task, err := a.GetTask(ctx, id)
	if err != nil {
		return domain.Task{}, err
	}
	if !task.Open() {
		return task, nil
	}

	now := a.now()
	return a.store.UpdateTaskCompleted(ctx, id, &now)
}

// ReopenTask clears completion. Repeating it on an open Task is a no-op.
func (a *App) ReopenTask(ctx context.Context, id int64) (domain.Task, error) {
	task, err := a.GetTask(ctx, id)
	if err != nil {
		return domain.Task{}, err
	}
	if task.Open() {
		return task, nil
	}
	return a.store.UpdateTaskCompleted(ctx, id, nil)
}

// EditTaskInput describes the fields the CLI asked to change. The Set flags
// distinguish "leave alone" from "set to the zero value".
type EditTaskInput struct {
	ID             int64
	Title          string
	SetTitle       bool
	Description    string
	SetDescription bool
	Priority       domain.Priority
	SetPriority    bool
	ProjectName    string
	SetProject     bool
	Inbox          bool
}

// EditTask applies all supplied fields in one update.
func (a *App) EditTask(ctx context.Context, in EditTaskInput) (domain.Task, error) {
	if _, err := a.GetTask(ctx, in.ID); err != nil {
		return domain.Task{}, err
	}
	if in.SetProject && in.Inbox {
		return domain.Task{}, domain.NewValidationError("--project and --inbox cannot be used together")
	}

	update := domain.TaskUpdate{}
	changed := false

	if in.SetTitle {
		title, err := domain.NormalizeTitle(in.Title)
		if err != nil {
			return domain.Task{}, err
		}
		update.Title = title
		update.SetTitle = true
		changed = true
	}

	if in.SetDescription {
		description := strings.TrimSpace(in.Description)
		if description == "" {
			update.Description = nil
		} else {
			update.Description = &description
		}
		update.SetDescription = true
		changed = true
	}

	if in.SetPriority {
		if !in.Priority.Valid() {
			return domain.Task{}, domain.NewValidationError("priority must be one of low, medium, or high (got %q)", in.Priority)
		}
		update.Priority = in.Priority
		update.SetPriority = true
		changed = true
	}

	if in.SetProject {
		project, err := a.findProject(ctx, in.ProjectName)
		if err != nil {
			return domain.Task{}, err
		}
		if project.Archived() {
			return domain.Task{}, &domain.ConflictError{
				Message: fmt.Sprintf("project %q is archived; unarchive it first", project.Name),
			}
		}
		id := project.ID
		update.ProjectID = &id
		update.SetProject = true
		changed = true
	}

	if in.Inbox {
		update.ProjectID = nil
		update.SetProject = true
		changed = true
	}

	if !changed {
		return domain.Task{}, domain.NewValidationError("nothing to change")
	}
	return a.store.UpdateTask(ctx, in.ID, update)
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
