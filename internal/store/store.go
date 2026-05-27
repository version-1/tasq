package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/version-1/tasq/internal/task"
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
	if err := store.ensureDefaultSettings(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			priority TEXT NOT NULL,
			agent_status TEXT NOT NULL,
			assignee TEXT NOT NULL DEFAULT '',
			source TEXT NOT NULL DEFAULT '',
			source_id TEXT NOT NULL DEFAULT '',
			workspace TEXT NOT NULL DEFAULT '',
			attempts INTEGER NOT NULL DEFAULT 0,
			last_error TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS tasks_status_idx ON tasks(status)`,
		`CREATE TABLE IF NOT EXISTS settings (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			poll_interval_seconds INTEGER NOT NULL,
			max_concurrent_runs INTEGER NOT NULL,
			workspace_root TEXT NOT NULL,
			tracker_provider TEXT NOT NULL,
			agent_command TEXT NOT NULL
		)`,
	}
	for _, statement := range statements {
		if _, err := s.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("migrate sqlite: %w", err)
		}
	}
	return nil
}

func (s *Store) ensureDefaultSettings(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO settings (
		id, poll_interval_seconds, max_concurrent_runs, workspace_root, tracker_provider, agent_command
	) VALUES (1, 30, 2, '.workspaces', 'linear', 'codex')`)
	if err != nil {
		return fmt.Errorf("seed settings: %w", err)
	}
	return nil
}

func (s *Store) CreateTask(ctx context.Context, input task.CreateTaskInput) (task.Task, error) {
	normalized, err := task.NormalizeCreate(input)
	if err != nil {
		return task.Task{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := s.db.ExecContext(ctx, `INSERT INTO tasks (
		title, description, status, priority, agent_status, assignee, source, source_id, workspace, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		normalized.Title,
		normalized.Description,
		normalized.Status,
		normalized.Priority,
		task.AgentStatusQueued,
		normalized.Assignee,
		normalized.Source,
		normalized.SourceID,
		normalized.Workspace,
		now,
		now,
	)
	if err != nil {
		return task.Task{}, fmt.Errorf("create task: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return task.Task{}, fmt.Errorf("read created task id: %w", err)
	}
	return s.Task(ctx, id)
}

func (s *Store) Tasks(ctx context.Context) ([]task.Task, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+taskColumns()+` FROM tasks ORDER BY updated_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []task.Task
	for rows.Next() {
		item, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}
	return tasks, nil
}

func (s *Store) Task(ctx context.Context, id int64) (task.Task, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+taskColumns()+` FROM tasks WHERE id = ?`, id)
	item, err := scanTask(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return task.Task{}, sql.ErrNoRows
		}
		return task.Task{}, fmt.Errorf("read task: %w", err)
	}
	return item, nil
}

func (s *Store) UpdateTask(ctx context.Context, id int64, input task.UpdateTaskInput) (task.Task, error) {
	current, err := s.Task(ctx, id)
	if err != nil {
		return task.Task{}, err
	}
	if input.Title != nil {
		current.Title = *input.Title
	}
	if input.Description != nil {
		current.Description = *input.Description
	}
	if input.Status != nil {
		if !task.IsValidStatus(*input.Status) {
			return task.Task{}, errors.New("status is invalid")
		}
		current.Status = *input.Status
	}
	if input.Priority != nil {
		if !task.IsValidPriority(*input.Priority) {
			return task.Task{}, errors.New("priority is invalid")
		}
		current.Priority = *input.Priority
	}
	if input.AgentStatus != nil {
		if !task.IsValidAgentStatus(*input.AgentStatus) {
			return task.Task{}, errors.New("agentStatus is invalid")
		}
		current.AgentStatus = *input.AgentStatus
	}
	if input.Assignee != nil {
		current.Assignee = *input.Assignee
	}
	if input.Workspace != nil {
		current.Workspace = *input.Workspace
	}
	if input.LastError != nil {
		current.LastError = *input.LastError
	}
	updatedAt := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = s.db.ExecContext(ctx, `UPDATE tasks SET
		title = ?, description = ?, status = ?, priority = ?, agent_status = ?, assignee = ?,
		workspace = ?, last_error = ?, updated_at = ?
		WHERE id = ?`,
		current.Title,
		current.Description,
		current.Status,
		current.Priority,
		current.AgentStatus,
		current.Assignee,
		current.Workspace,
		current.LastError,
		updatedAt,
		id,
	)
	if err != nil {
		return task.Task{}, fmt.Errorf("update task: %w", err)
	}
	return s.Task(ctx, id)
}

func (s *Store) Settings(ctx context.Context) (task.Settings, error) {
	var settings task.Settings
	err := s.db.QueryRowContext(ctx, `SELECT poll_interval_seconds, max_concurrent_runs, workspace_root, tracker_provider, agent_command FROM settings WHERE id = 1`).
		Scan(&settings.PollIntervalSeconds, &settings.MaxConcurrentRuns, &settings.WorkspaceRoot, &settings.TrackerProvider, &settings.AgentCommand)
	if err != nil {
		return task.Settings{}, fmt.Errorf("read settings: %w", err)
	}
	return settings, nil
}

func (s *Store) UpdateSettings(ctx context.Context, settings task.Settings) (task.Settings, error) {
	if settings.PollIntervalSeconds <= 0 {
		return task.Settings{}, errors.New("pollIntervalSeconds must be greater than zero")
	}
	if settings.MaxConcurrentRuns <= 0 {
		return task.Settings{}, errors.New("maxConcurrentRuns must be greater than zero")
	}
	_, err := s.db.ExecContext(ctx, `UPDATE settings SET
		poll_interval_seconds = ?, max_concurrent_runs = ?, workspace_root = ?, tracker_provider = ?, agent_command = ?
		WHERE id = 1`,
		settings.PollIntervalSeconds,
		settings.MaxConcurrentRuns,
		settings.WorkspaceRoot,
		settings.TrackerProvider,
		settings.AgentCommand,
	)
	if err != nil {
		return task.Settings{}, fmt.Errorf("update settings: %w", err)
	}
	return s.Settings(ctx)
}

func (s *Store) Summary(ctx context.Context) (task.Summary, error) {
	tasks, err := s.Tasks(ctx)
	if err != nil {
		return task.Summary{}, err
	}
	settings, err := s.Settings(ctx)
	if err != nil {
		return task.Summary{}, err
	}
	columns := make([]task.Column, 0, len(task.OrderedStatuses()))
	for _, status := range task.OrderedStatuses() {
		column := task.Column{Status: status, Title: task.StatusTitle(status), Tasks: []task.Task{}}
		for _, item := range tasks {
			if item.Status == status {
				column.Tasks = append(column.Tasks, item)
			}
		}
		columns = append(columns, column)
	}
	agents := []task.Task{}
	for _, item := range tasks {
		if item.AgentStatus == task.AgentStatusQueued || item.AgentStatus == task.AgentStatusRunning || item.AgentStatus == task.AgentStatusWaiting {
			agents = append(agents, item)
		}
	}
	return task.Summary{
		Columns:     columns,
		Agents:      agents,
		Settings:    settings,
		GeneratedAt: time.Now().UTC(),
	}, nil
}

func taskColumns() string {
	return `id, title, description, status, priority, agent_status, assignee, source, source_id, workspace, attempts, last_error, created_at, updated_at`
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTask(row rowScanner) (task.Task, error) {
	var item task.Task
	var createdAt string
	var updatedAt string
	err := row.Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.Status,
		&item.Priority,
		&item.AgentStatus,
		&item.Assignee,
		&item.Source,
		&item.SourceID,
		&item.Workspace,
		&item.Attempts,
		&item.LastError,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return task.Task{}, err
	}
	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return task.Task{}, fmt.Errorf("parse task created_at: %w", err)
	}
	parsedUpdatedAt, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return task.Task{}, fmt.Errorf("parse task updated_at: %w", err)
	}
	item.CreatedAt = parsedCreatedAt
	item.UpdatedAt = parsedUpdatedAt
	return item, nil
}
