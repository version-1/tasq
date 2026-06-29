package store

import (
	"context"
	"fmt"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

type dependencyStatus struct {
	DependencyID int64
	Status       entity.Status
}

func (s *Store) Queue(ctx context.Context, filter IssueFilter) (entity.Queue, error) {
	ready := entity.StatusReady
	filter.States = []entity.Status{ready}
	list, err := s.issuesByFilterOrder(ctx, filter, queueIssueOrderClause())
	if err != nil {
		return entity.Queue{}, err
	}
	issues := list.Issues
	dependencies, err := s.dependencyStatusesForParents(ctx, issueIDs(issues))
	if err != nil {
		return entity.Queue{}, err
	}
	queue := entity.Queue{
		Queued:  []entity.QueueIssue{},
		Pending: []entity.QueueIssue{},
	}
	for _, item := range issues {
		blocked := blockingDependencyIDs(dependencies[item.ID])
		if len(blocked) > 0 {
			queue.Pending = append(queue.Pending, entity.QueueIssue{Issue: item, BlockedDependencyIDs: blocked})
			continue
		}
		queue.Queued = append(queue.Queued, entity.QueueIssue{Issue: item})
	}
	return queue, nil
}

func (s *Store) dependencyStatusesForParents(ctx context.Context, parentIDs []int64) (map[int64][]dependencyStatus, error) {
	statuses := map[int64][]dependencyStatus{}
	if len(parentIDs) == 0 {
		return statuses, nil
	}
	args := make([]any, 0, len(parentIDs))
	for _, id := range parentIDs {
		args = append(args, id)
	}
	rows, err := s.db.QueryContext(ctx, `SELECT issue_dependencies.parent_issue_id, issue_dependencies.dependency_issue_id, dependencies.status
		FROM issue_dependencies
		JOIN issues AS dependencies ON dependencies.id = issue_dependencies.dependency_issue_id
		WHERE issue_dependencies.parent_issue_id IN (`+placeholders(len(parentIDs))+`)
		ORDER BY issue_dependencies.parent_issue_id ASC, issue_dependencies.dependency_issue_id ASC`, args...)
	if err != nil {
		return nil, fmt.Errorf("list dependency statuses: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var parentID int64
		var status dependencyStatus
		if err := rows.Scan(&parentID, &status.DependencyID, &status.Status); err != nil {
			return nil, fmt.Errorf("scan dependency status: %w", err)
		}
		statuses[parentID] = append(statuses[parentID], status)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dependency statuses: %w", err)
	}
	return statuses, nil
}

func issueIDs(issues []entity.Issue) []int64 {
	ids := make([]int64, 0, len(issues))
	for _, issue := range issues {
		ids = append(ids, issue.ID)
	}
	return ids
}

func blockingDependencyIDs(dependencies []dependencyStatus) []int64 {
	ids := []int64{}
	for _, dependency := range dependencies {
		if !entity.IsSatisfiedDependencyStatus(dependency.Status) {
			ids = append(ids, dependency.DependencyID)
		}
	}
	return ids
}

func queueIssueOrderClause() string {
	return `CASE issues.priority
		WHEN 'urgent' THEN 4
		WHEN 'high' THEN 3
		WHEN 'normal' THEN 2
		WHEN 'low' THEN 1
		ELSE 0
	END DESC, issues.id ASC`
}
