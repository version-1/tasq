package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func (s *Store) CreateComment(ctx context.Context, input entity.CreateCommentInput) (entity.Comment, error) {
	normalized, err := entity.NormalizeCreateComment(input)
	if err != nil {
		return entity.Comment{}, err
	}
	if _, err := s.Issue(ctx, normalized.IssueID); err != nil {
		return entity.Comment{}, err
	}
	now := nowString()
	result, err := s.db.ExecContext(ctx, `INSERT INTO comments (
		issue_id, author, type, body, created_at
	) VALUES (?, ?, ?, ?, ?)`,
		normalized.IssueID,
		normalized.Author,
		normalized.Type,
		normalized.Body,
		now,
	)
	if err != nil {
		return entity.Comment{}, fmt.Errorf("create comment: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return entity.Comment{}, fmt.Errorf("read created comment id: %w", err)
	}
	return s.Comment(ctx, id)
}

func (s *Store) CommentsByIssueID(ctx context.Context, issueID int64, cursor int64, limit int) ([]entity.Comment, error) {
	if issueID <= 0 {
		return nil, errors.New("issueId is required")
	}
	if cursor < 0 {
		return nil, errors.New("cursor is invalid")
	}
	if _, err := s.Issue(ctx, issueID); err != nil {
		return nil, err
	}
	limit = normalizeCommentLimit(limit)
	rows, err := s.db.QueryContext(ctx, `SELECT `+commentColumns()+` FROM comments
		WHERE issue_id = ? AND id > ?
		ORDER BY id ASC
		LIMIT ?`, issueID, cursor, limit)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()

	comments := []entity.Comment{}
	for rows.Next() {
		item, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		comments = append(comments, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate comments: %w", err)
	}
	return comments, nil
}

func (s *Store) Comment(ctx context.Context, id int64) (entity.Comment, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+commentColumns()+` FROM comments WHERE id = ?`, id)
	item, err := scanComment(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Comment{}, sql.ErrNoRows
		}
		return entity.Comment{}, fmt.Errorf("read comment: %w", err)
	}
	return item, nil
}

func (s *Store) UpdateComment(ctx context.Context, id int64, input entity.UpdateCommentInput) (entity.Comment, error) {
	normalized, err := entity.NormalizeUpdateComment(input)
	if err != nil {
		return entity.Comment{}, err
	}
	current, err := s.Comment(ctx, id)
	if err != nil {
		return entity.Comment{}, err
	}
	if normalized.Body != nil {
		current.Body = *normalized.Body
	}
	_, err = s.db.ExecContext(ctx, `UPDATE comments SET body = ? WHERE id = ?`, current.Body, id)
	if err != nil {
		return entity.Comment{}, fmt.Errorf("update comment: %w", err)
	}
	return s.Comment(ctx, id)
}

func (s *Store) CreateChangeRequest(ctx context.Context, input entity.CreateChangeRequestInput) (entity.ChangeRequest, error) {
	normalized, err := entity.NormalizeCreateChangeRequest(input)
	if err != nil {
		return entity.ChangeRequest{}, err
	}
	if _, err := s.Issue(ctx, normalized.IssueID); err != nil {
		return entity.ChangeRequest{}, err
	}
	now := nowString()
	result, err := s.db.ExecContext(ctx, `INSERT INTO change_requests (
		issue_id, author, body, status, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?)`,
		normalized.IssueID,
		normalized.Author,
		normalized.Body,
		entity.ChangeRequestOpen,
		now,
		now,
	)
	if err != nil {
		return entity.ChangeRequest{}, fmt.Errorf("create change request: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return entity.ChangeRequest{}, fmt.Errorf("read created change request id: %w", err)
	}
	return s.ChangeRequest(ctx, id)
}

func (s *Store) ChangeRequestsByIssueID(ctx context.Context, issueID int64, status entity.ChangeRequestStatus, limit int) ([]entity.ChangeRequest, error) {
	if issueID <= 0 {
		return nil, errors.New("issueId is required")
	}
	if status != "" && !entity.IsValidChangeRequestStatus(status) {
		return nil, errors.New("status is invalid")
	}
	if _, err := s.Issue(ctx, issueID); err != nil {
		return nil, err
	}
	limit = normalizeChangeRequestLimit(limit)
	query := `SELECT ` + changeRequestColumns() + ` FROM change_requests WHERE issue_id = ?`
	args := []any{issueID}
	if status != "" {
		query += ` AND status = ?`
		args = append(args, status)
	}
	query += ` ORDER BY created_at ASC, id ASC LIMIT ?`
	args = append(args, limit)
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list change requests: %w", err)
	}
	defer rows.Close()

	items := []entity.ChangeRequest{}
	for rows.Next() {
		item, err := scanChangeRequest(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate change requests: %w", err)
	}
	return items, nil
}

func (s *Store) ChangeRequest(ctx context.Context, id int64) (entity.ChangeRequest, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+changeRequestColumns()+` FROM change_requests WHERE id = ?`, id)
	item, err := scanChangeRequest(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.ChangeRequest{}, sql.ErrNoRows
		}
		return entity.ChangeRequest{}, fmt.Errorf("read change request: %w", err)
	}
	return item, nil
}

func (s *Store) UpdateChangeRequest(ctx context.Context, id int64, input entity.UpdateChangeRequestInput) (entity.ChangeRequest, error) {
	current, err := s.ChangeRequest(ctx, id)
	if err != nil {
		return entity.ChangeRequest{}, err
	}
	normalized, err := entity.NormalizeUpdateChangeRequest(current.Status, input)
	if err != nil {
		return entity.ChangeRequest{}, err
	}
	if normalized.ResultCommentID != nil {
		comment, err := s.Comment(ctx, *normalized.ResultCommentID)
		if err != nil {
			return entity.ChangeRequest{}, err
		}
		if comment.IssueID != current.IssueID {
			return entity.ChangeRequest{}, errors.New("resultCommentId must reference the same issue")
		}
	}
	body := current.Body
	if normalized.Body != nil {
		body = *normalized.Body
	}
	status := current.Status
	if normalized.Status != nil {
		status = *normalized.Status
	}
	resolvedAt := current.ResolvedAt
	if current.Status != entity.ChangeRequestResolved && status == entity.ChangeRequestResolved {
		now, err := parseTime(nowString())
		if err != nil {
			return entity.ChangeRequest{}, err
		}
		resolvedAt = &now
	}
	resolvedByRunID := current.ResolvedByRunID
	if normalized.ResolvedByRunID != nil {
		resolvedByRunID = normalized.ResolvedByRunID
	}
	resultCommentID := current.ResultCommentID
	if normalized.ResultCommentID != nil {
		resultCommentID = normalized.ResultCommentID
	}
	now := nowString()
	_, err = s.db.ExecContext(ctx, `UPDATE change_requests
		SET body = ?, status = ?, updated_at = ?, resolved_at = ?, resolved_by_run_id = ?, result_comment_id = ?
		WHERE id = ?`,
		body,
		status,
		now,
		formatOptionalTime(resolvedAt),
		stringPtrValue(resolvedByRunID),
		int64PtrValue(resultCommentID),
		id,
	)
	if err != nil {
		return entity.ChangeRequest{}, fmt.Errorf("update change request: %w", err)
	}
	return s.ChangeRequest(ctx, id)
}

func (s *Store) CancelChangeRequest(ctx context.Context, id int64) (entity.ChangeRequest, error) {
	status := entity.ChangeRequestCanceled
	return s.UpdateChangeRequest(ctx, id, entity.UpdateChangeRequestInput{Status: &status})
}

func (s *Store) Summary(ctx context.Context) (entity.Summary, error) {
	issues, err := s.Issues(ctx)
	if err != nil {
		return entity.Summary{}, err
	}
	commentCounts, err := s.commentCountsByIssueID(ctx)
	if err != nil {
		return entity.Summary{}, err
	}
	dependencies, err := s.dependencyStatusesForParents(ctx, issueIDs(issues))
	if err != nil {
		return entity.Summary{}, err
	}
	columns := make([]entity.Column, 0, len(entity.OrderedStatuses()))
	for _, status := range entity.OrderedStatuses() {
		column := entity.Column{Status: status, Title: entity.StatusTitle(status), Issues: []entity.IssueSummary{}}
		for _, item := range issues {
			if item.Status != status {
				continue
			}
			blockingDependencies := blockingDependencyIDs(dependencies[item.ID])
			column.Issues = append(column.Issues, entity.IssueSummary{
				Issue:       item,
				QueueStatus: entity.IssueQueueStatus(item.Status, len(blockingDependencies) > 0),
				Stats: entity.IssueStats{
					CommentCount: commentCounts[item.ID],
				},
			})
		}
		columns = append(columns, column)
	}
	return entity.Summary{Columns: columns, GeneratedAt: time.Now().UTC()}, nil
}

func (s *Store) commentCountsByIssueID(ctx context.Context) (map[int64]int, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT issue_id, COUNT(*) FROM comments GROUP BY issue_id`)
	if err != nil {
		return nil, fmt.Errorf("count issue comments: %w", err)
	}
	defer rows.Close()

	counts := map[int64]int{}
	for rows.Next() {
		var issueID int64
		var count int
		if err := rows.Scan(&issueID, &count); err != nil {
			return nil, fmt.Errorf("scan issue comment count: %w", err)
		}
		counts[issueID] = count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate issue comment counts: %w", err)
	}
	return counts, nil
}

func normalizeCommentLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	if limit > 100 {
		return 100
	}
	return limit
}

func normalizeChangeRequestLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	if limit > 100 {
		return 100
	}
	return limit
}
