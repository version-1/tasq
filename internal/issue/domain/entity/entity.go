package entity

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"regexp"
	"time"
	"unicode/utf8"
)

const (
	maxAssigneeLength    = 200
	maxDescriptionLength = 10000
	maxIssueTitleLength  = 500
	maxNameLength        = 200
	maxPathLength        = 1000
)

var projectKeyPattern = regexp.MustCompile(`^([A-Z][A-Z0-9_]{0,19}|[a-z][a-z0-9-]{0,63})$`)
var sha256HexPattern = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)

type Status string

const (
	StatusBacklog    Status = "backlog"
	StatusReady      Status = "ready"
	StatusInProgress Status = "in_progress"
	StatusReview     Status = "review"
	StatusDone       Status = "done"
	StatusBlocked    Status = "blocked"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
	StatusDuplicate  Status = "duplicate"
)

type QueueStatus string

const (
	QueueStatusBacklog    QueueStatus = "backlog"
	QueueStatusPending    QueueStatus = "pending"
	QueueStatusQueued     QueueStatus = "queued"
	QueueStatusProcessing QueueStatus = "processing"
	QueueStatusCompleted  QueueStatus = "completed"
	QueueStatusInactive   QueueStatus = "inactive"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

type CommentType string

const (
	CommentProgress CommentType = "progress"
	CommentBlocker  CommentType = "blocker"
	CommentHandoff  CommentType = "handoff"
	CommentGeneral  CommentType = "general"
)

type Project struct {
	ID               int64     `json:"id"`
	Key              string    `json:"key"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Location         string    `json:"location"`
	WorkflowChecksum string    `json:"workflowChecksum"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type ProjectWorkflow struct {
	ProjectID   int64          `json:"projectId"`
	Frontmatter map[string]any `json:"frontmatter"`
	Body        string         `json:"body"`
	Checksum    string         `json:"checksum"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

type CreateProjectInput struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Location    string `json:"location"`
}

type UpdateProjectInput struct {
	Key         *string `json:"key"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Location    *string `json:"location"`
}

type UpsertProjectWorkflowInput struct {
	ProjectID       int64  `json:"projectId"`
	FrontmatterJSON string `json:"frontmatterJson"`
	Body            string `json:"body"`
	Checksum        string `json:"checksum"`
}

type Issue struct {
	ID            int64     `json:"id"`
	ProjectID     int64     `json:"projectId"`
	ProjectKey    string    `json:"projectKey"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Status        Status    `json:"status"`
	Priority      Priority  `json:"priority"`
	Assignee      string    `json:"assignee"`
	DependencyIDs []int64   `json:"dependency_ids"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type IssueState struct {
	ID     int64  `json:"id"`
	Status Status `json:"status"`
}

type CreateIssueInput struct {
	ProjectID     int64    `json:"projectId"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Status        Status   `json:"status"`
	Priority      Priority `json:"priority"`
	Assignee      string   `json:"assignee"`
	DependencyIDs []int64  `json:"dependency_ids"`
}

type UpdateIssueInput struct {
	Title         *string   `json:"title"`
	Description   *string   `json:"description"`
	Status        *Status   `json:"status"`
	Priority      *Priority `json:"priority"`
	Assignee      *string   `json:"assignee"`
	DependencyIDs *[]int64  `json:"dependency_ids,omitempty"`
}

type QueueIssue struct {
	Issue
	BlockedDependencyIDs []int64 `json:"blocked_dependency_ids,omitempty"`
}

type Queue struct {
	Queued  []QueueIssue `json:"queued"`
	Pending []QueueIssue `json:"pending"`
}

type Comment struct {
	ID        int64       `json:"id"`
	IssueID   int64       `json:"issueId"`
	Author    string      `json:"author"`
	Type      CommentType `json:"type"`
	Body      string      `json:"body"`
	CreatedAt time.Time   `json:"createdAt"`
}

type CreateCommentInput struct {
	IssueID int64       `json:"issueId"`
	Author  string      `json:"author"`
	Type    CommentType `json:"type"`
	Body    string      `json:"body"`
}

type UpdateCommentInput struct {
	Body *string `json:"body"`
}

type IssueStats struct {
	CommentCount int `json:"commentCount"`
}

type IssueSummary struct {
	Issue
	QueueStatus QueueStatus `json:"queueStatus"`
	Stats       IssueStats  `json:"stats"`
}

type Column struct {
	Status Status         `json:"status"`
	Title  string         `json:"title"`
	Issues []IssueSummary `json:"issues"`
}

type Summary struct {
	Columns     []Column  `json:"columns"`
	GeneratedAt time.Time `json:"generatedAt"`
}

func NormalizeCreate(input CreateIssueInput) (CreateIssueInput, error) {
	if input.ProjectID <= 0 {
		return input, errors.New("projectId is required")
	}
	if input.Title == "" {
		return input, errors.New("title is required")
	}
	if runeCount(input.Title) > maxIssueTitleLength {
		return input, errors.New("title must be 500 characters or fewer")
	}
	if runeCount(input.Description) > maxDescriptionLength {
		return input, errors.New("description must be 10000 characters or fewer")
	}
	if runeCount(input.Assignee) > maxAssigneeLength {
		return input, errors.New("assignee must be 200 characters or fewer")
	}
	if input.Status == "" {
		input.Status = StatusBacklog
	}
	if !IsValidStatus(input.Status) {
		return input, errors.New("status is invalid")
	}
	if input.Priority == "" {
		input.Priority = PriorityNormal
	}
	if !IsValidPriority(input.Priority) {
		return input, errors.New("priority is invalid")
	}
	if err := validateDependencyIDs(input.DependencyIDs); err != nil {
		return input, err
	}
	return input, nil
}

func NormalizeCreateProject(input CreateProjectInput) (CreateProjectInput, error) {
	if input.Key == "" {
		return input, errors.New("key is required")
	}
	if !projectKeyPattern.MatchString(input.Key) {
		return input, errors.New("key is invalid")
	}
	if input.Name == "" {
		return input, errors.New("name is required")
	}
	if runeCount(input.Name) > maxNameLength {
		return input, errors.New("name must be 200 characters or fewer")
	}
	if runeCount(input.Description) > maxDescriptionLength {
		return input, errors.New("description must be 10000 characters or fewer")
	}
	if err := validateAbsolutePath("location", input.Location); err != nil {
		return input, err
	}
	return input, nil
}

func NormalizeUpdateIssue(input UpdateIssueInput) (UpdateIssueInput, error) {
	if input.Title != nil {
		if *input.Title == "" {
			return input, errors.New("title is required")
		}
		if runeCount(*input.Title) > maxIssueTitleLength {
			return input, errors.New("title must be 500 characters or fewer")
		}
	}
	if input.Description != nil && runeCount(*input.Description) > maxDescriptionLength {
		return input, errors.New("description must be 10000 characters or fewer")
	}
	if input.Status != nil && !IsValidStatus(*input.Status) {
		return input, errors.New("status is invalid")
	}
	if input.Priority != nil && !IsValidPriority(*input.Priority) {
		return input, errors.New("priority is invalid")
	}
	if input.Assignee != nil && runeCount(*input.Assignee) > maxAssigneeLength {
		return input, errors.New("assignee must be 200 characters or fewer")
	}
	if input.DependencyIDs != nil {
		if err := validateDependencyIDs(*input.DependencyIDs); err != nil {
			return input, err
		}
	}
	return input, nil
}

func validateDependencyIDs(ids []int64) error {
	seen := map[int64]struct{}{}
	for _, id := range ids {
		if id <= 0 {
			return errors.New("dependency_ids contains invalid issue id")
		}
		if _, ok := seen[id]; ok {
			return errors.New("dependency_ids contains duplicate issue id")
		}
		seen[id] = struct{}{}
	}
	return nil
}

func NormalizeUpdateProject(input UpdateProjectInput) (UpdateProjectInput, error) {
	if input.Key != nil {
		if *input.Key == "" {
			return input, errors.New("key is required")
		}
		if !projectKeyPattern.MatchString(*input.Key) {
			return input, errors.New("key is invalid")
		}
	}
	if input.Name != nil {
		if *input.Name == "" {
			return input, errors.New("name is required")
		}
		if runeCount(*input.Name) > maxNameLength {
			return input, errors.New("name must be 200 characters or fewer")
		}
	}
	if input.Description != nil && runeCount(*input.Description) > maxDescriptionLength {
		return input, errors.New("description must be 10000 characters or fewer")
	}
	if input.Location != nil {
		if err := validateAbsolutePath("location", *input.Location); err != nil {
			return input, err
		}
	}
	return input, nil
}

func NormalizeUpsertProjectWorkflow(input UpsertProjectWorkflowInput) (UpsertProjectWorkflowInput, error) {
	if input.ProjectID <= 0 {
		return input, errors.New("projectId is required")
	}
	if input.FrontmatterJSON == "" {
		return input, errors.New("frontmatterJson is required")
	}
	if !json.Valid([]byte(input.FrontmatterJSON)) {
		return input, errors.New("frontmatterJson must be valid JSON")
	}
	if input.Checksum == "" {
		return input, errors.New("checksum is required")
	}
	if !sha256HexPattern.MatchString(input.Checksum) {
		return input, errors.New("checksum must be a SHA256 hex string")
	}
	return input, nil
}

func NormalizeCreateComment(input CreateCommentInput) (CreateCommentInput, error) {
	if input.IssueID <= 0 {
		return input, errors.New("issueId is required")
	}
	if input.Author == "" {
		return input, errors.New("author is required")
	}
	if input.Type == "" {
		input.Type = CommentGeneral
	}
	if !IsValidCommentType(input.Type) {
		return input, errors.New("type is invalid")
	}
	if input.Body == "" {
		return input, errors.New("body is required")
	}
	if utf8.RuneCountInString(input.Body) > 10000 {
		return input, errors.New("body must be 10000 characters or fewer")
	}
	return input, nil
}

func NormalizeUpdateComment(input UpdateCommentInput) (UpdateCommentInput, error) {
	if input.Body != nil {
		if *input.Body == "" {
			return input, errors.New("body is required")
		}
		if runeCount(*input.Body) > maxDescriptionLength {
			return input, errors.New("body must be 10000 characters or fewer")
		}
	}
	return input, nil
}

func IsValidStatus(status Status) bool {
	switch status {
	case StatusBacklog, StatusReady, StatusInProgress, StatusReview, StatusDone, StatusBlocked, StatusFailed, StatusCancelled, StatusDuplicate:
		return true
	default:
		return false
	}
}

func IsActiveDependencyStatus(status Status) bool {
	switch status {
	case StatusBacklog, StatusReady, StatusInProgress, StatusReview:
		return true
	default:
		return false
	}
}

func IsSatisfiedDependencyStatus(status Status) bool {
	switch status {
	case StatusDone, StatusCancelled, StatusDuplicate:
		return true
	default:
		return false
	}
}

func IssueQueueStatus(status Status, hasActiveDependency bool) QueueStatus {
	switch status {
	case StatusBacklog:
		return QueueStatusBacklog
	case StatusReady:
		if hasActiveDependency {
			return QueueStatusPending
		}
		return QueueStatusQueued
	case StatusInProgress:
		return QueueStatusProcessing
	case StatusDone:
		return QueueStatusCompleted
	default:
		return QueueStatusInactive
	}
}

func IsValidPriority(priority Priority) bool {
	switch priority {
	case PriorityLow, PriorityNormal, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

func IsValidCommentType(commentType CommentType) bool {
	switch commentType {
	case CommentProgress, CommentBlocker, CommentHandoff, CommentGeneral:
		return true
	default:
		return false
	}
}

func validateAbsolutePath(field string, value string) error {
	if value == "" {
		return errors.New(field + " is required")
	}
	if runeCount(value) > maxPathLength {
		return errors.New(field + " must be 1000 characters or fewer")
	}
	if !filepath.IsAbs(value) {
		return errors.New(field + " must be an absolute path")
	}
	return nil
}

func runeCount(value string) int {
	return utf8.RuneCountInString(value)
}

func OrderedStatuses() []Status {
	return []Status{StatusBacklog, StatusReady, StatusInProgress, StatusReview, StatusBlocked, StatusFailed, StatusCancelled, StatusDuplicate, StatusDone}
}

func StatusTitle(status Status) string {
	switch status {
	case StatusBacklog:
		return "Backlog"
	case StatusReady:
		return "Ready"
	case StatusInProgress:
		return "In Progress"
	case StatusReview:
		return "Review"
	case StatusBlocked:
		return "Blocked"
	case StatusFailed:
		return "Failed"
	case StatusCancelled:
		return "Cancelled"
	case StatusDuplicate:
		return "Duplicate"
	case StatusDone:
		return "Done"
	default:
		return string(status)
	}
}
