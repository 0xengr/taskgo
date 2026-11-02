package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/0xengr/taskgo/internal/domain"
)

// timestampLayout is fixed-width so text ordering matches chronological order.
const timestampLayout = "2006-01-02T15:04:05.000000000Z"

const taskSelect = `
	SELECT t.id, t.title, t.description, t.priority, t.created_at,
	       t.completed_at, t.archived_at, p.id, p.name
	FROM tasks AS t
	LEFT JOIN projects AS p ON p.id = t.project_id`

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
		`INSERT INTO tasks (title, description, priority, project_id, created_at)
		 VALUES (?, ?, ?, ?, ?)`,
		task.Title,
		nullableString(task.Description),
		string(task.Priority),
		nullableInt64(task.ProjectID),
		formatTimestamp(createdAt),
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

func (d *DB) ListTasks(ctx context.Context, filter domain.TaskQuery) ([]domain.Task, error) {
	filter = filter.Normalized()

	conditions := make([]string, 0, 3)
	switch filter.State {
	case domain.StateOpen:
		conditions = append(conditions, "t.completed_at IS NULL")
	case domain.StateDone:
		conditions = append(conditions, "t.completed_at IS NOT NULL")
	}
	switch filter.Visibility {
	case domain.VisibilityActive:
		conditions = append(conditions,
			"t.archived_at IS NULL",
			"(p.id IS NULL OR p.archived_at IS NULL)",
		)
	case domain.VisibilityArchived:
		conditions = append(conditions,
			"(t.archived_at IS NOT NULL OR p.archived_at IS NOT NULL)",
		)
	}

	query := taskSelect
	if len(conditions) > 0 {
		query += "\n\t\tWHERE " + strings.Join(conditions, "\n\t\t  AND ")
	}
	query += `
		ORDER BY CASE t.priority
		             WHEN 'high' THEN 0
		             WHEN 'medium' THEN 1
		             ELSE 2
		         END,
		         t.created_at ASC,
		         t.id ASC`

	rows, err := d.db.QueryContext(ctx, query)
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

func (d *DB) FindTaskByID(ctx context.Context, id int64) (domain.Task, error) {
	row := d.db.QueryRowContext(ctx, taskSelect+`
		WHERE t.id = ?`, id)

	task, err := scanTask(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Task{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Task{}, err
	}
	return task, nil
}

func (d *DB) UpdateTaskCompleted(ctx context.Context, id int64, completedAt *time.Time) (domain.Task, error) {
	var value any
	if completedAt != nil {
		value = formatTimestamp(*completedAt)
	}
	if _, err := d.db.ExecContext(ctx,
		`UPDATE tasks SET completed_at = ? WHERE id = ?`, value, id,
	); err != nil {
		return domain.Task{}, fmt.Errorf("update task completion: %w", err)
	}
	return d.FindTaskByID(ctx, id)
}

func (d *DB) UpdateTaskArchived(ctx context.Context, id int64, archivedAt *time.Time) (domain.Task, error) {
	var value any
	if archivedAt != nil {
		value = formatTimestamp(*archivedAt)
	}
	if _, err := d.db.ExecContext(ctx,
		`UPDATE tasks SET archived_at = ? WHERE id = ?`, value, id,
	); err != nil {
		return domain.Task{}, fmt.Errorf("update task archive state: %w", err)
	}
	return d.FindTaskByID(ctx, id)
}

func (d *DB) UpdateTask(ctx context.Context, id int64, update domain.TaskUpdate) (domain.Task, error) {
	sets := make([]string, 0, 4)
	args := make([]any, 0, 5)

	if update.SetTitle {
		sets = append(sets, "title = ?")
		args = append(args, update.Title)
	}
	if update.SetDescription {
		sets = append(sets, "description = ?")
		args = append(args, nullableString(update.Description))
	}
	if update.SetPriority {
		sets = append(sets, "priority = ?")
		args = append(args, string(update.Priority))
	}
	if update.SetProject {
		sets = append(sets, "project_id = ?")
		args = append(args, nullableInt64(update.ProjectID))
	}
	if len(sets) == 0 {
		return d.FindTaskByID(ctx, id)
	}

	args = append(args, id)
	if _, err := d.db.ExecContext(ctx,
		`UPDATE tasks SET `+strings.Join(sets, ", ")+` WHERE id = ?`, args...,
	); err != nil {
		return domain.Task{}, fmt.Errorf("update task: %w", err)
	}
	return d.FindTaskByID(ctx, id)
}

func scanTask(row rowScanner) (domain.Task, error) {
	var (
		task                             domain.Task
		description, completed, archived sql.NullString
		projectName                      sql.NullString
		projectID                        sql.NullInt64
		priority, created                string
	)
	if err := row.Scan(
		&task.ID,
		&task.Title,
		&description,
		&priority,
		&created,
		&completed,
		&archived,
		&projectID,
		&projectName,
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
	if projectID.Valid {
		task.Project = &domain.ProjectRef{ID: projectID.Int64, Name: projectName.String}
	}
	return task, nil
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableInt64(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}
