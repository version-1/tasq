package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func (s *Store) CreateProject(ctx context.Context, input entity.CreateProjectInput) (entity.Project, error) {
	normalized, err := entity.NormalizeCreateProject(input)
	if err != nil {
		return entity.Project{}, err
	}
	now := nowString()
	result, err := s.db.ExecContext(ctx, `INSERT INTO projects (
		key, name, description, location, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?)`,
		normalized.Key,
		normalized.Name,
		normalized.Description,
		normalized.Location,
		now,
		now,
	)
	if err != nil {
		return entity.Project{}, fmt.Errorf("create project: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return entity.Project{}, fmt.Errorf("read created project id: %w", err)
	}
	return s.Project(ctx, id)
}

func (s *Store) Projects(ctx context.Context) ([]entity.Project, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+projectColumns()+` FROM projects ORDER BY updated_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	projects := []entity.Project{}
	for rows.Next() {
		item, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}
	return projects, nil
}

func (s *Store) Project(ctx context.Context, id int64) (entity.Project, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+projectColumns()+` FROM projects WHERE id = ?`, id)
	item, err := scanProject(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Project{}, sql.ErrNoRows
		}
		return entity.Project{}, fmt.Errorf("read project: %w", err)
	}
	return item, nil
}

func (s *Store) ProjectByKey(ctx context.Context, key string) (entity.Project, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+projectColumns()+` FROM projects WHERE key = ?`, key)
	item, err := scanProject(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Project{}, sql.ErrNoRows
		}
		return entity.Project{}, fmt.Errorf("read project by key: %w", err)
	}
	return item, nil
}

func projectExistsTx(ctx context.Context, tx *sql.Tx, id int64) error {
	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT 1 FROM projects WHERE id = ?`, id).Scan(&exists); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sql.ErrNoRows
		}
		return fmt.Errorf("read project: %w", err)
	}
	return nil
}

func (s *Store) UpdateProject(ctx context.Context, id int64, input entity.UpdateProjectInput) (entity.Project, error) {
	normalized, err := entity.NormalizeUpdateProject(input)
	if err != nil {
		return entity.Project{}, err
	}
	current, err := s.Project(ctx, id)
	if err != nil {
		return entity.Project{}, err
	}
	if normalized.Key != nil {
		current.Key = *normalized.Key
	}
	if normalized.Name != nil {
		current.Name = *normalized.Name
	}
	if normalized.Description != nil {
		current.Description = *normalized.Description
	}
	if normalized.Location != nil {
		current.Location = *normalized.Location
	}
	_, err = s.db.ExecContext(ctx, `UPDATE projects SET
		key = ?, name = ?, description = ?, location = ?, updated_at = ?
		WHERE id = ?`,
		current.Key,
		current.Name,
		current.Description,
		current.Location,
		nowString(),
		id,
	)
	if err != nil {
		return entity.Project{}, fmt.Errorf("update project: %w", err)
	}
	return s.Project(ctx, id)
}

func (s *Store) DeleteProject(ctx context.Context, id int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	attachments, err := projectAttachments(ctx, tx, id)
	if err != nil {
		return err
	}
	var storage *AttachmentStorage
	if len(attachments) > 0 {
		storage, err = s.resolveAttachmentStorage()
		if err != nil {
			return fmt.Errorf("resolve attachment storage: %w", err)
		}
		for _, attachment := range attachments {
			if _, err := storage.Resolve(attachment.Path); err != nil {
				return fmt.Errorf("resolve attachment file %s: %w", attachment.ID, err)
			}
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM attachments
		WHERE entity_type = ? AND entity_id IN (
			SELECT CAST(id AS TEXT) FROM issues WHERE project_id = ?
		)`, entity.AttachmentEntityIssue, id); err != nil {
		return fmt.Errorf("delete issue attachments: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM attachments
		WHERE entity_type = ? AND entity_id IN (
			SELECT CAST(comments.id AS TEXT)
			FROM comments
			JOIN issues ON issues.id = comments.issue_id
			WHERE issues.project_id = ?
		)`, entity.AttachmentEntityComment, id); err != nil {
		return fmt.Errorf("delete comment attachments: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM issue_dependencies
		WHERE parent_issue_id IN (SELECT id FROM issues WHERE project_id = ?)
		   OR dependency_issue_id IN (SELECT id FROM issues WHERE project_id = ?)`, id, id); err != nil {
		return fmt.Errorf("delete issue dependencies: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM change_requests
		WHERE issue_id IN (SELECT id FROM issues WHERE project_id = ?)`, id); err != nil {
		return fmt.Errorf("delete change requests: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM comments
		WHERE issue_id IN (SELECT id FROM issues WHERE project_id = ?)`, id); err != nil {
		return fmt.Errorf("delete comments: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM issues WHERE project_id = ?`, id); err != nil {
		return fmt.Errorf("delete issues: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM project_workflows WHERE project_id = ?`, id); err != nil {
		return fmt.Errorf("delete project workflow: %w", err)
	}
	result, err := tx.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	if storage != nil {
		for _, attachment := range attachments {
			if err := storage.Delete(attachment.Path); err != nil {
				return fmt.Errorf("delete attachment file %s: %w", attachment.ID, err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	tx = nil
	return nil
}

func (s *Store) ProjectIssueIDs(ctx context.Context, projectID int64) ([]int64, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM issues WHERE project_id = ? ORDER BY id ASC`, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project issue ids: %w", err)
	}
	defer rows.Close()

	var issueIDs []int64
	for rows.Next() {
		var issueID int64
		if err := rows.Scan(&issueID); err != nil {
			return nil, err
		}
		issueIDs = append(issueIDs, issueID)
	}
	return issueIDs, rows.Err()
}

func (s *Store) resolveAttachmentStorage() (*AttachmentStorage, error) {
	if s.attachmentStorage != nil {
		return s.attachmentStorage, nil
	}
	return NewAttachmentStorageFromHome()
}

func projectAttachments(ctx context.Context, tx *sql.Tx, projectID int64) ([]entity.Attachment, error) {
	rows, err := tx.QueryContext(ctx, `SELECT `+attachmentColumns()+` FROM attachments
		WHERE entity_type = ? AND entity_id IN (
			SELECT CAST(id AS TEXT) FROM issues WHERE project_id = ?
		)
		UNION ALL
		SELECT `+attachmentColumns()+` FROM attachments
		WHERE entity_type = ? AND entity_id IN (
			SELECT CAST(comments.id AS TEXT)
			FROM comments
			JOIN issues ON issues.id = comments.issue_id
			WHERE issues.project_id = ?
		)
		ORDER BY created_at ASC, id ASC`, entity.AttachmentEntityIssue, projectID, entity.AttachmentEntityComment, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project attachments: %w", err)
	}
	defer rows.Close()

	var attachments []entity.Attachment
	for rows.Next() {
		attachment, err := scanAttachment(rows)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, attachment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate project attachments: %w", err)
	}
	return attachments, nil
}

func (s *Store) UpsertProjectWorkflow(ctx context.Context, input entity.UpsertProjectWorkflowInput) (entity.ProjectWorkflow, error) {
	normalized, err := entity.NormalizeUpsertProjectWorkflow(input)
	if err != nil {
		return entity.ProjectWorkflow{}, err
	}
	current, err := s.ProjectWorkflow(ctx, normalized.ProjectID)
	if err == nil && current.Checksum == normalized.Checksum {
		return current, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return entity.ProjectWorkflow{}, err
	}
	if _, err := s.Project(ctx, normalized.ProjectID); err != nil {
		return entity.ProjectWorkflow{}, err
	}
	now := nowString()
	_, err = s.db.ExecContext(ctx, `INSERT INTO project_workflows (
		project_id, frontmatter_json, body, checksum, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?)
	ON CONFLICT(project_id) DO UPDATE SET
		frontmatter_json = excluded.frontmatter_json,
		body = excluded.body,
		checksum = excluded.checksum,
		updated_at = excluded.updated_at`,
		normalized.ProjectID,
		normalized.FrontmatterJSON,
		normalized.Body,
		normalized.Checksum,
		now,
		now,
	)
	if err != nil {
		return entity.ProjectWorkflow{}, fmt.Errorf("upsert project workflow: %w", err)
	}
	return s.ProjectWorkflow(ctx, normalized.ProjectID)
}

func (s *Store) ProjectWorkflow(ctx context.Context, projectID int64) (entity.ProjectWorkflow, error) {
	if projectID <= 0 {
		return entity.ProjectWorkflow{}, errors.New("projectId is required")
	}
	row := s.db.QueryRowContext(ctx, `SELECT `+projectWorkflowColumns()+` FROM project_workflows WHERE project_id = ?`, projectID)
	item, err := scanProjectWorkflow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ProjectWorkflow{}, sql.ErrNoRows
		}
		return entity.ProjectWorkflow{}, fmt.Errorf("read project workflow: %w", err)
	}
	return item, nil
}

func (s *Store) DeleteProjectWorkflow(ctx context.Context, projectID int64) error {
	if projectID <= 0 {
		return errors.New("projectId is required")
	}
	if _, err := s.Project(ctx, projectID); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM project_workflows WHERE project_id = ?`, projectID); err != nil {
		return fmt.Errorf("delete project workflow: %w", err)
	}
	return nil
}
