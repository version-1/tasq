package orchestrator

import "time"

type RunStatus string

const (
	RunQueued    RunStatus = "queued"
	RunRunning   RunStatus = "running"
	RunSucceeded RunStatus = "succeeded"
	RunFailed    RunStatus = "failed"
)

type Run struct {
	ID             int64     `json:"id"`
	RunID          string    `json:"runId"`
	IssueID        int64     `json:"issueId"`
	WorkItemID     int64     `json:"workItemId"`
	ClaimToken     string    `json:"claimToken"`
	Status         RunStatus `json:"status"`
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
	Status         RunStatus  `json:"status"`
	Workspace      string     `json:"workspace"`
	Attempt        int        `json:"attempt"`
	Error          string     `json:"error"`
	OrchestratorID string     `json:"orchestratorId"`
	OccurredAt     time.Time  `json:"occurredAt"`
	SentAt         *time.Time `json:"sentAt,omitempty"`
}
