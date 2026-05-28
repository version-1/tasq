package entity

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

type RunEntity struct {
	ID             int64  `db:"id"`
	RunID          string `db:"run_id"`
	IssueID        int64  `db:"issue_id"`
	WorkItemID     int64  `db:"work_item_id"`
	ClaimToken     string `db:"claim_token"`
	Status         string `db:"status"`
	Workspace      string `db:"workspace"`
	Attempt        int    `db:"attempt"`
	Error          string `db:"error"`
	OrchestratorID string `db:"orchestrator_id"`
	CreatedAt      string `db:"created_at"`
	UpdatedAt      string `db:"updated_at"`
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

type OutboxEventEntity struct {
	ID             int64  `db:"id"`
	EventID        string `db:"event_id"`
	RunID          string `db:"run_id"`
	IssueID        int64  `db:"issue_id"`
	WorkItemID     int64  `db:"work_item_id"`
	ClaimToken     string `db:"claim_token"`
	Status         string `db:"status"`
	Workspace      string `db:"workspace"`
	Attempt        int    `db:"attempt"`
	Error          string `db:"error"`
	OrchestratorID string `db:"orchestrator_id"`
	OccurredAt     string `db:"occurred_at"`
	SentAt         string `db:"sent_at"`
}
