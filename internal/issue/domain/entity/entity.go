package entity

import (
	"errors"
	"time"
)

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
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type CreateProjectInput struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateProjectInput struct {
	Key         *string `json:"key"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
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
	if input.Name == "" {
		return input, errors.New("name is required")
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
	if input.Path == "" {
		return input, errors.New("path is required")
	}
	if input.Status == "" {
		input.Status = WorkspaceActive
	}
	if !IsValidWorkspaceStatus(input.Status) {
		return input, errors.New("status is invalid")
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

func IsValidWorkspaceStatus(status WorkspaceStatus) bool {
	switch status {
	case WorkspaceActive, WorkspaceInactive, WorkspaceArchived:
		return true
	default:
		return false
	}
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
