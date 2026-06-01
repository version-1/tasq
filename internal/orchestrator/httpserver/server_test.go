package httpserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/version-1/tasq/internal/orchestrator/run"
)

func TestStateReturnsActiveRuns(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 31, 10, 0, 0, 0, time.UTC)
	store := &fakeStore{
		activeRuns: []run.Run{{
			RunID:     "run-1",
			IssueID:   7,
			Status:    run.StatusRunning,
			Workspace: "/tmp/workspaces/issue-7",
			CreatedAt: now,
			UpdatedAt: now,
		}},
		events: map[string][]run.RunnerEvent{
			"run-1": {{
				RunID:      "run-1",
				EventType:  "turn_started",
				Message:    "turn_id=turn-1",
				OccurredAt: now.Add(time.Second),
			}},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/state", nil)
	rec := httptest.NewRecorder()

	New(store, nil).Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Counts struct {
			Running int `json:"running"`
		} `json:"counts"`
		Running []struct {
			IssueIdentifier string `json:"issue_identifier"`
			State           string `json:"state"`
			TurnCount       int    `json:"turn_count"`
		} `json:"running"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Counts.Running != 1 || len(payload.Running) != 1 {
		t.Fatalf("payload = %+v", payload)
	}
	if payload.Running[0].IssueIdentifier != "issue-7" || payload.Running[0].State != "running" || payload.Running[0].TurnCount != 1 {
		t.Fatalf("running row = %+v", payload.Running[0])
	}
}

func TestStateReturnsTokenSummaries(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 31, 10, 0, 0, 0, time.UTC)
	store := &fakeStore{
		activeRuns: []run.Run{
			{RunID: "run-1", IssueID: 7, Status: run.StatusRunning, CreatedAt: now, UpdatedAt: now},
			{RunID: "run-2", IssueID: 8, Status: run.StatusRunning, CreatedAt: now, UpdatedAt: now},
		},
		events: map[string][]run.RunnerEvent{},
		tokens: map[string]run.TokenSummary{
			"run-1": {InputTokens: 10, OutputTokens: 5, TotalTokens: 15},
			"run-2": {InputTokens: 2, OutputTokens: 3, TotalTokens: 5},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/state", nil)
	rec := httptest.NewRecorder()

	New(store, nil).Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Running []struct {
			Tokens tokenSummary `json:"tokens"`
		} `json:"running"`
		CodexTotals tokenSummary `json:"codex_totals"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Running) != 2 {
		t.Fatalf("running = %+v", payload.Running)
	}
	if payload.Running[0].Tokens.TotalTokens != 15 || payload.Running[1].Tokens.TotalTokens != 5 {
		t.Fatalf("running tokens = %+v", payload.Running)
	}
	if payload.CodexTotals.InputTokens != 12 || payload.CodexTotals.OutputTokens != 8 || payload.CodexTotals.TotalTokens != 20 {
		t.Fatalf("codex totals = %+v", payload.CodexTotals)
	}
}

func TestIssueDetailHandlesMissingIssue(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issue-999", nil)
	rec := httptest.NewRecorder()

	New(&fakeStore{runByIssueErr: sql.ErrNoRows}, nil).Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
}

func TestRefreshRequestsWorkerRefresh(t *testing.T) {
	t.Parallel()

	refresher := &fakeRefresher{}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/refresh", nil)
	rec := httptest.NewRecorder()

	New(&fakeStore{}, refresher).Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	if refresher.count != 1 {
		t.Fatalf("refresh count = %d", refresher.count)
	}
}

func TestRefreshUnavailableWithoutWorker(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/refresh", nil)
	rec := httptest.NewRecorder()

	New(&fakeStore{}, nil).Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
}

type fakeStore struct {
	activeRuns    []run.Run
	runByIssue    run.Run
	runByIssueErr error
	events        map[string][]run.RunnerEvent
	tokens        map[string]run.TokenSummary
}

func (f *fakeStore) ActiveRuns(ctx context.Context) ([]run.Run, error) {
	return f.activeRuns, nil
}

func (f *fakeStore) RunByIssueID(ctx context.Context, issueID int64) (run.Run, error) {
	if f.runByIssueErr != nil {
		return run.Run{}, f.runByIssueErr
	}
	return f.runByIssue, nil
}

func (f *fakeStore) RunnerEvents(ctx context.Context, runID string, limit int) ([]run.RunnerEvent, error) {
	return f.events[runID], nil
}

func (f *fakeStore) TokensByRunIDs(ctx context.Context, runIDs []string) (map[string]run.TokenSummary, error) {
	output := make(map[string]run.TokenSummary, len(runIDs))
	for _, runID := range runIDs {
		output[runID] = f.tokens[runID]
	}
	return output, nil
}

type fakeRefresher struct {
	count int
}

func (f *fakeRefresher) RequestRefresh() {
	f.count++
}
