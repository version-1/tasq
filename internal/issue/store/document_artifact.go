package store

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func (s *Store) CreateDocumentArtifact(ctx context.Context, issueID int64, input entity.CreateDocumentArtifactInput) (entity.DocumentArtifact, error) {
	if issueID <= 0 {
		return entity.DocumentArtifact{}, fmt.Errorf("issue id is invalid")
	}
	if err := s.ensureIssueExists(ctx, issueID); err != nil {
		return entity.DocumentArtifact{}, err
	}
	normalized, err := entity.NormalizeCreateDocumentArtifactInput(input)
	if err != nil {
		return entity.DocumentArtifact{}, err
	}
	id, err := documentArtifactID()
	if err != nil {
		return entity.DocumentArtifact{}, err
	}
	now := nowString()
	if _, err := s.db.ExecContext(ctx, `INSERT INTO issue_document_artifacts (id, issue_id, type, title, body, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, id, issueID, normalized.Type, normalized.Title, normalized.Body, now, now); err != nil {
		return entity.DocumentArtifact{}, fmt.Errorf("create issue document artifact: %w", err)
	}
	return s.DocumentArtifact(ctx, issueID, id)
}

func (s *Store) DocumentArtifacts(ctx context.Context, issueID int64) ([]entity.DocumentArtifact, error) {
	if err := s.ensureIssueExists(ctx, issueID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, issue_id, type, title, created_at, updated_at FROM issue_document_artifacts WHERE issue_id = ? ORDER BY created_at ASC, id ASC`, issueID)
	if err != nil {
		return nil, fmt.Errorf("list issue document artifacts: %w", err)
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate issue document artifacts: %w", err)
	}
	return items, nil
}

func (s *Store) DocumentArtifact(ctx context.Context, issueID int64, id string) (entity.DocumentArtifact, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, issue_id, type, title, body, created_at, updated_at FROM issue_document_artifacts WHERE issue_id = ? AND id = ?`, issueID, id)
	item, err := scanDocumentArtifact(row, true)
	if err != nil {
		return entity.DocumentArtifact{}, err
	}
	return item, nil
}

func (s *Store) UpdateDocumentArtifact(ctx context.Context, issueID int64, id string, input entity.UpdateDocumentArtifactInput) (entity.DocumentArtifact, error) {
	normalized, err := entity.NormalizeUpdateDocumentArtifactInput(input)
	if err != nil {
		return entity.DocumentArtifact{}, err
	}
	assignments := []string{"updated_at = ?"}
	args := []any{nowString()}
	if normalized.Title != nil {
		assignments = append(assignments, "title = ?")
		args = append(args, *normalized.Title)
	}
	if normalized.Body != nil {
		assignments = append(assignments, "body = ?")
		args = append(args, *normalized.Body)
	}
	args = append(args, issueID, id)
	result, err := s.db.ExecContext(ctx, `UPDATE issue_document_artifacts SET `+strings.Join(assignments, ", ")+` WHERE issue_id = ? AND id = ?`, args...)
	if err != nil {
		return entity.DocumentArtifact{}, fmt.Errorf("update issue document artifact: %w", err)
	}
	if affected, err := result.RowsAffected(); err != nil {
		return entity.DocumentArtifact{}, err
	} else if affected == 0 {
		return entity.DocumentArtifact{}, sql.ErrNoRows
	}
	return s.DocumentArtifact(ctx, issueID, id)
}

type documentArtifactScanner interface{ Scan(...any) error }

func scanDocumentArtifact(row documentArtifactScanner, body bool) (entity.DocumentArtifact, error) {
	var item entity.DocumentArtifact
	var created, updated string
	var err error
	if body {
		err = row.Scan(&item.ID, &item.IssueID, &item.Type, &item.Title, &item.Body, &created, &updated)
	} else {
		err = row.Scan(&item.ID, &item.IssueID, &item.Type, &item.Title, &created, &updated)
	}
	if err != nil {
		return entity.DocumentArtifact{}, err
	}
	item.CreatedAt, err = parseTime(created)
	if err != nil {
		return entity.DocumentArtifact{}, err
	}
	item.UpdatedAt, err = parseTime(updated)
	if err != nil {
		return entity.DocumentArtifact{}, err
	}
	return item, nil
}
func documentArtifactID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate document artifact id: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
