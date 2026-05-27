package task

import (
	"errors"
	"time"
)

type Status string

const (
	StatusBacklog Status = "backlog"
	StatusReady   Status = "ready"
	StatusRunning Status = "running"
	StatusReview  Status = "review"
	StatusDone    Status = "done"
	StatusBlocked Status = "blocked"
	StatusFailed  Status = "failed"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityNormal Priority = "normal"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

type AgentStatus string

const (
	AgentStatusIdle      AgentStatus = "idle"
	AgentStatusQueued    AgentStatus = "queued"
	AgentStatusRunning   AgentStatus = "running"
	AgentStatusWaiting   AgentStatus = "waiting_for_input"
	AgentStatusSucceeded AgentStatus = "succeeded"
	AgentStatusFailed    AgentStatus = "failed"
)

type Task struct {
	ID          int64       `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Status      Status      `json:"status"`
	Priority    Priority    `json:"priority"`
	AgentStatus AgentStatus `json:"agentStatus"`
	Assignee    string      `json:"assignee"`
	Source      string      `json:"source"`
	SourceID    string      `json:"sourceId"`
	Workspace   string      `json:"workspace"`
	Attempts    int         `json:"attempts"`
	LastError   string      `json:"lastError"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type CreateTaskInput struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Status      Status   `json:"status"`
	Priority    Priority `json:"priority"`
	Assignee    string   `json:"assignee"`
	Source      string   `json:"source"`
	SourceID    string   `json:"sourceId"`
	Workspace   string   `json:"workspace"`
}

type UpdateTaskInput struct {
	Title       *string      `json:"title"`
	Description *string      `json:"description"`
	Status      *Status      `json:"status"`
	Priority    *Priority    `json:"priority"`
	AgentStatus *AgentStatus `json:"agentStatus"`
	Assignee    *string      `json:"assignee"`
	Workspace   *string      `json:"workspace"`
	LastError   *string      `json:"lastError"`
}

type Settings struct {
	PollIntervalSeconds int    `json:"pollIntervalSeconds"`
	MaxConcurrentRuns   int    `json:"maxConcurrentRuns"`
	WorkspaceRoot       string `json:"workspaceRoot"`
	TrackerProvider     string `json:"trackerProvider"`
	AgentCommand        string `json:"agentCommand"`
}

type Column struct {
	Status Status `json:"status"`
	Title  string `json:"title"`
	Tasks  []Task `json:"tasks"`
}

type Summary struct {
	Columns     []Column  `json:"columns"`
	Agents      []Task    `json:"agents"`
	Settings    Settings  `json:"settings"`
	GeneratedAt time.Time `json:"generatedAt"`
}

func NormalizeCreate(input CreateTaskInput) (CreateTaskInput, error) {
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

func IsValidStatus(status Status) bool {
	switch status {
	case StatusBacklog, StatusReady, StatusRunning, StatusReview, StatusDone, StatusBlocked, StatusFailed:
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

func IsValidAgentStatus(status AgentStatus) bool {
	switch status {
	case AgentStatusIdle, AgentStatusQueued, AgentStatusRunning, AgentStatusWaiting, AgentStatusSucceeded, AgentStatusFailed:
		return true
	default:
		return false
	}
}

func OrderedStatuses() []Status {
	return []Status{StatusBacklog, StatusReady, StatusRunning, StatusReview, StatusBlocked, StatusFailed, StatusDone}
}

func StatusTitle(status Status) string {
	switch status {
	case StatusBacklog:
		return "Backlog"
	case StatusReady:
		return "Ready"
	case StatusRunning:
		return "Running"
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
