package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

type IssueFilter struct {
	States        []entity.Status
	ProjectID     *int64
	ProjectIDs    []int64
	Priorities    []entity.Priority
	Assignee      *string
	Search        string
	Limit         int
	Offset        int
	SortBy        IssueSortBy
	SortDirection SortDirection
}

type IssueList struct {
	Issues []entity.Issue
	Total  int
}

type IssueSortBy string

type SortDirection string

func (s *Store) CreateIssue(ctx context.Context, input entity.CreateIssueInput) (entity.Issue, error) {
	normalized, err := entity.NormalizeCreate(input)
	if err != nil {
		return entity.Issue{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return entity.Issue{}, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()
	if err := projectExistsTx(ctx, tx, normalized.ProjectID); err != nil {
		return entity.Issue{}, err
	}
	now := nowString()
	result, err := tx.ExecContext(ctx, `INSERT INTO issues (
		project_id, title, description, status, priority, assignee, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		normalized.ProjectID,
		normalized.Title,
		normalized.Description,
		normalized.Status,
		normalized.Priority,
		normalized.Assignee,
		now,
		now,
	)
	if err != nil {
		return entity.Issue{}, fmt.Errorf("create issue: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return entity.Issue{}, fmt.Errorf("read created issue id: %w", err)
	}
	if err := setDependencyIDsTx(ctx, tx, id, normalized.DependencyIDs); err != nil {
		return entity.Issue{}, err
	}
	if err := tx.Commit(); err != nil {
		return entity.Issue{}, err
	}
	tx = nil
	return s.Issue(ctx, id)
}

func (s *Store) Issues(ctx context.Context) ([]entity.Issue, error) {
	return s.IssuesByFilter(ctx, IssueFilter{})
}

func (s *Store) IssuesByStates(ctx context.Context, states []entity.Status) ([]entity.Issue, error) {
	return s.IssuesByFilter(ctx, IssueFilter{States: states})
}

func (s *Store) IssuesByFilter(ctx context.Context, filter IssueFilter) ([]entity.Issue, error) {
	list, err := s.IssuesPageByFilter(ctx, filter)
	if err != nil {
		return nil, err
	}
	return list.Issues, nil
}

func (s *Store) IssuesPageByFilter(ctx context.Context, filter IssueFilter) (IssueList, error) {
	orderClause, err := issueOrderClause(filter)
	if err != nil {
		return IssueList{}, err
	}
	return s.issuesByFilterOrder(ctx, filter, orderClause)
}

func issueOrderClause(filter IssueFilter) (string, error) {
	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = IssueSortByUpdatedAt
	}
	direction := filter.SortDirection
	if direction == "" {
		direction = SortDirectionDesc
	}

	var column string
	switch sortBy {
	case IssueSortByID:
		column = "issues.id"
	case IssueSortByPriority:
		column = issuePriorityOrderExpression()
	case IssueSortByCreatedAt:
		column = "issues.created_at"
	case IssueSortByUpdatedAt:
		column = "issues.updated_at"
	default:
		return "", errors.New("sortBy is invalid")
	}

	var sqlDirection string
	switch direction {
	case SortDirectionAsc:
		sqlDirection = "ASC"
	case SortDirectionDesc:
		sqlDirection = "DESC"
	default:
		return "", errors.New("sortDirection is invalid")
	}

	if sortBy == IssueSortByID {
		return column + " " + sqlDirection, nil
	}
	return column + " " + sqlDirection + ", issues.id " + sqlDirection, nil
}

func issuePriorityOrderExpression() string {
	return `CASE issues.priority
		WHEN 'urgent' THEN 4
		WHEN 'high' THEN 3
		WHEN 'normal' THEN 2
		WHEN 'low' THEN 1
		ELSE 0
	END`
}

func (s *Store) issuesByFilterOrder(ctx context.Context, filter IssueFilter, orderClause string) (IssueList, error) {
	clauses := []string{}
	args := []any{}
	if len(filter.States) > 0 {
		for _, status := range filter.States {
			if !entity.IsValidStatus(status) {
				return IssueList{}, errors.New("status is invalid")
			}
			args = append(args, status)
		}
		clauses = append(clauses, `issues.status IN (`+placeholders(len(filter.States))+`)`)
	}
	if filter.ProjectID != nil {
		if *filter.ProjectID <= 0 {
			return IssueList{}, errors.New("projectId is invalid")
		}
		args = append(args, *filter.ProjectID)
		clauses = append(clauses, `issues.project_id = ?`)
	}
	if len(filter.ProjectIDs) > 0 {
		for _, projectID := range filter.ProjectIDs {
			if projectID <= 0 {
				return IssueList{}, errors.New("projectIds is invalid")
			}
			args = append(args, projectID)
		}
		clauses = append(clauses, `issues.project_id IN (`+placeholders(len(filter.ProjectIDs))+`)`)
	}
	if len(filter.Priorities) > 0 {
		for _, priority := range filter.Priorities {
			if !entity.IsValidPriority(priority) {
				return IssueList{}, errors.New("priority is invalid")
			}
			args = append(args, priority)
		}
		clauses = append(clauses, `issues.priority IN (`+placeholders(len(filter.Priorities))+`)`)
	}
	if filter.Assignee != nil {
		args = append(args, *filter.Assignee)
		clauses = append(clauses, `issues.assignee = ?`)
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		titlePattern := "%" + escapeLikePattern(strings.ToLower(search)) + "%"
		if id, err := strconv.ParseInt(search, 10, 64); err == nil && id > 0 {
			args = append(args, id, titlePattern)
			clauses = append(clauses, `(issues.id = ? OR lower(issues.title) LIKE ? ESCAPE '\')`)
		} else {
			args = append(args, titlePattern)
			clauses = append(clauses, `lower(issues.title) LIKE ? ESCAPE '\'`)
		}
	}
	if filter.Limit < 0 {
		return IssueList{}, errors.New("limit is invalid")
	}
	if filter.Offset < 0 {
		return IssueList{}, errors.New("offset is invalid")
	}
	where := ""
	if len(clauses) > 0 {
		where = " WHERE " + strings.Join(clauses, " AND ")
	}
	var total int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM issues JOIN projects ON projects.id = issues.project_id`+where, args...).Scan(&total); err != nil {
		return IssueList{}, fmt.Errorf("count issues: %w", err)
	}
	query := `SELECT ` + issueColumns() + ` FROM issues JOIN projects ON projects.id = issues.project_id` + where + ` ORDER BY ` + orderClause
	queryArgs := append([]any{}, args...)
	if filter.Limit > 0 {
		query += ` LIMIT ? OFFSET ?`
		queryArgs = append(queryArgs, filter.Limit, filter.Offset)
	}
	rows, err := s.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return IssueList{}, fmt.Errorf("list issues: %w", err)
	}
	defer rows.Close()

	var issues []entity.Issue
	for rows.Next() {
		item, err := scanIssue(rows)
		if err != nil {
			return IssueList{}, err
		}
		issues = append(issues, item)
	}
	if err := rows.Err(); err != nil {
		return IssueList{}, fmt.Errorf("iterate issues: %w", err)
	}
	if err := s.hydrateIssueDependencyIDs(ctx, issues); err != nil {
		return IssueList{}, err
	}
	return IssueList{Issues: issues, Total: total}, nil
}

func escapeLikePattern(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(value)
}

func (s *Store) IssueStatesByIDs(ctx context.Context, ids []int64) ([]entity.IssueState, error) {
	if len(ids) == 0 {
		return []entity.IssueState{}, nil
	}
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id, status FROM issues WHERE id IN (`+placeholders(len(ids))+`) ORDER BY id ASC`, args...)
	if err != nil {
		return nil, fmt.Errorf("list issue states by ids: %w", err)
	}
	defer rows.Close()

	states := []entity.IssueState{}
	for rows.Next() {
		var state entity.IssueState
		if err := rows.Scan(&state.ID, &state.Status); err != nil {
			return nil, err
		}
		states = append(states, state)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate issue states by ids: %w", err)
	}
	return states, nil
}

func (s *Store) Issue(ctx context.Context, id int64) (entity.Issue, error) {
	row := s.db.QueryRowContext(ctx, `SELECT `+issueColumns()+` FROM issues JOIN projects ON projects.id = issues.project_id WHERE issues.id = ?`, id)
	item, err := scanIssue(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Issue{}, sql.ErrNoRows
		}
		return entity.Issue{}, fmt.Errorf("read issue: %w", err)
	}
	dependencyIDs, err := s.DependencyIDs(ctx, id)
	if err != nil {
		return entity.Issue{}, err
	}
	item.DependencyIDs = dependencyIDs
	return item, nil
}

func (s *Store) UpdateIssue(ctx context.Context, id int64, input entity.UpdateIssueInput) (entity.Issue, error) {
	normalized, err := entity.NormalizeUpdateIssue(input)
	if err != nil {
		return entity.Issue{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return entity.Issue{}, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()
	current, err := issueByIDTx(ctx, tx, id)
	if err != nil {
		return entity.Issue{}, err
	}
	if normalized.Title != nil {
		current.Title = *normalized.Title
	}
	if normalized.Description != nil {
		current.Description = *normalized.Description
	}
	if normalized.Status != nil {
		current.Status = *normalized.Status
	}
	if normalized.Priority != nil {
		current.Priority = *normalized.Priority
	}
	if normalized.Assignee != nil {
		current.Assignee = *normalized.Assignee
	}
	_, err = tx.ExecContext(ctx, `UPDATE issues SET
		title = ?, description = ?, status = ?, priority = ?, assignee = ?, updated_at = ?
		WHERE id = ?`,
		current.Title,
		current.Description,
		current.Status,
		current.Priority,
		current.Assignee,
		nowString(),
		id,
	)
	if err != nil {
		return entity.Issue{}, fmt.Errorf("update issue: %w", err)
	}
	if normalized.DependencyIDs != nil {
		if err := setDependencyIDsTx(ctx, tx, id, *normalized.DependencyIDs); err != nil {
			return entity.Issue{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return entity.Issue{}, err
	}
	tx = nil
	return s.Issue(ctx, id)
}
