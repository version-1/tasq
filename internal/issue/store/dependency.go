package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

type dependencyEdge struct {
	ParentID     int64
	DependencyID int64
}

type DependencyCycleError struct {
	IssueIDs []int64
}

func (e DependencyCycleError) Error() string {
	parts := make([]string, 0, len(e.IssueIDs))
	for _, id := range e.IssueIDs {
		parts = append(parts, fmt.Sprintf("%d", id))
	}
	return "dependency cycle detected: " + strings.Join(parts, " -> ")
}

func (s *Store) DependencyIDs(ctx context.Context, issueID int64) ([]int64, error) {
	if issueID <= 0 {
		return nil, errors.New("issueId is required")
	}
	rows, err := s.db.QueryContext(ctx, `SELECT dependency_issue_id FROM issue_dependencies WHERE parent_issue_id = ? ORDER BY dependency_issue_id ASC`, issueID)
	if err != nil {
		return nil, fmt.Errorf("list issue dependencies: %w", err)
	}
	defer rows.Close()
	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan issue dependency: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate issue dependencies: %w", err)
	}
	return ids, nil
}

func (s *Store) SetDependencyIDs(ctx context.Context, issueID int64, dependencyIDs []int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()
	if err := setDependencyIDsTx(ctx, tx, issueID, dependencyIDs); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	tx = nil
	return nil
}

func setDependencyIDsTx(ctx context.Context, tx *sql.Tx, issueID int64, dependencyIDs []int64) error {
	if issueID <= 0 {
		return errors.New("issueId is required")
	}
	if _, err := issueByIDTx(ctx, tx, issueID); err != nil {
		return err
	}
	seen := map[int64]struct{}{}
	for _, dependencyID := range dependencyIDs {
		if dependencyID <= 0 {
			return errors.New("dependency_ids contains invalid issue id")
		}
		if dependencyID == issueID {
			return DependencyCycleError{IssueIDs: []int64{issueID, issueID}}
		}
		if _, ok := seen[dependencyID]; ok {
			return errors.New("dependency_ids contains duplicate issue id")
		}
		seen[dependencyID] = struct{}{}
		if _, err := issueByIDTx(ctx, tx, dependencyID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("dependency issue %d not found", dependencyID)
			}
			return err
		}
	}
	edges, err := dependencyEdgesTx(ctx, tx)
	if err != nil {
		return err
	}
	edges = replaceDependencyEdges(edges, issueID, dependencyIDs)
	if cycle := dependencyCycle(edges, issueID); len(cycle) > 0 {
		return DependencyCycleError{IssueIDs: cycle}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM issue_dependencies WHERE parent_issue_id = ?`, issueID); err != nil {
		return fmt.Errorf("delete issue dependencies: %w", err)
	}
	now := nowString()
	for _, dependencyID := range dependencyIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO issue_dependencies (parent_issue_id, dependency_issue_id, created_at) VALUES (?, ?, ?)`, issueID, dependencyID, now); err != nil {
			return fmt.Errorf("insert issue dependency: %w", err)
		}
	}
	return nil
}

func issueByIDTx(ctx context.Context, tx *sql.Tx, id int64) (entity.Issue, error) {
	row := tx.QueryRowContext(ctx, `SELECT `+issueColumns()+` FROM issues JOIN projects ON projects.id = issues.project_id WHERE issues.id = ?`, id)
	item, err := scanIssue(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Issue{}, sql.ErrNoRows
		}
		return entity.Issue{}, fmt.Errorf("read issue: %w", err)
	}
	return item, nil
}

func dependencyEdgesTx(ctx context.Context, tx *sql.Tx) ([]dependencyEdge, error) {
	rows, err := tx.QueryContext(ctx, `SELECT parent_issue_id, dependency_issue_id FROM issue_dependencies`)
	if err != nil {
		return nil, fmt.Errorf("list dependency edges: %w", err)
	}
	defer rows.Close()
	edges := []dependencyEdge{}
	for rows.Next() {
		var edge dependencyEdge
		if err := rows.Scan(&edge.ParentID, &edge.DependencyID); err != nil {
			return nil, fmt.Errorf("scan dependency edge: %w", err)
		}
		edges = append(edges, edge)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dependency edges: %w", err)
	}
	return edges, nil
}

func replaceDependencyEdges(edges []dependencyEdge, parentID int64, dependencyIDs []int64) []dependencyEdge {
	replaced := make([]dependencyEdge, 0, len(edges)+len(dependencyIDs))
	for _, edge := range edges {
		if edge.ParentID != parentID {
			replaced = append(replaced, edge)
		}
	}
	for _, dependencyID := range dependencyIDs {
		replaced = append(replaced, dependencyEdge{ParentID: parentID, DependencyID: dependencyID})
	}
	return replaced
}

func dependencyCycle(edges []dependencyEdge, startID int64) []int64 {
	graph := map[int64][]int64{}
	for _, edge := range edges {
		graph[edge.ParentID] = append(graph[edge.ParentID], edge.DependencyID)
	}
	path := []int64{startID}
	visiting := map[int64]struct{}{startID: {}}
	var visit func(id int64) []int64
	visit = func(id int64) []int64 {
		for _, next := range graph[id] {
			if next == startID {
				return append(append([]int64{}, path...), startID)
			}
			if _, ok := visiting[next]; ok {
				continue
			}
			visiting[next] = struct{}{}
			path = append(path, next)
			if cycle := visit(next); len(cycle) > 0 {
				return cycle
			}
			path = path[:len(path)-1]
			delete(visiting, next)
		}
		return nil
	}
	return visit(startID)
}

func (s *Store) hydrateIssueDependencyIDs(ctx context.Context, issues []entity.Issue) error {
	if len(issues) == 0 {
		return nil
	}
	ids := issueIDs(issues)
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT parent_issue_id, dependency_issue_id FROM issue_dependencies WHERE parent_issue_id IN (`+placeholders(len(ids))+`) ORDER BY parent_issue_id ASC, dependency_issue_id ASC`, args...)
	if err != nil {
		return fmt.Errorf("list issue dependencies: %w", err)
	}
	defer rows.Close()
	byParent := map[int64][]int64{}
	for rows.Next() {
		var parentID int64
		var dependencyID int64
		if err := rows.Scan(&parentID, &dependencyID); err != nil {
			return fmt.Errorf("scan issue dependency: %w", err)
		}
		byParent[parentID] = append(byParent[parentID], dependencyID)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate issue dependencies: %w", err)
	}
	for i := range issues {
		issues[i].DependencyIDs = append([]int64{}, byParent[issues[i].ID]...)
		if issues[i].DependencyIDs == nil {
			issues[i].DependencyIDs = []int64{}
		}
	}
	return nil
}
