package runstore

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/version-1/tasq/internal/orchestrator/run"
)

func (s *Store) DeleteProjectIssueData(ctx context.Context, issueIDs []int64) error {
	if len(issueIDs) == 0 {
		return nil
	}
	placeholders := queryPlaceholders(len(issueIDs))
	args := int64Args(issueIDs)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var runningCount int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM runs
		WHERE issue_id IN (`+placeholders+`)
			AND status = ?`, append(args, string(run.StatusRunning))...).Scan(&runningCount); err != nil {
		return fmt.Errorf("count running project runs: %w", err)
	}
	if runningCount > 0 {
		return fmt.Errorf("%w: %d running run(s)", ErrProjectHasRunningRuns, runningCount)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM runner_events
		WHERE run_id IN (
			SELECT run_id FROM runs WHERE issue_id IN (`+placeholders+`)
		)`, args...); err != nil {
		return fmt.Errorf("delete project runner events: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM workspace_setup_failures
		WHERE issue_id IN (`+placeholders+`)`, args...); err != nil {
		return fmt.Errorf("delete project workspace setup failures: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM workspace_metadata
		WHERE issue_id IN (`+placeholders+`)`, args...); err != nil {
		return fmt.Errorf("delete project workspace metadata: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM runs
		WHERE issue_id IN (`+placeholders+`)`, args...); err != nil {
		return fmt.Errorf("delete project runs: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit project run cleanup: %w", err)
	}
	return nil
}

func scanRun(row scanner) (run.Run, error) {
	var storedRun run.Run
	var createdAt string
	var updatedAt string
	var threadID sql.NullString
	if err := row.Scan(&storedRun.ID, &storedRun.RunID, &storedRun.IssueID, &storedRun.Status, &storedRun.Workspace, &threadID, &storedRun.Attempt, &storedRun.Error, &storedRun.OrchestratorID, &createdAt, &updatedAt); err != nil {
		return run.Run{}, err
	}
	if threadID.Valid {
		storedRun.ThreadID = threadID.String
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

func nullString(value string) sql.NullString {
	return sql.NullString{String: value, Valid: value != ""}
}

func queryPlaceholders(count int) string {
	return strings.TrimRight(strings.Repeat("?,", count), ",")
}

func int64Args(values []int64) []any {
	args := make([]any, 0, len(values))
	for _, value := range values {
		args = append(args, value)
	}
	return args
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
