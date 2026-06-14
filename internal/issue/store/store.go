package store

import (
	"context"
	"database/sql"
	"encoding/json"
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

type IssueFilter struct {
	States    []entity.Status
	ProjectID *int64
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
	if err := s.rebuildIssuesForProjectScopeIfNeeded(ctx); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, schema.IssueTracker); err != nil {
		return fmt.Errorf("migrate issue tracker sqlite: %w", err)
	}
	if err := s.addColumnIfMissing(ctx, "projects", "location", "TEXT NOT NULL DEFAULT ''"); err != nil {
		return err
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
	var issueCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM issues WHERE project_id = ?`, id).Scan(&issueCount); err != nil {
		return fmt.Errorf("count project issues: %w", err)
	}
	if issueCount > 0 {
		return errors.New("project has linked issues")
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
	if err := tx.Commit(); err != nil {
		return err
	}
	tx = nil
	return nil
}

func (s *Store) UpsertProjectWorkflow(ctx context.Context, input entity.UpsertProjectWorkflowInput) (entity.ProjectWorkflow, error) {
	normalized, err := entity.NormalizeUpsertProjectWorkflow(input)
	if err != nil {
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
	result, err := s.db.ExecContext(ctx, `DELETE FROM project_workflows WHERE project_id = ?`, projectID)
	if err != nil {
		return fmt.Errorf("delete project workflow: %w", err)
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
	if _, err := s.Project(ctx, normalized.ProjectID); err != nil {
		return entity.Issue{}, err
	}
	now := nowString()
	result, err := s.db.ExecContext(ctx, `INSERT INTO issues (
		project_id, title, description, status, priority, assignee, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		normalized.ProjectID,
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
	return s.IssuesByFilter(ctx, IssueFilter{})
}

func (s *Store) IssuesByStates(ctx context.Context, states []entity.Status) ([]entity.Issue, error) {
	return s.IssuesByFilter(ctx, IssueFilter{States: states})
}

func (s *Store) IssuesByFilter(ctx context.Context, filter IssueFilter) ([]entity.Issue, error) {
	clauses := []string{}
	args := []any{}
	if len(filter.States) > 0 {
		for _, status := range filter.States {
			if !entity.IsValidStatus(status) {
				return nil, errors.New("status is invalid")
			}
			args = append(args, status)
		}
		clauses = append(clauses, `issues.status IN (`+placeholders(len(filter.States))+`)`)
	}
	if filter.ProjectID != nil {
		if *filter.ProjectID <= 0 {
			return nil, errors.New("projectId is invalid")
		}
		args = append(args, *filter.ProjectID)
		clauses = append(clauses, `issues.project_id = ?`)
	}
	where := ""
	if len(clauses) > 0 {
		where = " WHERE " + strings.Join(clauses, " AND ")
	}
	rows, err := s.db.QueryContext(ctx, `SELECT `+issueColumns()+` FROM issues JOIN projects ON projects.id = issues.project_id`+where+` ORDER BY issues.updated_at DESC, issues.id DESC`, args...)
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
	row := s.db.QueryRowContext(ctx, `SELECT `+issueColumns()+` FROM issues JOIN projects ON projects.id = issues.project_id WHERE issues.id = ?`, id)
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
	normalized, err := entity.NormalizeUpdateIssue(input)
	if err != nil {
		return entity.Issue{}, err
	}
	current, err := s.Issue(ctx, id)
	if err != nil {
		return entity.Issue{}, err
	}
	if normalized.Title != nil {
		current.Title = *normalized.Title
	}
	if normalized.Description != nil {
		current.Description = *normalized.Description
	}
	if normalized.Status != nil {
		current.Status = *normalized.Status
	}
	if normalized.Priority != nil {
		current.Priority = *normalized.Priority
	}
	if normalized.Assignee != nil {
		current.Assignee = *normalized.Assignee
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

func (s *Store) UpdateComment(ctx context.Context, id int64, input entity.UpdateCommentInput) (entity.Comment, error) {
	normalized, err := entity.NormalizeUpdateComment(input)
	if err != nil {
		return entity.Comment{}, err
	}
	current, err := s.Comment(ctx, id)
	if err != nil {
		return entity.Comment{}, err
	}
	if normalized.Body != nil {
		current.Body = *normalized.Body
	}
	_, err = s.db.ExecContext(ctx, `UPDATE comments SET body = ? WHERE id = ?`, current.Body, id)
	if err != nil {
		return entity.Comment{}, fmt.Errorf("update comment: %w", err)
	}
	return s.Comment(ctx, id)
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
	return `issues.id, issues.project_id, projects.key, issues.title, issues.description, issues.status, issues.priority, issues.assignee, issues.created_at, issues.updated_at`
}

func commentColumns() string {
	return `id, issue_id, author, type, body, created_at`
}

func projectColumns() string {
	return `id, key, name, description, location, created_at, updated_at`
}

func projectWorkflowColumns() string {
	return `project_id, frontmatter_json, body, checksum, created_at, updated_at`
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
	err := row.Scan(&item.ID, &item.ProjectID, &item.ProjectKey, &item.Title, &item.Description, &item.Status, &item.Priority, &item.Assignee, &createdAt, &updatedAt)
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
	err := row.Scan(&item.ID, &item.Key, &item.Name, &item.Description, &item.Location, &createdAt, &updatedAt)
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

func scanProjectWorkflow(row rowScanner) (entity.ProjectWorkflow, error) {
	var item entity.ProjectWorkflow
	var frontmatterJSON string
	var createdAt string
	var updatedAt string
	err := row.Scan(&item.ProjectID, &frontmatterJSON, &item.Body, &item.Checksum, &createdAt, &updatedAt)
	if err != nil {
		return entity.ProjectWorkflow{}, err
	}
	if err := json.Unmarshal([]byte(frontmatterJSON), &item.Frontmatter); err != nil {
		return entity.ProjectWorkflow{}, fmt.Errorf("parse project workflow frontmatter: %w", err)
	}
	if item.Frontmatter == nil {
		item.Frontmatter = map[string]any{}
	}
	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return entity.ProjectWorkflow{}, err
	}
	parsedUpdatedAt, err := parseTime(updatedAt)
	if err != nil {
		return entity.ProjectWorkflow{}, err
	}
	item.CreatedAt = parsedCreatedAt
	item.UpdatedAt = parsedUpdatedAt
	return item, nil
}

func (s *Store) addColumnIfMissing(ctx context.Context, table string, column string, definition string) error {
	exists, err := s.hasColumn(ctx, table, column)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE `+table+` ADD COLUMN `+column+` `+definition); err != nil {
		return fmt.Errorf("add %s.%s column: %w", table, column, err)
	}
	return nil
}

func (s *Store) hasColumn(ctx context.Context, table string, column string) (bool, error) {
	rows, err := s.db.QueryContext(ctx, `PRAGMA table_info(`+table+`)`)
	if err != nil {
		return false, fmt.Errorf("inspect %s columns: %w", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name string
		var typ string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	return false, nil
}

func (s *Store) rebuildIssuesForProjectScopeIfNeeded(ctx context.Context) error {
	hasIssues, err := s.tableExists(ctx, "issues")
	if err != nil {
		return err
	}
	if !hasIssues {
		return nil
	}
	hasProjectID, err := s.hasColumn(ctx, "issues", "project_id")
	if err != nil {
		return err
	}
	if hasProjectID {
		return nil
	}
	hasAttachments, err := s.tableExists(ctx, "attachments")
	if err != nil {
		return err
	}
	hasComments, err := s.tableExists(ctx, "comments")
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()
	statements := []string{}
	if hasAttachments {
		statements = append(statements, `DELETE FROM attachments WHERE entity_type IN ('issue', 'comment')`)
	}
	if hasComments {
		statements = append(statements, `DELETE FROM comments`)
	}
	statements = append(statements,
		`DROP TABLE issues`,
		`CREATE TABLE issues (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			priority TEXT NOT NULL,
			assignee TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY(project_id) REFERENCES projects(id)
		)`,
		`CREATE INDEX IF NOT EXISTS issues_project_id_idx ON issues(project_id)`,
		`CREATE INDEX IF NOT EXISTS issues_status_idx ON issues(status)`,
	)
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("rebuild issues for project scope: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	tx = nil
	return nil
}

func (s *Store) tableExists(ctx context.Context, table string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS (
		SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = ?
	)`, table).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("inspect %s table: %w", table, err)
	}
	return exists, nil
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
