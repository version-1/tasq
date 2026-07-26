package runstore

import (
	"context"
	"fmt"
	"time"

	"github.com/version-1/tasq/internal/orchestrator/run"
)

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

func (s *Store) ConversationEvents(ctx context.Context, runID string) ([]run.RunnerEvent, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, run_id, event_type, message, payload_json, occurred_at
		FROM runner_events
		WHERE run_id = ?
			AND event_type IN (
				'running',
				'succeeded',
				'failed',
				'item/completed',
				'thread/tokenUsage/updated',
				'account/rateLimits/updated'
			)
		ORDER BY id ASC`, runID)
	if err != nil {
		return nil, fmt.Errorf("list conversation events: %w", err)
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
