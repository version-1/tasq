package runstore

import (
	"context"
	"fmt"

	"github.com/version-1/tasq/internal/orchestrator/run"
)

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
		run_id, issue_id, status, workspace, thread_id, attempt, orchestrator_id, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		runID,
		input.IssueID,
		run.StatusQueued,
		input.Workspace,
		nullString(input.ThreadID),
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
	ThreadID       string
	Attempt        int
	OrchestratorID string
}

func (s *Store) UpdateRunThreadID(ctx context.Context, runID string, threadID string) (run.Run, error) {
	if runID == "" {
		return run.Run{}, fmt.Errorf("runId is required")
	}
	if runeCount(threadID) > maxRunThreadIDLength {
		return run.Run{}, fmt.Errorf("threadId must be 200 characters or fewer")
	}
	now := nowString()
	_, err := s.db.ExecContext(ctx, `UPDATE runs SET thread_id = ?, updated_at = ? WHERE run_id = ?`, nullString(threadID), now, runID)
	if err != nil {
		return run.Run{}, fmt.Errorf("update run thread id: %w", err)
	}
	updatedRun, err := s.RunByRunID(ctx, runID)
	if err != nil {
		return run.Run{}, err
	}
	return updatedRun, nil
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

func (s *Store) RunByRunID(ctx context.Context, runID string) (run.Run, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, run_id, issue_id, status, workspace, thread_id, attempt, error, orchestrator_id, created_at, updated_at
		FROM runs WHERE run_id = ?`, runID)
	return scanRun(row)
}

func (s *Store) ActiveRuns(ctx context.Context) ([]run.Run, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, run_id, issue_id, status, workspace, thread_id, attempt, error, orchestrator_id, created_at, updated_at
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
	row := s.db.QueryRowContext(ctx, `SELECT id, run_id, issue_id, status, workspace, thread_id, attempt, error, orchestrator_id, created_at, updated_at
		FROM runs
		WHERE issue_id = ?
		ORDER BY id DESC
		LIMIT 1`, issueID)
	return scanRun(row)
}

func (s *Store) RunsByIssueID(ctx context.Context, issueID int64) ([]run.Run, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, run_id, issue_id, status, workspace, thread_id, attempt, error, orchestrator_id, created_at, updated_at
		FROM runs
		WHERE issue_id = ?
		ORDER BY id DESC`, issueID)
	if err != nil {
		return nil, fmt.Errorf("list runs by issue id: %w", err)
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

func (s *Store) LatestResumeThreadIDByIssueID(ctx context.Context, issueID int64) (string, error) {
	var threadID string
	err := s.db.QueryRowContext(ctx, `SELECT thread_id
		FROM runs
		WHERE issue_id = ?
			AND thread_id IS NOT NULL
			AND thread_id <> ''
		ORDER BY id DESC
		LIMIT 1`, issueID).Scan(&threadID)
	if err != nil {
		return "", err
	}
	return threadID, nil
}
