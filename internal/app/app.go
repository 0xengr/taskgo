package app

import (
	"context"
	"time"

	"github.com/0xengr/taskgo/internal/domain"
)

// App coordinates domain rules over the persistence port.
type App struct {
	store TaskStore
	now   func() time.Time
}

func New(store TaskStore) *App {
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
