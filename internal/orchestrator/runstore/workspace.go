package runstore

import (
	"context"
	"fmt"
)

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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin workspace cleanup transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `UPDATE workspace_metadata
			SET cleanup_status = ?, cleanup_at = ?, last_error = ?, updated_at = ?
			WHERE workspace_key = ?`, status, now, errText, now, workspaceKey)
	if err != nil {
		return fmt.Errorf("mark workspace cleanup: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE runs
			SET thread_id = NULL, updated_at = ?
			WHERE issue_id = (
				SELECT issue_id FROM workspace_metadata WHERE workspace_key = ?
			)`, now, workspaceKey); err != nil {
		return fmt.Errorf("invalidate workspace resume thread ids: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit workspace cleanup transaction: %w", err)
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
