package runstore

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/version-1/tasq/db/schema"
	"github.com/version-1/tasq/internal/orchestrator/run"
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
	if _, err := s.db.ExecContext(ctx, schema.Orchestrator); err != nil {
		return fmt.Errorf("migrate orchestrator sqlite: %w", err)
	}
	for _, column := range []struct {
		name       string
		definition string
	}{
		{name: "source_path", definition: "TEXT NOT NULL DEFAULT ''"},
		{name: "populated_at", definition: "TEXT NOT NULL DEFAULT ''"},
		{name: "cleanup_status", definition: "TEXT NOT NULL DEFAULT ''"},
		{name: "cleanup_at", definition: "TEXT NOT NULL DEFAULT ''"},
		{name: "last_error", definition: "TEXT NOT NULL DEFAULT ''"},
	} {
		if err := s.addColumnIfMissing(ctx, "workspace_metadata", column.name, column.definition); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) CreateRun(ctx context.Context, input CreateRunInput) (run.Run, error) {
	if err := ValidateCreateRun(input); err != nil {
		return run.Run{}, err
	}
	runID, err := randomID("run")
	if err != nil {
		return run.Run{}, err
	}
	now := nowString()
	_, err = s.db.ExecContext(ctx, `INSERT INTO runs (
		run_id, issue_id, status, workspace, attempt, orchestrator_id, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		runID,
		input.IssueID,
		run.StatusQueued,
		input.Workspace,
		input.Attempt,
		input.OrchestratorID,
		now,
		now,
	)
	if err != nil {
		return run.Run{}, fmt.Errorf("create run: %w", err)
	}
	createdRun, err := s.RunByRunID(ctx, runID)
	if err != nil {
		return run.Run{}, err
	}
	return createdRun, nil
}

type CreateRunInput struct {
	IssueID        int64
	Workspace      string
	Attempt        int
	OrchestratorID string
}

func (s *Store) UpdateRunStatus(ctx context.Context, runID string, status run.Status, errText string) (run.Run, error) {
	if runID == "" {
		return run.Run{}, fmt.Errorf("runId is required")
	}
	if err := ValidateRunStatus(status); err != nil {
		return run.Run{}, err
	}
	if runeCount(errText) > maxRunErrorLength {
		return run.Run{}, fmt.Errorf("error must be 10000 characters or fewer")
	}
	now := nowString()
	_, err := s.db.ExecContext(ctx, `UPDATE runs SET status = ?, error = ?, updated_at = ? WHERE run_id = ?`, status, errText, now, runID)
	if err != nil {
		return run.Run{}, fmt.Errorf("update run status: %w", err)
	}
	updatedRun, err := s.RunByRunID(ctx, runID)
	if err != nil {
		return run.Run{}, err
	}
	return updatedRun, nil
}

func (s *Store) RecordRunnerEvent(ctx context.Context, runID string, eventType string, message string, payloadJSON string) error {
	if eventType == "" {
		eventType = "event"
	}
	occurredAt := time.Now().UTC()
	if err := ValidateRunnerEvent(runID, eventType, message, payloadJSON, occurredAt); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO runner_events (
		run_id, event_type, message, payload_json, occurred_at
	) VALUES (?, ?, ?, ?, ?)`, runID, eventType, message, payloadJSON, formatTime(occurredAt))
	if err != nil {
		return fmt.Errorf("record runner event: %w", err)
	}
	return nil
}

func (s *Store) RunnerEvents(ctx context.Context, runID string, limit int) ([]run.RunnerEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, run_id, event_type, message, payload_json, occurred_at
		FROM runner_events
		WHERE run_id = ?
		ORDER BY id ASC
		LIMIT ?`, runID, limit)
	if err != nil {
		return nil, fmt.Errorf("list runner events: %w", err)
	}
	defer rows.Close()
	var events []run.RunnerEvent
	for rows.Next() {
		var event run.RunnerEvent
		var occurredAt string
		if err := rows.Scan(&event.ID, &event.RunID, &event.EventType, &event.Message, &event.PayloadJSON, &occurredAt); err != nil {
			return nil, err
		}
		parsed, err := parseTime(occurredAt)
		if err != nil {
			return nil, err
		}
		event.OccurredAt = parsed
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *Store) UpsertWorkspaceMetadata(ctx context.Context, input WorkspaceMetadataInput) error {
	if err := ValidateWorkspaceMetadata(input); err != nil {
		return err
	}
	now := nowString()
	_, err := s.db.ExecContext(ctx, `INSERT INTO workspace_metadata (
		workspace_key, issue_id, path, created_now, source_path, populated_at, cleanup_status, cleanup_at, last_error, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(workspace_key) DO UPDATE SET
		issue_id = excluded.issue_id,
		path = excluded.path,
		created_now = excluded.created_now,
		source_path = excluded.source_path,
		populated_at = excluded.populated_at,
		cleanup_status = '',
		cleanup_at = '',
		last_error = '',
		updated_at = excluded.updated_at`,
		input.WorkspaceKey,
		input.IssueID,
		input.Path,
		boolInt(input.CreatedNow),
		input.SourcePath,
		now,
		"",
		"",
		"",
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("upsert workspace metadata: %w", err)
	}
	return nil
}

type WorkspaceMetadataInput struct {
	WorkspaceKey string
	IssueID      int64
	Path         string
	CreatedNow   bool
	SourcePath   string
}

func (s *Store) MarkWorkspaceCleanup(ctx context.Context, workspaceKey string, status string, errText string) error {
	now := nowString()
	_, err := s.db.ExecContext(ctx, `UPDATE workspace_metadata
		SET cleanup_status = ?, cleanup_at = ?, last_error = ?, updated_at = ?
		WHERE workspace_key = ?`, status, now, errText, now, workspaceKey)
	if err != nil {
		return fmt.Errorf("mark workspace cleanup: %w", err)
	}
	return nil
}

func (s *Store) RecordWorkspaceSetupFailure(ctx context.Context, issueID int64, workspaceKey string, path string, errText string) error {
	if err := ValidateWorkspaceSetupFailure(issueID, workspaceKey, path, errText); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT INTO workspace_setup_failures (
		issue_id, workspace_key, path, error, occurred_at
	) VALUES (?, ?, ?, ?, ?)`, issueID, workspaceKey, path, errText, nowString())
	if err != nil {
		return fmt.Errorf("record workspace setup failure: %w", err)
	}
	return nil
}

func (s *Store) WorkspaceSetupFailureCount(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM workspace_setup_failures`).Scan(&count); err != nil {
		return 0, fmt.Errorf("count workspace setup failures: %w", err)
	}
	return count, nil
}

func (s *Store) RunByRunID(ctx context.Context, runID string) (run.Run, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, run_id, issue_id, status, workspace, attempt, error, orchestrator_id, created_at, updated_at
		FROM runs WHERE run_id = ?`, runID)
	return scanRun(row)
}

func (s *Store) ActiveRuns(ctx context.Context) ([]run.Run, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, run_id, issue_id, status, workspace, attempt, error, orchestrator_id, created_at, updated_at
		FROM runs
		WHERE status IN (?, ?)
		ORDER BY updated_at DESC, id DESC`, run.StatusQueued, run.StatusRunning)
	if err != nil {
		return nil, fmt.Errorf("list active runs: %w", err)
	}
	defer rows.Close()

	var runs []run.Run
	for rows.Next() {
		storedRun, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		runs = append(runs, storedRun)
	}
	return runs, rows.Err()
}

func (s *Store) RunByIssueID(ctx context.Context, issueID int64) (run.Run, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, run_id, issue_id, status, workspace, attempt, error, orchestrator_id, created_at, updated_at
		FROM runs
		WHERE issue_id = ?
		ORDER BY id DESC
		LIMIT 1`, issueID)
	return scanRun(row)
}

func scanRun(row scanner) (run.Run, error) {
	var storedRun run.Run
	var createdAt string
	var updatedAt string
	if err := row.Scan(&storedRun.ID, &storedRun.RunID, &storedRun.IssueID, &storedRun.Status, &storedRun.Workspace, &storedRun.Attempt, &storedRun.Error, &storedRun.OrchestratorID, &createdAt, &updatedAt); err != nil {
		return run.Run{}, err
	}
	var err error
	storedRun.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return run.Run{}, err
	}
	storedRun.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return run.Run{}, err
	}
	return storedRun, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func (s *Store) addColumnIfMissing(ctx context.Context, table string, column string, definition string) error {
	rows, err := s.db.QueryContext(ctx, `PRAGMA table_info(`+table+`)`)
	if err != nil {
		return fmt.Errorf("inspect %s columns: %w", table, err)
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
			return err
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `ALTER TABLE `+table+` ADD COLUMN `+column+` `+definition); err != nil {
		return fmt.Errorf("add %s.%s column: %w", table, column, err)
	}
	return nil
}

func randomID(prefix string) (string, error) {
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate %s id: %w", prefix, err)
	}
	return prefix + "_" + hex.EncodeToString(raw[:]), nil
}

func nowString() string {
	return formatTime(time.Now().UTC())
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
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
