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

func TestIssueDetailHandlesMissingIssue(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issue-999", nil)
	rec := httptest.NewRecorder()

	New(&fakeStore{runByIssueErr: sql.ErrNoRows}, nil).Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
}

func TestIssueDetailReturnsRuns(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 8, 1, 0, 0, 0, time.UTC)
	store := &fakeStore{
		runByIssue: run.Run{
			RunID:     "run-latest",
			IssueID:   12,
			Status:    run.StatusSucceeded,
			Workspace: "/tmp/workspaces/issue-12",
			Attempt:   2,
			CreatedAt: now,
			UpdatedAt: now,
		},
		runsByIssue: []run.Run{{
			RunID:     "run-latest",
			IssueID:   12,
			Status:    run.StatusSucceeded,
			Attempt:   2,
			CreatedAt: now,
			UpdatedAt: now,
		}, {
			RunID:     "run-previous",
			IssueID:   12,
			Status:    run.StatusFailed,
			Attempt:   1,
			CreatedAt: now.Add(-time.Hour),
			UpdatedAt: now.Add(-time.Minute),
		}},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/issue-12", nil)
	rec := httptest.NewRecorder()

	New(store, nil).Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	var payload struct {
		IssueIdentifier string `json:"issue_identifier"`
		Runs            []struct {
			RunID   string `json:"run_id"`
			Status  string `json:"status"`
			Attempt int    `json:"attempt"`
		} `json:"runs"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.IssueIdentifier != "issue-12" || len(payload.Runs) != 2 {
		t.Fatalf("payload = %+v", payload)
	}
	if payload.Runs[0].RunID != "run-latest" || payload.Runs[1].RunID != "run-previous" {
		t.Fatalf("runs = %+v", payload.Runs)
	}
}

func TestConversationReturnsEvents(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 8, 1, 0, 0, 0, time.UTC)
	store := &fakeStore{
		runByRunID: run.Run{
			RunID:   "run-1",
			IssueID: 12,
		},
		conversationEvents: map[string][]run.RunnerEvent{
			"run-1": {{
				RunID:       "run-1",
				EventType:   "running",
				Message:     "runner started",
				PayloadJSON: "",
				OccurredAt:  now,
			}, {
				RunID:       "run-1",
				EventType:   "turn_completed",
				Message:     "turn_id=turn-1",
				PayloadJSON: `{"aggregatedOutput":"done"}`,
				OccurredAt:  now.Add(time.Second),
			}},
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/issue-12/runs/run-1/conversations", nil)
	rec := httptest.NewRecorder()

	New(store, nil).Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body = %s", rec.Code, rec.Body.String())
	}
	var payload struct {
		IssueID string `json:"issue_id"`
		RunID   string `json:"run_id"`
		Events  []struct {
			Event       string `json:"event"`
			PayloadJSON string `json:"payload_json"`
		} `json:"events"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.IssueID != "12" || payload.RunID != "run-1" || len(payload.Events) != 2 {
		t.Fatalf("payload = %+v", payload)
	}
	if payload.Events[1].Event != "turn_completed" || payload.Events[1].PayloadJSON != `{"aggregatedOutput":"done"}` {
		t.Fatalf("events = %+v", payload.Events)
	}
}

func TestConversationRejectsRunFromOtherIssue(t *testing.T) {
	t.Parallel()

	store := &fakeStore{
		runByRunID: run.Run{
			RunID:   "run-1",
			IssueID: 99,
		},
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/issue-12/runs/run-1/conversations", nil)
	rec := httptest.NewRecorder()

	New(store, nil).Handler().ServeHTTP(rec, req)

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
	activeRuns         []run.Run
	runByIssue         run.Run
	runByIssueErr      error
	runByRunID         run.Run
	runByRunIDErr      error
	runsByIssue        []run.Run
	conversationEvents map[string][]run.RunnerEvent
	events             map[string][]run.RunnerEvent
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

func (f *fakeStore) RunByRunID(ctx context.Context, runID string) (run.Run, error) {
	if f.runByRunIDErr != nil {
		return run.Run{}, f.runByRunIDErr
	}
	return f.runByRunID, nil
}

func (f *fakeStore) RunsByIssueID(ctx context.Context, issueID int64) ([]run.Run, error) {
	return f.runsByIssue, nil
}

func (f *fakeStore) ConversationEvents(ctx context.Context, runID string) ([]run.RunnerEvent, error) {
	return f.conversationEvents[runID], nil
}

func (f *fakeStore) RunnerEvents(ctx context.Context, runID string, limit int) ([]run.RunnerEvent, error) {
	return f.events[runID], nil
}

type fakeRefresher struct {
	count int
}

func (f *fakeRefresher) RequestRefresh() {
	f.count++
}
