package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func issueColumns() string {
	return `issues.id, issues.project_id, projects.key, issues.title, issues.description, issues.status, issues.priority, issues.assignee, issues.created_at, issues.updated_at`
}

func commentColumns() string {
	return `id, issue_id, author, type, body, created_at`
}

func changeRequestColumns() string {
	return `id, issue_id, author, body, status, created_at, updated_at, resolved_at, resolved_by_run_id, result_comment_id`
}

func projectColumns() string {
	return `id, key, name, description, location, COALESCE((SELECT checksum FROM project_workflows WHERE project_workflows.project_id = projects.id), ''), created_at, updated_at`
}

func projectWorkflowColumns() string {
	return `project_id, frontmatter_json, body, checksum, created_at, updated_at`
}

func placeholders(count int) string {
	return strings.TrimRight(strings.Repeat("?,", count), ",")
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanIssue(row rowScanner) (entity.Issue, error) {
	var item entity.Issue
	var createdAt string
	var updatedAt string
	err := row.Scan(&item.ID, &item.ProjectID, &item.ProjectKey, &item.Title, &item.Description, &item.Status, &item.Priority, &item.Assignee, &createdAt, &updatedAt)
	if err != nil {
		return entity.Issue{}, err
	}
	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return entity.Issue{}, err
	}
	parsedUpdatedAt, err := parseTime(updatedAt)
	if err != nil {
		return entity.Issue{}, err
	}
	item.CreatedAt = parsedCreatedAt
	item.UpdatedAt = parsedUpdatedAt
	return item, nil
}

func scanComment(row rowScanner) (entity.Comment, error) {
	var item entity.Comment
	var createdAt string
	err := row.Scan(&item.ID, &item.IssueID, &item.Author, &item.Type, &item.Body, &createdAt)
	if err != nil {
		return entity.Comment{}, err
	}
	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return entity.Comment{}, err
	}
	item.CreatedAt = parsedCreatedAt
	return item, nil
}

func scanChangeRequest(row rowScanner) (entity.ChangeRequest, error) {
	var item entity.ChangeRequest
	var createdAt string
	var updatedAt string
	var resolvedAt sql.NullString
	var resolvedByRunID sql.NullString
	var resultCommentID sql.NullInt64
	err := row.Scan(&item.ID, &item.IssueID, &item.Author, &item.Body, &item.Status, &createdAt, &updatedAt, &resolvedAt, &resolvedByRunID, &resultCommentID)
	if err != nil {
		return entity.ChangeRequest{}, err
	}
	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return entity.ChangeRequest{}, err
	}
	parsedUpdatedAt, err := parseTime(updatedAt)
	if err != nil {
		return entity.ChangeRequest{}, err
	}
	item.CreatedAt = parsedCreatedAt
	item.UpdatedAt = parsedUpdatedAt
	if resolvedAt.Valid {
		parsedResolvedAt, err := parseTime(resolvedAt.String)
		if err != nil {
			return entity.ChangeRequest{}, err
		}
		item.ResolvedAt = &parsedResolvedAt
	}
	if resolvedByRunID.Valid {
		value := resolvedByRunID.String
		item.ResolvedByRunID = &value
	}
	if resultCommentID.Valid {
		value := resultCommentID.Int64
		item.ResultCommentID = &value
	}
	return item, nil
}

func scanProject(row rowScanner) (entity.Project, error) {
	var item entity.Project
	var createdAt string
	var updatedAt string
	err := row.Scan(&item.ID, &item.Key, &item.Name, &item.Description, &item.Location, &item.WorkflowChecksum, &createdAt, &updatedAt)
	if err != nil {
		return entity.Project{}, err
	}
	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return entity.Project{}, err
	}
	parsedUpdatedAt, err := parseTime(updatedAt)
	if err != nil {
		return entity.Project{}, err
	}
	item.CreatedAt = parsedCreatedAt
	item.UpdatedAt = parsedUpdatedAt
	return item, nil
}

func scanProjectWorkflow(row rowScanner) (entity.ProjectWorkflow, error) {
	var item entity.ProjectWorkflow
	var frontmatterJSON string
	var createdAt string
	var updatedAt string
	err := row.Scan(&item.ProjectID, &frontmatterJSON, &item.Body, &item.Checksum, &createdAt, &updatedAt)
	if err != nil {
		return entity.ProjectWorkflow{}, err
	}
	if err := json.Unmarshal([]byte(frontmatterJSON), &item.Frontmatter); err != nil {
		return entity.ProjectWorkflow{}, fmt.Errorf("parse project workflow frontmatter: %w", err)
	}
	if item.Frontmatter == nil {
		item.Frontmatter = map[string]any{}
	}
	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return entity.ProjectWorkflow{}, err
	}
	parsedUpdatedAt, err := parseTime(updatedAt)
	if err != nil {
		return entity.ProjectWorkflow{}, err
	}
	item.CreatedAt = parsedCreatedAt
	item.UpdatedAt = parsedUpdatedAt
	return item, nil
}

func nowString() string {
	return formatTime(time.Now().UTC())
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func formatOptionalTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return formatTime(*value)
}

func stringPtrValue(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func int64PtrValue(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse time: %w", err)
	}
	return parsed, nil
}
