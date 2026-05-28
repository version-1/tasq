package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/version-1/tasq/db/schema"
	"github.com/version-1/tasq/internal/orchestrator/domain/entity"
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
	return nil
}

func (s *Store) CreateRun(ctx context.Context, input CreateRunInput) (entity.Run, error) {
	runID, err := randomID("run")
	if err != nil {
		return entity.Run{}, err
	}
	now := nowString()
	_, err = s.db.ExecContext(ctx, `INSERT INTO runs (
		run_id, issue_id, work_item_id, claim_token, status, workspace, attempt, orchestrator_id, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		runID,
		input.IssueID,
		input.WorkItemID,
		input.ClaimToken,
		entity.RunQueued,
		input.Workspace,
		input.Attempt,
		input.OrchestratorID,
		now,
		now,
	)
	if err != nil {
		return entity.Run{}, fmt.Errorf("create run: %w", err)
	}
	run, err := s.RunByRunID(ctx, runID)
	if err != nil {
		return entity.Run{}, err
	}
	if err := s.EnqueueRunEvent(ctx, run); err != nil {
		return entity.Run{}, err
	}
	return run, nil
}

type CreateRunInput struct {
	IssueID        int64
	WorkItemID     int64
	ClaimToken     string
	Workspace      string
	Attempt        int
	OrchestratorID string
}

func (s *Store) UpdateRunStatus(ctx context.Context, runID string, status entity.RunStatus, errText string) (entity.Run, error) {
	now := nowString()
	_, err := s.db.ExecContext(ctx, `UPDATE runs SET status = ?, error = ?, updated_at = ? WHERE run_id = ?`, status, errText, now, runID)
	if err != nil {
		return entity.Run{}, fmt.Errorf("update run status: %w", err)
	}
	run, err := s.RunByRunID(ctx, runID)
	if err != nil {
		return entity.Run{}, err
	}
	if err := s.EnqueueRunEvent(ctx, run); err != nil {
		return entity.Run{}, err
	}
	return run, nil
}

func (s *Store) RunByRunID(ctx context.Context, runID string) (entity.Run, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, run_id, issue_id, work_item_id, claim_token, status, workspace, attempt, error, orchestrator_id, created_at, updated_at
		FROM runs WHERE run_id = ?`, runID)
	var run entity.Run
	var createdAt string
	var updatedAt string
	err := row.Scan(&run.ID, &run.RunID, &run.IssueID, &run.WorkItemID, &run.ClaimToken, &run.Status, &run.Workspace, &run.Attempt, &run.Error, &run.OrchestratorID, &createdAt, &updatedAt)
	if err != nil {
		return entity.Run{}, err
	}
	run.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return entity.Run{}, err
	}
	run.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return entity.Run{}, err
	}
	return run, nil
}

func (s *Store) EnqueueRunEvent(ctx context.Context, run entity.Run) error {
	eventID, err := randomID("evt")
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO outbox_events (
		event_id, run_id, issue_id, work_item_id, claim_token, status, workspace, attempt, error, orchestrator_id, occurred_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		eventID,
		run.RunID,
		run.IssueID,
		run.WorkItemID,
		run.ClaimToken,
		run.Status,
		run.Workspace,
		run.Attempt,
		run.Error,
		run.OrchestratorID,
		formatTime(run.UpdatedAt),
	)
	if err != nil {
		return fmt.Errorf("enqueue outbox event: %w", err)
	}
	return nil
}

func (s *Store) UnsentOutboxEvents(ctx context.Context, limit int) ([]entity.OutboxEvent, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, event_id, run_id, issue_id, work_item_id, claim_token, status, workspace, attempt, error, orchestrator_id, occurred_at, sent_at
		FROM outbox_events
		WHERE sent_at = ''
		ORDER BY id ASC
		LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("list unsent outbox events: %w", err)
	}
	defer rows.Close()

	var events []entity.OutboxEvent
	for rows.Next() {
		event, err := scanOutboxEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *Store) MarkOutboxEventSent(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE outbox_events SET sent_at = ? WHERE id = ?`, nowString(), id)
	if err != nil {
		return fmt.Errorf("mark outbox event sent: %w", err)
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanOutboxEvent(row scanner) (entity.OutboxEvent, error) {
	var event entity.OutboxEvent
	var occurredAt string
	var sentAt string
	err := row.Scan(&event.ID, &event.EventID, &event.RunID, &event.IssueID, &event.WorkItemID, &event.ClaimToken, &event.Status, &event.Workspace, &event.Attempt, &event.Error, &event.OrchestratorID, &occurredAt, &sentAt)
	if err != nil {
		return entity.OutboxEvent{}, err
	}
	event.OccurredAt, err = parseTime(occurredAt)
	if err != nil {
		return entity.OutboxEvent{}, err
	}
	if sentAt != "" {
		parsed, err := parseTime(sentAt)
		if err != nil {
			return entity.OutboxEvent{}, err
		}
		event.SentAt = &parsed
	}
	return event, nil
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
