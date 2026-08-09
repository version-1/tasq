package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func (s *Store) UpsertArtifact(ctx context.Context, issueID int64, artifactType entity.ArtifactType, input entity.UpsertArtifactInput) (entity.Artifact, error) {
	if issueID <= 0 {
		return entity.Artifact{}, fmt.Errorf("issue id is invalid")
	}
	if err := s.ensureIssueExists(ctx, issueID); err != nil {
		return entity.Artifact{}, err
	}
	artifact, err := entity.NormalizeArtifactInput(artifactType, input)
	if err != nil {
		return entity.Artifact{}, err
	}
	now := nowString()
	_, err = s.db.ExecContext(ctx, `INSERT INTO issue_artifacts (issue_id, type, data_type, data_value, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT(issue_id, type) DO UPDATE SET data_type = excluded.data_type, data_value = excluded.data_value, updated_at = excluded.updated_at`, issueID, artifact.Type, artifact.DataType, artifact.DataValue, now, now)
	if err != nil {
		return entity.Artifact{}, fmt.Errorf("upsert issue artifact: %w", err)
	}
	return artifact, nil
}

func (s *Store) DeleteArtifact(ctx context.Context, issueID int64, artifactType entity.ArtifactType) error {
	if issueID <= 0 {
		return fmt.Errorf("issue id is invalid")
	}
	if artifactType != entity.ArtifactTypePullRequest {
		return fmt.Errorf("artifact type is unsupported")
	}
	if err := s.ensureIssueExists(ctx, issueID); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM issue_artifacts WHERE issue_id = ? AND type = ?`, issueID, artifactType)
	if err != nil {
		return fmt.Errorf("delete issue artifact: %w", err)
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

func (s *Store) ensureIssueExists(ctx context.Context, issueID int64) error {
	var found int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM issues WHERE id = ?`, issueID).Scan(&found)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) hydrateIssueArtifacts(ctx context.Context, issues []entity.Issue) error {
	if len(issues) == 0 {
		return nil
	}
	ids := issueIDs(issues)
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	rows, err := s.db.QueryContext(ctx, `SELECT issue_id, type, data_type, data_value FROM issue_artifacts WHERE issue_id IN (`+placeholders(len(ids))+`) ORDER BY issue_id ASC, type ASC`, args...)
	if err != nil {
		return fmt.Errorf("list issue artifacts: %w", err)
	}
	defer rows.Close()
	byIssue := map[int64][]entity.Artifact{}
	for rows.Next() {
		var issueID int64
		var item entity.Artifact
		if err := rows.Scan(&issueID, &item.Type, &item.DataType, &item.DataValue); err != nil {
			return fmt.Errorf("scan issue artifact: %w", err)
		}
		byIssue[issueID] = append(byIssue[issueID], item)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate issue artifacts: %w", err)
	}
	for i := range issues {
		issues[i].Artifacts = append([]entity.Artifact{}, byIssue[issues[i].ID]...)
		if issues[i].Artifacts == nil {
			issues[i].Artifacts = []entity.Artifact{}
		}
	}
	return nil
}

func (s *Store) artifactsForIssue(ctx context.Context, issueID int64) ([]entity.Artifact, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT type, data_type, data_value FROM issue_artifacts WHERE issue_id = ? ORDER BY type ASC`, issueID)
	if err != nil {
		return nil, fmt.Errorf("list issue artifacts: %w", err)
	}
	defer rows.Close()
	artifacts := []entity.Artifact{}
	for rows.Next() {
		var item entity.Artifact
		if err := rows.Scan(&item.Type, &item.DataType, &item.DataValue); err != nil {
			return nil, fmt.Errorf("scan issue artifact: %w", err)
		}
		artifacts = append(artifacts, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate issue artifacts: %w", err)
	}
	return artifacts, nil
}
