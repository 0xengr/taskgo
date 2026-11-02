package app

import (
	"context"
	"time"

	"github.com/0xengr/taskgo/internal/domain"
)

// TaskStore is the persistence port for Tasks. Infrastructure implements it;
// the application layer depends only on this interface.
type TaskStore interface {
	InsertTask(ctx context.Context, task domain.NewTask, createdAt time.Time) (domain.Task, error)
	ListOpenActive(ctx context.Context) ([]domain.Task, error)
}
