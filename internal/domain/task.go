package domain

import (
	"strings"
	"time"
)

// ProjectRef is the identity a Task carries about its Project.
type ProjectRef struct {
	ID   int64
	Name string
}

// Task is a single unit of work. It is either Open or Done, and separately
// either active or Archived.
type Task struct {
	ID          int64
	Title       string
	Description *string
	Priority    Priority
	Project     *ProjectRef
	CreatedAt   time.Time
	CompletedAt *time.Time
	ArchivedAt  *time.Time
}

func (t Task) Open() bool     { return t.CompletedAt == nil }
func (t Task) Archived() bool { return t.ArchivedAt != nil }

// NewTask is a validated Task that has not been persisted yet.
type NewTask struct {
	Title       string
	Description *string
	Priority    Priority
	ProjectID   *int64
}

// TaskUpdate describes which Task fields to change. Unset fields are left
// untouched.
type TaskUpdate struct {
	Title          string
	SetTitle       bool
	Description    *string
	SetDescription bool
	Priority       Priority
	SetPriority    bool
	ProjectID      *int64
	SetProject     bool
}

// BuildTask validates capture input into a NewTask.
func BuildTask(title, description string, priority Priority) (NewTask, error) {
	normalized, err := NormalizeTitle(title)
	if err != nil {
		return NewTask{}, err
	}

	if priority == "" {
		priority = PriorityMedium
	}
	if !priority.Valid() {
		return NewTask{}, NewValidationError("priority must be one of low, medium, or high (got %q)", priority)
	}

	description = strings.TrimSpace(description)
	var desc *string
	if description != "" {
		desc = &description
	}

	return NewTask{Title: normalized, Description: desc, Priority: priority}, nil
}
