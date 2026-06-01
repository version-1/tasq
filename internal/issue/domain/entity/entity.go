package entity

import (
	"errors"
	"os"
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

type Status string

const (
	StatusBacklog    Status = "backlog"
	StatusReady      Status = "ready"
	StatusInProgress Status = "in_progress"
	StatusReview     Status = "review"
	StatusDone       Status = "done"
	StatusBlocked    Status = "blocked"
	StatusFailed     Status = "failed"
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

type WorkspaceStatus string

const (
	WorkspaceActive   WorkspaceStatus = "active"
	WorkspaceInactive WorkspaceStatus = "inactive"
	WorkspaceArchived WorkspaceStatus = "archived"
)

type Project struct {
	ID          int64     `json:"id"`
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
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

type Workspace struct {
	ID        int64           `json:"id"`
	ProjectID int64           `json:"projectId"`
	Name      string          `json:"name"`
	Path      string          `json:"path"`
	Status    WorkspaceStatus `json:"status"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
}

type CreateWorkspaceInput struct {
	ProjectID int64           `json:"projectId"`
	Name      string          `json:"name"`
	Path      string          `json:"path"`
	Status    WorkspaceStatus `json:"status"`
}

type UpdateWorkspaceInput struct {
	ProjectID *int64           `json:"projectId"`
	Name      *string          `json:"name"`
	Path      *string          `json:"path"`
	Status    *WorkspaceStatus `json:"status"`
}

type Issue struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	Priority    Priority  `json:"priority"`
	Assignee    string    `json:"assignee"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type IssueState struct {
	ID     int64  `json:"id"`
	Status Status `json:"status"`
}

type CreateIssueInput struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Status      Status   `json:"status"`
	Priority    Priority `json:"priority"`
	Assignee    string   `json:"assignee"`
}

type UpdateIssueInput struct {
	Title       *string   `json:"title"`
	Description *string   `json:"description"`
	Status      *Status   `json:"status"`
	Priority    *Priority `json:"priority"`
	Assignee    *string   `json:"assignee"`
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

type IssueSummary struct {
	Issue
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
	if err := validateExistingDirectoryPath("location", input.Location); err != nil {
		return input, err
	}
	return input, nil
}

func NormalizeCreateWorkspace(input CreateWorkspaceInput) (CreateWorkspaceInput, error) {
	if input.ProjectID <= 0 {
		return input, errors.New("projectId is required")
	}
	if input.Name == "" {
		return input, errors.New("name is required")
	}
	if runeCount(input.Name) > maxNameLength {
		return input, errors.New("name must be 200 characters or fewer")
	}
	if input.Path == "" {
		return input, errors.New("path is required")
	}
	if err := validateExistingDirectoryPath("path", input.Path); err != nil {
		return input, err
	}
	if input.Status == "" {
		input.Status = WorkspaceActive
	}
	if !IsValidWorkspaceStatus(input.Status) {
		return input, errors.New("status is invalid")
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
	return input, nil
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
		if err := validateExistingDirectoryPath("location", *input.Location); err != nil {
			return input, err
		}
	}
	return input, nil
}

func NormalizeUpdateWorkspace(input UpdateWorkspaceInput) (UpdateWorkspaceInput, error) {
	if input.ProjectID != nil && *input.ProjectID <= 0 {
		return input, errors.New("projectId is required")
	}
	if input.Name != nil {
		if *input.Name == "" {
			return input, errors.New("name is required")
		}
		if runeCount(*input.Name) > maxNameLength {
			return input, errors.New("name must be 200 characters or fewer")
		}
	}
	if input.Path != nil {
		if err := validateExistingDirectoryPath("path", *input.Path); err != nil {
			return input, err
		}
	}
	if input.Status != nil && !IsValidWorkspaceStatus(*input.Status) {
		return input, errors.New("status is invalid")
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

func IsValidStatus(status Status) bool {
	switch status {
	case StatusBacklog, StatusReady, StatusInProgress, StatusReview, StatusDone, StatusBlocked, StatusFailed:
		return true
	default:
		return false
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

func IsValidWorkspaceStatus(status WorkspaceStatus) bool {
	switch status {
	case WorkspaceActive, WorkspaceInactive, WorkspaceArchived:
		return true
	default:
		return false
	}
}

func validateExistingDirectoryPath(field string, value string) error {
	if value == "" {
		return errors.New(field + " is required")
	}
	if runeCount(value) > maxPathLength {
		return errors.New(field + " must be 1000 characters or fewer")
	}
	if !filepath.IsAbs(value) {
		return errors.New(field + " must be an absolute path")
	}
	info, err := os.Stat(value)
	if err != nil {
		return errors.New(field + " must exist")
	}
	if !info.IsDir() {
		return errors.New(field + " must be a directory")
	}
	return nil
}

func runeCount(value string) int {
	return utf8.RuneCountInString(value)
}

func OrderedStatuses() []Status {
	return []Status{StatusBacklog, StatusReady, StatusInProgress, StatusReview, StatusBlocked, StatusFailed, StatusDone}
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
	case StatusDone:
		return "Done"
	default:
		return string(status)
	}
}
