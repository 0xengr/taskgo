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

const projectColumns = `id, name, repo_path, created_at, archived_at`

func (d *DB) InsertProject(ctx context.Context, project domain.NewProject, createdAt time.Time) (domain.Project, error) {
	result, err := d.db.ExecContext(ctx,
		`INSERT INTO projects (name, name_key, repo_path, created_at) VALUES (?, ?, ?, ?)`,
		project.Name, project.NameKey, nullableString(project.RepoPath), formatTimestamp(createdAt),
	)
	if err != nil {
		if uniqueViolation(err, "projects.name_key") {
			return domain.Project{}, &domain.ConflictError{
				Message: fmt.Sprintf("a project named %q already exists", project.Name),
			}
		}
		if uniqueViolation(err, "projects.repo_path") {
			return domain.Project{}, &domain.ConflictError{
				Message: fmt.Sprintf("repository path is already bound to another project: %s", deref(project.RepoPath)),
			}
		}
		return domain.Project{}, fmt.Errorf("insert project: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return domain.Project{}, fmt.Errorf("read inserted project id: %w", err)
	}
	return domain.Project{
		ID:        id,
		Name:      project.Name,
		RepoPath:  project.RepoPath,
		CreatedAt: createdAt.UTC(),
	}, nil
}

func (d *DB) FindProjectByKey(ctx context.Context, key string) (domain.Project, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT `+projectColumns+` FROM projects WHERE name_key = ?`, key)

	project, err := scanProject(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Project{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Project{}, err
	}
	return project, nil
}

func (d *DB) ListProjects(ctx context.Context, visibility domain.Visibility) ([]domain.Project, error) {
	query := `SELECT ` + projectColumns + ` FROM projects`
	switch visibility {
	case domain.VisibilityActive:
		query += ` WHERE archived_at IS NULL`
	case domain.VisibilityArchived:
		query += ` WHERE archived_at IS NOT NULL`
	}
	query += ` ORDER BY name_key ASC, id ASC`

	rows, err := d.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	projects := make([]domain.Project, 0)
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	return projects, nil
}

func (d *DB) UpdateProjectName(ctx context.Context, id int64, name, nameKey string) (domain.Project, error) {
	if _, err := d.db.ExecContext(ctx,
		`UPDATE projects SET name = ?, name_key = ? WHERE id = ?`,
		name, nameKey, id,
	); err != nil {
		if uniqueViolation(err, "projects.name_key") {
			return domain.Project{}, &domain.ConflictError{
				Message: fmt.Sprintf("a project named %q already exists", name),
			}
		}
		return domain.Project{}, fmt.Errorf("rename project: %w", err)
	}
	return d.projectByID(ctx, id)
}

func (d *DB) UpdateProjectRepo(ctx context.Context, id int64, repoPath *string) (domain.Project, error) {
	if _, err := d.db.ExecContext(ctx,
		`UPDATE projects SET repo_path = ? WHERE id = ?`,
		nullableString(repoPath), id,
	); err != nil {
		if uniqueViolation(err, "projects.repo_path") {
			return domain.Project{}, &domain.ConflictError{
				Message: fmt.Sprintf("repository path is already bound to another project: %s", deref(repoPath)),
			}
		}
		return domain.Project{}, fmt.Errorf("edit project repository path: %w", err)
	}
	return d.projectByID(ctx, id)
}

func (d *DB) UpdateProjectArchived(ctx context.Context, id int64, archivedAt *time.Time) (domain.Project, error) {
	var value any
	if archivedAt != nil {
		value = formatTimestamp(*archivedAt)
	}
	if _, err := d.db.ExecContext(ctx,
		`UPDATE projects SET archived_at = ? WHERE id = ?`, value, id,
	); err != nil {
		return domain.Project{}, fmt.Errorf("update project archive state: %w", err)
	}
	return d.projectByID(ctx, id)
}

func (d *DB) projectByID(ctx context.Context, id int64) (domain.Project, error) {
	row := d.db.QueryRowContext(ctx,
		`SELECT `+projectColumns+` FROM projects WHERE id = ?`, id)
	return scanProject(row)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProject(row rowScanner) (domain.Project, error) {
	var (
		project         domain.Project
		repoPath        sql.NullString
		archived        sql.NullString
		createdAtString string
	)
	if err := row.Scan(&project.ID, &project.Name, &repoPath, &createdAtString, &archived); err != nil {
		return domain.Project{}, fmt.Errorf("scan project: %w", err)
	}

	createdAt, err := parseTimestamp(createdAtString)
	if err != nil {
		return domain.Project{}, err
	}
	project.CreatedAt = createdAt

	if repoPath.Valid {
		project.RepoPath = &repoPath.String
	}
	if archived.Valid {
		archivedAt, err := parseTimestamp(archived.String)
		if err != nil {
			return domain.Project{}, err
		}
		project.ArchivedAt = &archivedAt
	}
	return project, nil
}

func uniqueViolation(err error, column string) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: "+column)
}

func deref(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
