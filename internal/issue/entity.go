package issue

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

type WorkItemStatus string

const (
	WorkItemPending WorkItemStatus = "pending"
	WorkItemClaimed WorkItemStatus = "claimed"
	WorkItemDone    WorkItemStatus = "done"
	WorkItemFailed  WorkItemStatus = "failed"
)

type RunStatus string

const (
	RunQueued          RunStatus = "queued"
	RunStarting        RunStatus = "starting"
	RunRunning         RunStatus = "running"
	RunWaitingForInput RunStatus = "waiting_for_input"
	RunSucceeded       RunStatus = "succeeded"
	RunFailed          RunStatus = "failed"
	RunCancelled       RunStatus = "cancelled"
)

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

type IssueEntity struct {
	ID          int64  `db:"id"`
	Title       string `db:"title"`
	Description string `db:"description"`
	Status      string `db:"status"`
	Priority    string `db:"priority"`
	Assignee    string `db:"assignee"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
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

type WorkItem struct {
	ID         int64          `json:"id"`
	IssueID    int64          `json:"issueId"`
	Status     WorkItemStatus `json:"status"`
	ClaimedBy  string         `json:"claimedBy"`
	ClaimToken string         `json:"claimToken"`
	LeaseUntil *time.Time     `json:"leaseUntil,omitempty"`
	Attempt    int            `json:"attempt"`
	CreatedAt  time.Time      `json:"createdAt"`
	UpdatedAt  time.Time      `json:"updatedAt"`
	Issue      Issue          `json:"issue"`
}

type WorkItemEntity struct {
	ID         int64  `db:"id"`
	IssueID    int64  `db:"issue_id"`
	Status     string `db:"status"`
	ClaimedBy  string `db:"claimed_by"`
	ClaimToken string `db:"claim_token"`
	LeaseUntil string `db:"lease_until"`
	Attempt    int    `db:"attempt"`
	CreatedAt  string `db:"created_at"`
	UpdatedAt  string `db:"updated_at"`
}

type ClaimWorkItemInput struct {
	OrchestratorID string `json:"orchestratorId"`
	LeaseSeconds   int    `json:"leaseSeconds"`
}

type ClaimWorkItemOutput struct {
	WorkItem *WorkItem `json:"workItem"`
}

type RunEventInput struct {
	EventID        string    `json:"eventId"`
	WorkItemID     int64     `json:"workItemId"`
	IssueID        int64     `json:"issueId"`
	RunID          string    `json:"runId"`
	ClaimToken     string    `json:"claimToken"`
	Status         RunStatus `json:"status"`
	Workspace      string    `json:"workspace"`
	Attempt        int       `json:"attempt"`
	Error          string    `json:"error"`
	OccurredAt     time.Time `json:"occurredAt"`
	OrchestratorID string    `json:"orchestratorId"`
}

type OrchestratorEventEntity struct {
	EventID        string `db:"event_id"`
	WorkItemID     int64  `db:"work_item_id"`
	IssueID        int64  `db:"issue_id"`
	RunID          string `db:"run_id"`
	ClaimToken     string `db:"claim_token"`
	Status         string `db:"status"`
	Workspace      string `db:"workspace"`
	Attempt        int    `db:"attempt"`
	Error          string `db:"error"`
	OccurredAt     string `db:"occurred_at"`
	OrchestratorID string `db:"orchestrator_id"`
	ReceivedAt     string `db:"received_at"`
}

type RunSnapshot struct {
	IssueID        int64     `json:"issueId"`
	WorkItemID     int64     `json:"workItemId"`
	RunID          string    `json:"runId"`
	Status         RunStatus `json:"status"`
	Workspace      string    `json:"workspace"`
	Attempt        int       `json:"attempt"`
	Error          string    `json:"error"`
	OrchestratorID string    `json:"orchestratorId"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type RunSnapshotEntity struct {
	IssueID        int64  `db:"issue_id"`
	WorkItemID     int64  `db:"work_item_id"`
	RunID          string `db:"run_id"`
	Status         string `db:"status"`
	Workspace      string `db:"workspace"`
	Attempt        int    `db:"attempt"`
	Error          string `db:"error"`
	OrchestratorID string `db:"orchestrator_id"`
	UpdatedAt      string `db:"updated_at"`
}

type IssueSummary struct {
	Issue
	Run *RunSnapshot `json:"run,omitempty"`
}

type Column struct {
	Status Status         `json:"status"`
	Title  string         `json:"title"`
	Issues []IssueSummary `json:"issues"`
}

type Summary struct {
	Columns     []Column      `json:"columns"`
	Runs        []RunSnapshot `json:"runs"`
	GeneratedAt time.Time     `json:"generatedAt"`
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
