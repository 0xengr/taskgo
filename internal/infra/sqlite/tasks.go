package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/0xengr/taskgo/internal/domain"
)

// timestampLayout is fixed-width so text ordering matches chronological order.
const timestampLayout = "2006-01-02T15:04:05.000000000Z"

func formatTimestamp(t time.Time) string {
	return t.UTC().Format(timestampLayout)
}

func parseTimestamp(raw string) (time.Time, error) {
	t, err := time.Parse(timestampLayout, raw)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse timestamp %q: %w", raw, err)
	}
	return t, nil
}

func (d *DB) InsertTask(ctx context.Context, task domain.NewTask, createdAt time.Time) (domain.Task, error) {
	result, err := d.db.ExecContext(ctx,
		`INSERT INTO tasks (title, description, priority, created_at) VALUES (?, ?, ?, ?)`,
		task.Title, nullableString(task.Description), string(task.Priority), formatTimestamp(createdAt),
	)
	if err != nil {
		return domain.Task{}, fmt.Errorf("insert task: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return domain.Task{}, fmt.Errorf("read inserted task id: %w", err)
	}
	return domain.Task{
		ID:          id,
		Title:       task.Title,
		Description: task.Description,
		Priority:    task.Priority,
		CreatedAt:   createdAt.UTC(),
	}, nil
}

func (d *DB) ListOpenActive(ctx context.Context) ([]domain.Task, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT id, title, description, priority, created_at, completed_at, archived_at
		FROM tasks
		WHERE completed_at IS NULL AND archived_at IS NULL
		ORDER BY CASE priority
		             WHEN 'high' THEN 0
		             WHEN 'medium' THEN 1
		             ELSE 2
		         END,
		         created_at ASC,
		         id ASC`)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	tasks := make([]domain.Task, 0)
	for rows.Next() {
		task, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	return tasks, nil
}

func scanTask(rows *sql.Rows) (domain.Task, error) {
	var (
		task                             domain.Task
		description, completed, archived sql.NullString
		priority, created                string
	)
	if err := rows.Scan(
		&task.ID,
		&task.Title,
		&description,
		&priority,
		&created,
		&completed,
		&archived,
	); err != nil {
		return domain.Task{}, fmt.Errorf("scan task: %w", err)
	}

	task.Priority = domain.Priority(priority)
	createdAt, err := parseTimestamp(created)
	if err != nil {
		return domain.Task{}, err
	}
	task.CreatedAt = createdAt
	if description.Valid {
		task.Description = &description.String
	}
	if completed.Valid {
		completedAt, err := parseTimestamp(completed.String)
		if err != nil {
			return domain.Task{}, err
		}
		task.CompletedAt = &completedAt
	}
	if archived.Valid {
		archivedAt, err := parseTimestamp(archived.String)
		if err != nil {
			return domain.Task{}, err
		}
		task.ArchivedAt = &archivedAt
	}
	return task, nil
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}
