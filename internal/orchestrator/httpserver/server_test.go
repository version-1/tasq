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

type fakeStore struct {
	activeRuns    []run.Run
	runByIssue    run.Run
	runByIssueErr error
	events        map[string][]run.RunnerEvent
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

type fakeRefresher struct {
	count int
}

func (f *fakeRefresher) RequestRefresh() {
	f.count++
}
