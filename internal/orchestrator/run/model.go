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
	ID             int64      `json:"id"`
	RunID          string     `json:"runId"`
	IssueID        int64      `json:"issueId"`
	Status         Status     `json:"status"`
	Workspace      string     `json:"workspace"`
	Attempt        int        `json:"attempt"`
	Error          string     `json:"error"`
	RetryAfter     *time.Time `json:"retryAfter,omitempty"`
	OrchestratorID string     `json:"orchestratorId"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type TokenSummary struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
	TotalTokens  int64 `json:"total_tokens"`
}

type RunnerEvent struct {
	ID          int64     `json:"id"`
	RunID       string    `json:"runId"`
	EventType   string    `json:"eventType"`
	Message     string    `json:"message"`
	PayloadJSON string    `json:"payloadJson"`
	OccurredAt  time.Time `json:"occurredAt"`
}
