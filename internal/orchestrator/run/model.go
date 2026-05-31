package run

import "time"

type Status string

const (
	StatusQueued    Status = "queued"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

type Run struct {
	ID             int64     `json:"id"`
	RunID          string    `json:"runId"`
	IssueID        int64     `json:"issueId"`
	WorkItemID     int64     `json:"workItemId"`
	ClaimToken     string    `json:"claimToken"`
	Status         Status    `json:"status"`
	Workspace      string    `json:"workspace"`
	Attempt        int       `json:"attempt"`
	Error          string    `json:"error"`
	OrchestratorID string    `json:"orchestratorId"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type OutboxEvent struct {
	ID             int64      `json:"id"`
	EventID        string     `json:"eventId"`
	RunID          string     `json:"runId"`
	IssueID        int64      `json:"issueId"`
	WorkItemID     int64      `json:"workItemId"`
	ClaimToken     string     `json:"claimToken"`
	Status         Status     `json:"status"`
	Workspace      string     `json:"workspace"`
	Attempt        int        `json:"attempt"`
	Error          string     `json:"error"`
	OrchestratorID string     `json:"orchestratorId"`
	OccurredAt     time.Time  `json:"occurredAt"`
	SentAt         *time.Time `json:"sentAt,omitempty"`
}

type RunnerEvent struct {
	ID          int64     `json:"id"`
	RunID       string    `json:"runId"`
	EventType   string    `json:"eventType"`
	Message     string    `json:"message"`
	PayloadJSON string    `json:"payloadJson"`
	OccurredAt  time.Time `json:"occurredAt"`
}
