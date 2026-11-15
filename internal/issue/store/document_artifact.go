package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/version-1/tasq/internal/issue/domain/entity"
	"strings"
)

var (
	ErrProjectHasExternalDocumentArtifactLinks = errors.New("project document artifact has external issue links")
	ErrDocumentArtifactStillLinked             = errors.New("document artifact is still linked")
)

func (s *Store) CreateDocumentArtifact(ctx context.Context, projectID int64, input entity.CreateDocumentArtifactInput) (entity.DocumentArtifact, error) {
	if _, err := s.Project(ctx, projectID); err != nil {
		return entity.DocumentArtifact{}, err
	}
	input, err := entity.NormalizeCreateDocumentArtifactInput(input)
	if err != nil {
		return entity.DocumentArtifact{}, err
	}
	id, err := documentArtifactID()
	if err != nil {
		return entity.DocumentArtifact{}, err
	}
	now := nowString()
	if _, err := s.db.ExecContext(ctx, `INSERT INTO document_artifacts (id, project_id, type, title, body, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, id, projectID, input.Type, input.Title, input.Body, now, now); err != nil {
		return entity.DocumentArtifact{}, fmt.Errorf("create document artifact: %w", err)
	}
	return s.DocumentArtifact(ctx, projectID, id)
}
func (s *Store) DocumentArtifacts(ctx context.Context, projectID int64) ([]entity.DocumentArtifact, error) {
	if _, err := s.Project(ctx, projectID); err != nil {
		return nil, err
	}
	return s.documentArtifacts(ctx, `SELECT id, project_id, type, title, created_at, updated_at FROM document_artifacts WHERE project_id = ? ORDER BY created_at, id`, projectID)
}
func (s *Store) DocumentArtifactsForIssue(ctx context.Context, issueID int64) ([]entity.DocumentArtifact, error) {
	if err := s.ensureIssueExists(ctx, issueID); err != nil {
		return nil, err
	}
	return s.documentArtifacts(ctx, `SELECT a.id, a.project_id, a.type, a.title, a.created_at, a.updated_at FROM document_artifacts a JOIN document_artifact_links l ON l.document_artifact_id = a.id WHERE l.target_type = 'issue' AND l.target_id = ? ORDER BY a.created_at, a.id`, issueID)
}
func (s *Store) documentArtifacts(ctx context.Context, query string, arg int64) ([]entity.DocumentArtifact, error) {
	rows, err := s.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("list document artifacts: %w", err)
	}
	defer rows.Close()
	items := []entity.DocumentArtifact{}
	for rows.Next() {
		item, err := scanDocumentArtifact(rows, false)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
func (s *Store) DocumentArtifact(ctx context.Context, projectID int64, id string) (entity.DocumentArtifact, error) {
	return scanDocumentArtifact(s.db.QueryRowContext(ctx, `SELECT id, project_id, type, title, body, created_at, updated_at FROM document_artifacts WHERE project_id = ? AND id = ?`, projectID, id), true)
}
func (s *Store) UpdateDocumentArtifact(ctx context.Context, projectID int64, id string, input entity.UpdateDocumentArtifactInput) (entity.DocumentArtifact, error) {
	input, err := entity.NormalizeUpdateDocumentArtifactInput(input)
	if err != nil {
		return entity.DocumentArtifact{}, err
	}
	assignments, args := []string{"updated_at = ?"}, []any{nowString()}
	if input.Title != nil {
		assignments = append(assignments, "title = ?")
		args = append(args, *input.Title)
	}
	if input.Body != nil {
		assignments = append(assignments, "body = ?")
		args = append(args, *input.Body)
	}
	args = append(args, projectID, id)
	result, err := s.db.ExecContext(ctx, `UPDATE document_artifacts SET `+strings.Join(assignments, ", ")+` WHERE project_id = ? AND id = ?`, args...)
	if err != nil {
		return entity.DocumentArtifact{}, fmt.Errorf("update document artifact: %w", err)
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return entity.DocumentArtifact{}, sql.ErrNoRows
	}
	return s.DocumentArtifact(ctx, projectID, id)
}
func (s *Store) DeleteDocumentArtifact(ctx context.Context, projectID int64, id string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE document_artifacts SET id = id WHERE project_id = ? AND id = ?`, projectID, id)
	if err != nil {
		return fmt.Errorf("lock document artifact: %w", err)
	}
	if n, err := result.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return sql.ErrNoRows
	}
	var linked bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM document_artifact_links WHERE document_artifact_id = ?)`, id).Scan(&linked); err != nil {
		return fmt.Errorf("check document artifact links: %w", err)
	}
	if linked {
		return ErrDocumentArtifactStillLinked
	}
	result, err = tx.ExecContext(ctx, `DELETE FROM document_artifacts WHERE project_id = ? AND id = ?`, projectID, id)
	if err != nil {
		return fmt.Errorf("delete document artifact: %w", err)
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}
func (s *Store) LinkDocumentArtifactToIssue(ctx context.Context, projectID int64, id string, issueID int64) error {
	if _, err := s.DocumentArtifact(ctx, projectID, id); err != nil {
		return err
	}
	if err := s.ensureIssueExists(ctx, issueID); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO document_artifact_links (document_artifact_id,target_type,target_id,created_at) VALUES (?,'issue',?,?)`, id, issueID, nowString())
	return err
}
func (s *Store) UnlinkDocumentArtifactFromIssue(ctx context.Context, projectID int64, id string, issueID int64) error {
	if _, err := s.DocumentArtifact(ctx, projectID, id); err != nil {
		return err
	}
	result, err := s.db.ExecContext(ctx, `DELETE FROM document_artifact_links WHERE document_artifact_id=? AND target_type='issue' AND target_id=?`, id, issueID)
	if err != nil {
		return err
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) ProjectHasExternalDocumentArtifactLinks(ctx context.Context, projectID int64) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM document_artifacts a JOIN document_artifact_links l ON l.document_artifact_id = a.id JOIN issues i ON i.id = l.target_id WHERE a.project_id = ? AND i.project_id <> ?)`, projectID, projectID).Scan(&exists)
	return exists, err
}

type documentArtifactScanner interface{ Scan(...any) error }

func scanDocumentArtifact(row documentArtifactScanner, body bool) (entity.DocumentArtifact, error) {
	var item entity.DocumentArtifact
	var created, updated string
	var err error
	if body {
		err = row.Scan(&item.ID, &item.ProjectID, &item.Type, &item.Title, &item.Body, &created, &updated)
	} else {
		err = row.Scan(&item.ID, &item.ProjectID, &item.Type, &item.Title, &created, &updated)
	}
	if err != nil {
		return entity.DocumentArtifact{}, err
	}
	item.CreatedAt, err = parseTime(created)
	if err != nil {
		return entity.DocumentArtifact{}, err
	}
	item.UpdatedAt, err = parseTime(updated)
	return item, err
}
func documentArtifactID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate document artifact id: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
