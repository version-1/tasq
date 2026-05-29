package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
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

	var projects []entity.Project
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
	if normalized.Status == entity.StatusReady {
		if err := s.ensurePendingWorkItem(ctx, id); err != nil {
			return entity.Issue{}, err
		}
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
	if current.Status == entity.StatusReady {
		if err := s.ensurePendingWorkItem(ctx, id); err != nil {
			return entity.Issue{}, err
		}
	}
	return s.Issue(ctx, id)
}

func (s *Store) ClaimWorkItem(ctx context.Context, input entity.ClaimWorkItemInput) (*entity.WorkItem, error) {
	if input.OrchestratorID == "" {
		return nil, errors.New("orchestratorId is required")
	}
	if input.LeaseSeconds <= 0 {
		input.LeaseSeconds = 30
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().UTC()
	row := tx.QueryRowContext(ctx, `SELECT id FROM work_items
		WHERE status = ? OR (status = ? AND lease_until < ?)
		ORDER BY created_at ASC, id ASC
		LIMIT 1`, entity.WorkItemPending, entity.WorkItemClaimed, formatTime(now))
	var id int64
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find claimable work item: %w", err)
	}
	token, err := randomToken()
	if err != nil {
		return nil, err
	}
	leaseUntil := now.Add(time.Duration(input.LeaseSeconds) * time.Second)
	_, err = tx.ExecContext(ctx, `UPDATE work_items SET
		status = ?, claimed_by = ?, claim_token = ?, lease_until = ?, attempt = attempt + 1, updated_at = ?
		WHERE id = ?`,
		entity.WorkItemClaimed,
		input.OrchestratorID,
		token,
		formatTime(leaseUntil),
		formatTime(now),
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("claim work item: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil
	item, err := s.WorkItem(ctx, id)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Store) WorkItem(ctx context.Context, id int64) (entity.WorkItem, error) {
	row := s.db.QueryRowContext(ctx, `SELECT
		w.id, w.issue_id, w.status, w.claimed_by, w.claim_token, w.lease_until, w.attempt, w.created_at, w.updated_at,
		i.id, i.title, i.description, i.status, i.priority, i.assignee, i.created_at, i.updated_at
		FROM work_items w
		JOIN issues i ON i.id = w.issue_id
		WHERE w.id = ?`, id)
	return scanWorkItem(row)
}

func (s *Store) ReceiveRunEvent(ctx context.Context, input entity.RunEventInput) error {
	if input.EventID == "" {
		return errors.New("eventId is required")
	}
	now := nowString()
	occurredAt := input.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}
	result, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO orchestrator_events (
		event_id, work_item_id, issue_id, run_id, claim_token, status, workspace, attempt, error, occurred_at, orchestrator_id, received_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		input.EventID,
		input.WorkItemID,
		input.IssueID,
		input.RunID,
		input.ClaimToken,
		input.Status,
		input.Workspace,
		input.Attempt,
		input.Error,
		formatTime(occurredAt),
		input.OrchestratorID,
		now,
	)
	if err != nil {
		return fmt.Errorf("record orchestrator event: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return nil
	}
	currentToken, err := s.currentClaimToken(ctx, input.WorkItemID)
	if err != nil {
		return err
	}
	if currentToken != input.ClaimToken {
		return nil
	}
	if err := s.upsertRunSnapshot(ctx, input, occurredAt); err != nil {
		return err
	}
	return s.applyRunStatus(ctx, input)
}

func (s *Store) Summary(ctx context.Context) (entity.Summary, error) {
	issues, err := s.Issues(ctx)
	if err != nil {
		return entity.Summary{}, err
	}
	runs, err := s.RunSnapshots(ctx)
	if err != nil {
		return entity.Summary{}, err
	}
	runByIssue := map[int64]entity.RunSnapshot{}
	for _, run := range runs {
		runByIssue[run.IssueID] = run
	}
	columns := make([]entity.Column, 0, len(entity.OrderedStatuses()))
	for _, status := range entity.OrderedStatuses() {
		column := entity.Column{Status: status, Title: entity.StatusTitle(status), Issues: []entity.IssueSummary{}}
		for _, item := range issues {
			if item.Status != status {
				continue
			}
			summary := entity.IssueSummary{Issue: item}
			if run, ok := runByIssue[item.ID]; ok {
				runCopy := run
				summary.Run = &runCopy
			}
			column.Issues = append(column.Issues, summary)
		}
		columns = append(columns, column)
	}
	return entity.Summary{Columns: columns, Runs: runs, GeneratedAt: time.Now().UTC()}, nil
}

func (s *Store) RunSnapshots(ctx context.Context) ([]entity.RunSnapshot, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT issue_id, work_item_id, run_id, status, workspace, attempt, error, orchestrator_id, updated_at
		FROM run_snapshots ORDER BY updated_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list run snapshots: %w", err)
	}
	defer rows.Close()
	var runs []entity.RunSnapshot
	for rows.Next() {
		var run entity.RunSnapshot
		var updatedAt string
		if err := rows.Scan(&run.IssueID, &run.WorkItemID, &run.RunID, &run.Status, &run.Workspace, &run.Attempt, &run.Error, &run.OrchestratorID, &updatedAt); err != nil {
			return nil, err
		}
		parsed, err := parseTime(updatedAt)
		if err != nil {
			return nil, err
		}
		run.UpdatedAt = parsed
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func (s *Store) ensurePendingWorkItem(ctx context.Context, issueID int64) error {
	var exists int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM work_items
		WHERE issue_id = ? AND status IN (?, ?)
		LIMIT 1`, issueID, entity.WorkItemPending, entity.WorkItemClaimed).Scan(&exists)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	now := nowString()
	_, err = s.db.ExecContext(ctx, `INSERT INTO work_items (
		issue_id, status, created_at, updated_at
	) VALUES (?, ?, ?, ?)`, issueID, entity.WorkItemPending, now, now)
	if err != nil {
		return fmt.Errorf("create work item: %w", err)
	}
	return nil
}

func (s *Store) currentClaimToken(ctx context.Context, workItemID int64) (string, error) {
	var token string
	err := s.db.QueryRowContext(ctx, `SELECT claim_token FROM work_items WHERE id = ?`, workItemID).Scan(&token)
	if err != nil {
		return "", fmt.Errorf("read claim token: %w", err)
	}
	return token, nil
}

func (s *Store) upsertRunSnapshot(ctx context.Context, input entity.RunEventInput, occurredAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO run_snapshots (
		issue_id, work_item_id, run_id, status, workspace, attempt, error, orchestrator_id, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(issue_id) DO UPDATE SET
		work_item_id = excluded.work_item_id,
		run_id = excluded.run_id,
		status = excluded.status,
		workspace = excluded.workspace,
		attempt = excluded.attempt,
		error = excluded.error,
		orchestrator_id = excluded.orchestrator_id,
		updated_at = excluded.updated_at`,
		input.IssueID,
		input.WorkItemID,
		input.RunID,
		input.Status,
		input.Workspace,
		input.Attempt,
		input.Error,
		input.OrchestratorID,
		formatTime(occurredAt),
	)
	if err != nil {
		return fmt.Errorf("upsert run snapshot: %w", err)
	}
	return nil
}

func (s *Store) applyRunStatus(ctx context.Context, input entity.RunEventInput) error {
	now := nowString()
	switch input.Status {
	case entity.RunQueued, entity.RunStarting, entity.RunRunning:
		_, err := s.db.ExecContext(ctx, `UPDATE issues SET status = ?, updated_at = ? WHERE id = ?`, entity.StatusInProgress, now, input.IssueID)
		return err
	case entity.RunWaitingForInput:
		_, err := s.db.ExecContext(ctx, `UPDATE issues SET status = ?, updated_at = ? WHERE id = ?`, entity.StatusBlocked, now, input.IssueID)
		return err
	case entity.RunSucceeded:
		_, err := s.db.ExecContext(ctx, `UPDATE work_items SET status = ?, updated_at = ? WHERE id = ?`, entity.WorkItemDone, now, input.WorkItemID)
		if err != nil {
			return err
		}
		_, err = s.db.ExecContext(ctx, `UPDATE issues SET status = ?, updated_at = ? WHERE id = ?`, entity.StatusReview, now, input.IssueID)
		return err
	case entity.RunFailed, entity.RunCancelled:
		_, err := s.db.ExecContext(ctx, `UPDATE work_items SET status = ?, updated_at = ? WHERE id = ?`, entity.WorkItemFailed, now, input.WorkItemID)
		if err != nil {
			return err
		}
		_, err = s.db.ExecContext(ctx, `UPDATE issues SET status = ?, updated_at = ? WHERE id = ?`, entity.StatusFailed, now, input.IssueID)
		return err
	default:
		return nil
	}
}

func issueColumns() string {
	return `id, title, description, status, priority, assignee, created_at, updated_at`
}

func projectColumns() string {
	return `id, key, name, description, created_at, updated_at`
}

func workspaceColumns() string {
	return `id, project_id, name, path, status, created_at, updated_at`
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

func scanWorkItem(row rowScanner) (entity.WorkItem, error) {
	var item entity.WorkItem
	var leaseUntil string
	var createdAt string
	var updatedAt string
	var issueCreatedAt string
	var issueUpdatedAt string
	err := row.Scan(
		&item.ID, &item.IssueID, &item.Status, &item.ClaimedBy, &item.ClaimToken, &leaseUntil, &item.Attempt, &createdAt, &updatedAt,
		&item.Issue.ID, &item.Issue.Title, &item.Issue.Description, &item.Issue.Status, &item.Issue.Priority, &item.Issue.Assignee, &issueCreatedAt, &issueUpdatedAt,
	)
	if err != nil {
		return entity.WorkItem{}, err
	}
	if leaseUntil != "" {
		parsed, err := parseTime(leaseUntil)
		if err != nil {
			return entity.WorkItem{}, err
		}
		item.LeaseUntil = &parsed
	}
	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return entity.WorkItem{}, err
	}
	parsedUpdatedAt, err := parseTime(updatedAt)
	if err != nil {
		return entity.WorkItem{}, err
	}
	parsedIssueCreatedAt, err := parseTime(issueCreatedAt)
	if err != nil {
		return entity.WorkItem{}, err
	}
	parsedIssueUpdatedAt, err := parseTime(issueUpdatedAt)
	if err != nil {
		return entity.WorkItem{}, err
	}
	item.CreatedAt = parsedCreatedAt
	item.UpdatedAt = parsedUpdatedAt
	item.Issue.CreatedAt = parsedIssueCreatedAt
	item.Issue.UpdatedAt = parsedIssueUpdatedAt
	return item, nil
}

func randomToken() (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate claim token: %w", err)
	}
	return hex.EncodeToString(raw[:]), nil
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
