package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/version-1/tasq/db/schema"
	"github.com/version-1/tasq/internal/issue/domain/entity"
)

type Store struct {
	db *sql.DB
}

func Open(ctx context.Context, path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)
	store := &Store{db: db}
	if err := store.migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate(ctx context.Context) error {
	if _, err := s.db.ExecContext(ctx, schema.IssueTracker); err != nil {
		return fmt.Errorf("migrate issue tracker sqlite: %w", err)
	}
	return nil
}

func (s *Store) CreateProject(ctx context.Context, input entity.CreateProjectInput) (entity.Project, error) {
	normalized, err := entity.NormalizeCreateProject(input)
	if err != nil {
		return entity.Project{}, err
	}
	now := nowString()
	result, err := s.db.ExecContext(ctx, `INSERT INTO projects (
		key, name, description, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?)`,
		normalized.Key,
		normalized.Name,
		normalized.Description,
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

func (s *Store) UpdateProject(ctx context.Context, id int64, input entity.UpdateProjectInput) (entity.Project, error) {
	current, err := s.Project(ctx, id)
	if err != nil {
		return entity.Project{}, err
	}
	if input.Key != nil {
		if *input.Key == "" {
			return entity.Project{}, errors.New("key is required")
		}
		current.Key = *input.Key
	}
	if input.Name != nil {
		if *input.Name == "" {
			return entity.Project{}, errors.New("name is required")
		}
		current.Name = *input.Name
	}
	if input.Description != nil {
		current.Description = *input.Description
	}
	_, err = s.db.ExecContext(ctx, `UPDATE projects SET
		key = ?, name = ?, description = ?, updated_at = ?
		WHERE id = ?`,
		current.Key,
		current.Name,
		current.Description,
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
	if _, err := tx.ExecContext(ctx, `DELETE FROM workspaces WHERE project_id = ?`, id); err != nil {
		return fmt.Errorf("delete project workspaces: %w", err)
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
	if err := tx.Commit(); err != nil {
		return err
	}
	tx = nil
	return nil
}

func (s *Store) CreateWorkspace(ctx context.Context, input entity.CreateWorkspaceInput) (entity.Workspace, error) {
	normalized, err := entity.NormalizeCreateWorkspace(input)
	if err != nil {
		return entity.Workspace{}, err
	}
	if _, err := s.Project(ctx, normalized.ProjectID); err != nil {
		return entity.Workspace{}, err
	}
	now := nowString()
	result, err := s.db.ExecContext(ctx, `INSERT INTO workspaces (
		project_id, name, path, status, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?)`,
		normalized.ProjectID,
		normalized.Name,
		normalized.Path,
		normalized.Status,
		now,
		now,
	)
	if err != nil {
		return entity.Workspace{}, fmt.Errorf("create workspace: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return entity.Workspace{}, fmt.Errorf("read created workspace id: %w", err)
	}
	return s.Workspace(ctx, id)
}

func (s *Store) Workspaces(ctx context.Context) ([]entity.Workspace, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+workspaceColumns()+` FROM workspaces ORDER BY updated_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []entity.Workspace
	for rows.Next() {
		item, err := scanWorkspace(rows)
		if err != nil {
			return nil, err
		}
		workspaces = append(workspaces, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate workspaces: %w", err)
	}
	return workspaces, nil
}

func (s *Store) Workspace(ctx context.Context, id int64) (entity.Workspace, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+workspaceColumns()+` FROM workspaces WHERE id = ?`, id)
	item, err := scanWorkspace(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Workspace{}, sql.ErrNoRows
		}
		return entity.Workspace{}, fmt.Errorf("read workspace: %w", err)
	}
	return item, nil
}

func (s *Store) UpdateWorkspace(ctx context.Context, id int64, input entity.UpdateWorkspaceInput) (entity.Workspace, error) {
	current, err := s.Workspace(ctx, id)
	if err != nil {
		return entity.Workspace{}, err
	}
	if input.ProjectID != nil {
		if *input.ProjectID <= 0 {
			return entity.Workspace{}, errors.New("projectId is required")
		}
		if _, err := s.Project(ctx, *input.ProjectID); err != nil {
			return entity.Workspace{}, err
		}
		current.ProjectID = *input.ProjectID
	}
	if input.Name != nil {
		if *input.Name == "" {
			return entity.Workspace{}, errors.New("name is required")
		}
		current.Name = *input.Name
	}
	if input.Path != nil {
		if *input.Path == "" {
			return entity.Workspace{}, errors.New("path is required")
		}
		current.Path = *input.Path
	}
	if input.Status != nil {
		if !entity.IsValidWorkspaceStatus(*input.Status) {
			return entity.Workspace{}, errors.New("status is invalid")
		}
		current.Status = *input.Status
	}
	_, err = s.db.ExecContext(ctx, `UPDATE workspaces SET
		project_id = ?, name = ?, path = ?, status = ?, updated_at = ?
		WHERE id = ?`,
		current.ProjectID,
		current.Name,
		current.Path,
		current.Status,
		nowString(),
		id,
	)
	if err != nil {
		return entity.Workspace{}, fmt.Errorf("update workspace: %w", err)
	}
	return s.Workspace(ctx, id)
}

func (s *Store) DeleteWorkspace(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM workspaces WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete workspace: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) CreateIssue(ctx context.Context, input entity.CreateIssueInput) (entity.Issue, error) {
	normalized, err := entity.NormalizeCreate(input)
	if err != nil {
		return entity.Issue{}, err
	}
	now := nowString()
	result, err := s.db.ExecContext(ctx, `INSERT INTO issues (
		title, description, status, priority, assignee, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		normalized.Title,
		normalized.Description,
		normalized.Status,
		normalized.Priority,
		normalized.Assignee,
		now,
		now,
	)
	if err != nil {
		return entity.Issue{}, fmt.Errorf("create issue: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return entity.Issue{}, fmt.Errorf("read created issue id: %w", err)
	}
	return s.Issue(ctx, id)
}

func (s *Store) Issues(ctx context.Context) ([]entity.Issue, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+issueColumns()+` FROM issues ORDER BY updated_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list issues: %w", err)
	}
	defer rows.Close()

	var issues []entity.Issue
	for rows.Next() {
		item, err := scanIssue(rows)
		if err != nil {
			return nil, err
		}
		issues = append(issues, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate issues: %w", err)
	}
	return issues, nil
}

func (s *Store) IssuesByStates(ctx context.Context, states []entity.Status) ([]entity.Issue, error) {
	if len(states) == 0 {
		return s.Issues(ctx)
	}
	args := make([]any, 0, len(states))
	for _, status := range states {
		if !entity.IsValidStatus(status) {
			return nil, errors.New("status is invalid")
		}
		args = append(args, status)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+issueColumns()+` FROM issues WHERE status IN (`+placeholders(len(states))+`) ORDER BY updated_at DESC, id DESC`, args...)
	if err != nil {
		return nil, fmt.Errorf("list issues by states: %w", err)
	}
	defer rows.Close()

	var issues []entity.Issue
	for rows.Next() {
		item, err := scanIssue(rows)
		if err != nil {
			return nil, err
		}
		issues = append(issues, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate issues by states: %w", err)
	}
	return issues, nil
}

func (s *Store) IssueStatesByIDs(ctx context.Context, ids []int64) ([]entity.IssueState, error) {
	if len(ids) == 0 {
		return []entity.IssueState{}, nil
	}
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, status FROM issues WHERE id IN (`+placeholders(len(ids))+`) ORDER BY id ASC`, args...)
	if err != nil {
		return nil, fmt.Errorf("list issue states by ids: %w", err)
	}
	defer rows.Close()

	states := []entity.IssueState{}
	for rows.Next() {
		var state entity.IssueState
		if err := rows.Scan(&state.ID, &state.Status); err != nil {
			return nil, err
		}
		states = append(states, state)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate issue states by ids: %w", err)
	}
	return states, nil
}

func (s *Store) Issue(ctx context.Context, id int64) (entity.Issue, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+issueColumns()+` FROM issues WHERE id = ?`, id)
	item, err := scanIssue(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Issue{}, sql.ErrNoRows
		}
		return entity.Issue{}, fmt.Errorf("read issue: %w", err)
	}
	return item, nil
}

func (s *Store) UpdateIssue(ctx context.Context, id int64, input entity.UpdateIssueInput) (entity.Issue, error) {
	current, err := s.Issue(ctx, id)
	if err != nil {
		return entity.Issue{}, err
	}
	if input.Title != nil {
		current.Title = *input.Title
	}
	if input.Description != nil {
		current.Description = *input.Description
	}
	if input.Status != nil {
		if !entity.IsValidStatus(*input.Status) {
			return entity.Issue{}, errors.New("status is invalid")
		}
		current.Status = *input.Status
	}
	if input.Priority != nil {
		if !entity.IsValidPriority(*input.Priority) {
			return entity.Issue{}, errors.New("priority is invalid")
		}
		current.Priority = *input.Priority
	}
	if input.Assignee != nil {
		current.Assignee = *input.Assignee
	}
	_, err = s.db.ExecContext(ctx, `UPDATE issues SET
		title = ?, description = ?, status = ?, priority = ?, assignee = ?, updated_at = ?
		WHERE id = ?`,
		current.Title,
		current.Description,
		current.Status,
		current.Priority,
		current.Assignee,
		nowString(),
		id,
	)
	if err != nil {
		return entity.Issue{}, fmt.Errorf("update issue: %w", err)
	}
	return s.Issue(ctx, id)
}

func (s *Store) CreateComment(ctx context.Context, input entity.CreateCommentInput) (entity.Comment, error) {
	normalized, err := entity.NormalizeCreateComment(input)
	if err != nil {
		return entity.Comment{}, err
	}
	if _, err := s.Issue(ctx, normalized.IssueID); err != nil {
		return entity.Comment{}, err
	}
	now := nowString()
	result, err := s.db.ExecContext(ctx, `INSERT INTO comments (
		issue_id, author, type, body, created_at
	) VALUES (?, ?, ?, ?, ?)`,
		normalized.IssueID,
		normalized.Author,
		normalized.Type,
		normalized.Body,
		now,
	)
	if err != nil {
		return entity.Comment{}, fmt.Errorf("create comment: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return entity.Comment{}, fmt.Errorf("read created comment id: %w", err)
	}
	return s.Comment(ctx, id)
}

func (s *Store) CommentsByIssueID(ctx context.Context, issueID int64, cursor int64, limit int) ([]entity.Comment, error) {
	if issueID <= 0 {
		return nil, errors.New("issueId is required")
	}
	if cursor < 0 {
		return nil, errors.New("cursor is invalid")
	}
	if _, err := s.Issue(ctx, issueID); err != nil {
		return nil, err
	}
	limit = normalizeCommentLimit(limit)
	rows, err := s.db.QueryContext(ctx, `SELECT `+commentColumns()+` FROM comments
		WHERE issue_id = ? AND id > ?
		ORDER BY id ASC
		LIMIT ?`, issueID, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()

	comments := []entity.Comment{}
	for rows.Next() {
		item, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		comments = append(comments, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate comments: %w", err)
	}
	return comments, nil
}

func (s *Store) Comment(ctx context.Context, id int64) (entity.Comment, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+commentColumns()+` FROM comments WHERE id = ?`, id)
	item, err := scanComment(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Comment{}, sql.ErrNoRows
		}
		return entity.Comment{}, fmt.Errorf("read comment: %w", err)
	}
	return item, nil
}

func (s *Store) Summary(ctx context.Context) (entity.Summary, error) {
	issues, err := s.Issues(ctx)
	if err != nil {
		return entity.Summary{}, err
	}
	columns := make([]entity.Column, 0, len(entity.OrderedStatuses()))
	for _, status := range entity.OrderedStatuses() {
		column := entity.Column{Status: status, Title: entity.StatusTitle(status), Issues: []entity.IssueSummary{}}
		for _, item := range issues {
			if item.Status != status {
				continue
			}
			column.Issues = append(column.Issues, entity.IssueSummary{Issue: item})
		}
		columns = append(columns, column)
	}
	return entity.Summary{Columns: columns, GeneratedAt: time.Now().UTC()}, nil
}

func issueColumns() string {
	return `id, title, description, status, priority, assignee, created_at, updated_at`
}

func commentColumns() string {
	return `id, issue_id, author, type, body, created_at`
}

func projectColumns() string {
	return `id, key, name, description, created_at, updated_at`
}

func normalizeCommentLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func workspaceColumns() string {
	return `id, project_id, name, path, status, created_at, updated_at`
}

func placeholders(count int) string {
	return strings.TrimRight(strings.Repeat("?,", count), ",")
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanIssue(row rowScanner) (entity.Issue, error) {
	var item entity.Issue
	var createdAt string
	var updatedAt string
	err := row.Scan(&item.ID, &item.Title, &item.Description, &item.Status, &item.Priority, &item.Assignee, &createdAt, &updatedAt)
	if err != nil {
		return entity.Issue{}, err
	}
	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return entity.Issue{}, err
	}
	parsedUpdatedAt, err := parseTime(updatedAt)
	if err != nil {
		return entity.Issue{}, err
	}
	item.CreatedAt = parsedCreatedAt
	item.UpdatedAt = parsedUpdatedAt
	return item, nil
}

func scanComment(row rowScanner) (entity.Comment, error) {
	var item entity.Comment
	var createdAt string
	err := row.Scan(&item.ID, &item.IssueID, &item.Author, &item.Type, &item.Body, &createdAt)
	if err != nil {
		return entity.Comment{}, err
	}
	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return entity.Comment{}, err
	}
	item.CreatedAt = parsedCreatedAt
	return item, nil
}

func scanProject(row rowScanner) (entity.Project, error) {
	var item entity.Project
	var createdAt string
	var updatedAt string
	err := row.Scan(&item.ID, &item.Key, &item.Name, &item.Description, &createdAt, &updatedAt)
	if err != nil {
		return entity.Project{}, err
	}
	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return entity.Project{}, err
	}
	parsedUpdatedAt, err := parseTime(updatedAt)
	if err != nil {
		return entity.Project{}, err
	}
	item.CreatedAt = parsedCreatedAt
	item.UpdatedAt = parsedUpdatedAt
	return item, nil
}

func scanWorkspace(row rowScanner) (entity.Workspace, error) {
	var item entity.Workspace
	var createdAt string
	var updatedAt string
	err := row.Scan(&item.ID, &item.ProjectID, &item.Name, &item.Path, &item.Status, &createdAt, &updatedAt)
	if err != nil {
		return entity.Workspace{}, err
	}
	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return entity.Workspace{}, err
	}
	parsedUpdatedAt, err := parseTime(updatedAt)
	if err != nil {
		return entity.Workspace{}, err
	}
	item.CreatedAt = parsedCreatedAt
	item.UpdatedAt = parsedUpdatedAt
	return item, nil
}

func nowString() string {
	return formatTime(time.Now().UTC())
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse time: %w", err)
	}
	return parsed, nil
}
