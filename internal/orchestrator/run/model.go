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
	Status         Status    `json:"status"`
	Workspace      string    `json:"workspace"`
	ThreadID       string    `json:"threadId,omitempty"`
	Attempt        int       `json:"attempt"`
	Error          string    `json:"error"`
	OrchestratorID string    `json:"orchestratorId"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type RunnerEvent struct {
	ID          int64     `json:"id"`
	RunID       string    `json:"runId"`
	EventType   string    `json:"eventType"`
	Message     string    `json:"message"`
	PayloadJSON string    `json:"payloadJson"`
	OccurredAt  time.Time `json:"occurredAt"`
}
