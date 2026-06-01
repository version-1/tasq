package runstore

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/version-1/tasq/internal/orchestrator/run"
)

func TestOpenAppliesOrchestratorSchema(t *testing.T) {
	t.Parallel()

	store, err := Open(context.Background(), filepath.Join(t.TempDir(), "orchestrator.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	for _, name := range []string{
		"runs",
		"runs_issue_idx",
		"runs_status_updated_idx",
		"runner_events",
		"runner_events_run_idx",
		"workspace_metadata",
		"workspace_setup_failures",
		"workspace_setup_failures_issue_idx",
	} {
		if !schemaObjectExists(t, store, name) {
			t.Fatalf("schema object %q does not exist", name)
		}
	}
}

func TestStoreQueriesActiveRunsAndLatestRunByIssueID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "orchestrator.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	queued := createRun(t, ctx, store, 1)
	running := createRun(t, ctx, store, 2)
	running, err = store.UpdateRunStatus(ctx, running.RunID, run.StatusRunning, "")
	if err != nil {
		t.Fatalf("mark running: %v", err)
	}
	succeeded := createRun(t, ctx, store, 3)
	if _, err := store.UpdateRunStatus(ctx, succeeded.RunID, run.StatusSucceeded, ""); err != nil {
		t.Fatalf("mark succeeded: %v", err)
	}
	latestForIssue := createRun(t, ctx, store, 2)

	activeRuns, err := store.ActiveRuns(ctx)
	if err != nil {
		t.Fatalf("active runs: %v", err)
	}
	activeByRunID := map[string]run.Run{}
	for _, activeRun := range activeRuns {
		activeByRunID[activeRun.RunID] = activeRun
	}
	if _, ok := activeByRunID[queued.RunID]; !ok {
		t.Fatalf("queued run missing from active runs: %+v", activeRuns)
	}
	if _, ok := activeByRunID[running.RunID]; !ok {
		t.Fatalf("running run missing from active runs: %+v", activeRuns)
	}
	if _, ok := activeByRunID[latestForIssue.RunID]; !ok {
		t.Fatalf("latest queued run missing from active runs: %+v", activeRuns)
	}
	if _, ok := activeByRunID[succeeded.RunID]; ok {
		t.Fatalf("succeeded run should not be active: %+v", activeRuns)
	}

	latest, err := store.RunByIssueID(ctx, 2)
	if err != nil {
		t.Fatalf("run by issue id: %v", err)
	}
	if latest.RunID != latestForIssue.RunID {
		t.Fatalf("latest run = %s, want %s", latest.RunID, latestForIssue.RunID)
	}

	if _, err := store.RunByIssueID(ctx, 99); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("missing issue error = %v, want sql.ErrNoRows", err)
	}
}

func TestStoreRecordsRunnerEventAndWorkspaceMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "orchestrator.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	if err := store.RecordRunnerEvent(ctx, "run-1", "turn/completed", "done", `{"ok":true}`); err != nil {
		t.Fatalf("record runner event: %v", err)
	}
	events, err := store.RunnerEvents(ctx, "run-1", 10)
	if err != nil {
		t.Fatalf("list runner events: %v", err)
	}
	if len(events) != 1 || events[0].EventType != "turn/completed" || events[0].PayloadJSON != `{"ok":true}` {
		t.Fatalf("events = %+v", events)
	}

	if err := store.UpsertWorkspaceMetadata(ctx, WorkspaceMetadataInput{
		WorkspaceKey: "ISSUE-1",
		IssueID:      1,
		Path:         "/tmp/workspaces/ISSUE-1",
		CreatedNow:   true,
		SourcePath:   "/tmp/repo",
	}); err != nil {
		t.Fatalf("upsert workspace metadata: %v", err)
	}
	if err := store.MarkWorkspaceCleanup(ctx, "ISSUE-1", "removed", ""); err != nil {
		t.Fatalf("mark workspace cleanup: %v", err)
	}
	if err := store.RecordWorkspaceSetupFailure(ctx, 2, "ISSUE-2", "/tmp/workspaces/ISSUE-2", "failed"); err != nil {
		t.Fatalf("record workspace setup failure: %v", err)
	}
	count, err := store.WorkspaceSetupFailureCount(ctx)
	if err != nil {
		t.Fatalf("count workspace setup failures: %v", err)
	}
	if count != 1 {
		t.Fatalf("workspace setup failure count = %d", count)
	}
	if !schemaObjectExists(t, store, "workspace_metadata") {
		t.Fatal("workspace_metadata does not exist")
	}
}

func TestStoreSchedulesAndRequeuesDueRetry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "orchestrator.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()
	storedRun := createRun(t, ctx, store, 1)
	storedRun, err = store.UpdateRunStatus(ctx, storedRun.RunID, run.StatusRunning, "")
	if err != nil {
		t.Fatalf("mark running: %v", err)
	}
	retryAfter := time.Now().UTC().Add(-time.Minute)

	scheduled, err := store.ScheduleRetry(ctx, storedRun.RunID, "failed", retryAfter)
	if err != nil {
		t.Fatalf("schedule retry: %v", err)
	}
	if scheduled.Status != run.StatusFailed || scheduled.RetryAfter == nil {
		t.Fatalf("scheduled run = %+v", scheduled)
	}
	due, err := store.DueRetries(ctx, time.Now().UTC())
	if err != nil {
		t.Fatalf("due retries: %v", err)
	}
	if len(due) != 1 || due[0].RunID != storedRun.RunID {
		t.Fatalf("due retries = %+v", due)
	}

	requeued, err := store.RequeueRetry(ctx, storedRun.RunID)
	if err != nil {
		t.Fatalf("requeue retry: %v", err)
	}
	if requeued.Status != run.StatusQueued || requeued.Attempt != storedRun.Attempt+1 || requeued.RetryAfter != nil || requeued.Error != "" {
		t.Fatalf("requeued run = %+v", requeued)
	}
}

func TestStoreCompletionMethodsDoNotOverwriteTerminalRuns(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "orchestrator.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()
	storedRun := createRun(t, ctx, store, 1)
	storedRun, err = store.UpdateRunStatus(ctx, storedRun.RunID, run.StatusRunning, "")
	if err != nil {
		t.Fatalf("mark running: %v", err)
	}
	if _, err := store.UpdateRunStatus(ctx, storedRun.RunID, run.StatusFailed, "stall timeout"); err != nil {
		t.Fatalf("mark failed: %v", err)
	}

	completed, err := store.CompleteRunningRun(ctx, storedRun.RunID, run.StatusCancelled, "context canceled")
	if err != nil {
		t.Fatalf("complete running run: %v", err)
	}
	if completed.Status != run.StatusFailed || completed.Error != "stall timeout" {
		t.Fatalf("completed run = %+v", completed)
	}
	scheduled, err := store.ScheduleRetry(ctx, storedRun.RunID, "runner failed", time.Now().UTC())
	if err != nil {
		t.Fatalf("schedule retry: %v", err)
	}
	if scheduled.Status != run.StatusFailed || scheduled.Error != "stall timeout" || scheduled.RetryAfter != nil {
		t.Fatalf("scheduled run = %+v", scheduled)
	}
}

func TestStoreDueRetriesIgnoresFutureRetry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "orchestrator.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()
	storedRun := createRun(t, ctx, store, 1)
	storedRun, err = store.UpdateRunStatus(ctx, storedRun.RunID, run.StatusRunning, "")
	if err != nil {
		t.Fatalf("mark running: %v", err)
	}
	if _, err := store.ScheduleRetry(ctx, storedRun.RunID, "failed", time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatalf("schedule retry: %v", err)
	}

	due, err := store.DueRetries(ctx, time.Now().UTC())
	if err != nil {
		t.Fatalf("due retries: %v", err)
	}
	if len(due) != 0 {
		t.Fatalf("due retries = %+v", due)
	}
}

func TestStoreLastEventTimesFallsBackToCreatedAt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "orchestrator.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()
	storedRun := createRun(t, ctx, store, 1)
	if err := store.RecordRunnerEvent(ctx, storedRun.RunID, "progress", "working", ""); err != nil {
		t.Fatalf("record event: %v", err)
	}

	lastEventTimes, err := store.LastEventTimes(ctx, []string{storedRun.RunID})
	if err != nil {
		t.Fatalf("last event times: %v", err)
	}
	if !lastEventTimes[storedRun.RunID].After(storedRun.CreatedAt) {
		t.Fatalf("last event time = %s, created at = %s", lastEventTimes[storedRun.RunID], storedRun.CreatedAt)
	}

	noEventRun := createRun(t, ctx, store, 2)
	lastEventTimes, err = store.LastEventTimes(ctx, []string{noEventRun.RunID})
	if err != nil {
		t.Fatalf("last event times: %v", err)
	}
	if !lastEventTimes[noEventRun.RunID].Equal(noEventRun.CreatedAt) {
		t.Fatalf("last event time = %s, want created at %s", lastEventTimes[noEventRun.RunID], noEventRun.CreatedAt)
	}
}

func TestExtractTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantIn    int64
		wantOut   int64
		wantTotal int64
	}{
		{name: "usage", input: `{"usage":{"input_tokens":10,"output_tokens":5,"total_tokens":15}}`, wantIn: 10, wantOut: 5, wantTotal: 15},
		{name: "missing", input: `{"ok":true}`},
		{name: "malformed", input: `{`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input, output, total, err := ExtractTokens(tt.input)
			if err != nil {
				t.Fatalf("extract tokens: %v", err)
			}
			if input != tt.wantIn || output != tt.wantOut || total != tt.wantTotal {
				t.Fatalf("tokens = %d/%d/%d, want %d/%d/%d", input, output, total, tt.wantIn, tt.wantOut, tt.wantTotal)
			}
		})
	}
}

func TestStoreTokensByRunIDsAggregatesTurnCompletedEvents(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "orchestrator.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()
	storedRun := createRun(t, ctx, store, 1)
	if err := store.RecordRunnerEvent(ctx, storedRun.RunID, "turn_completed", "done", `{"usage":{"input_tokens":10,"output_tokens":5,"total_tokens":15}}`); err != nil {
		t.Fatalf("record event: %v", err)
	}
	if err := store.RecordRunnerEvent(ctx, storedRun.RunID, "turn_completed", "done", `{"usage":{"input_tokens":2,"output_tokens":3,"total_tokens":5}}`); err != nil {
		t.Fatalf("record event: %v", err)
	}
	if err := store.RecordRunnerEvent(ctx, storedRun.RunID, "turn/completed", "raw", `{"usage":{"input_tokens":100,"output_tokens":100,"total_tokens":200}}`); err != nil {
		t.Fatalf("record event: %v", err)
	}

	tokens, err := store.TokensByRunIDs(ctx, []string{storedRun.RunID, "missing-run"})
	if err != nil {
		t.Fatalf("tokens by run IDs: %v", err)
	}
	got := tokens[storedRun.RunID]
	if got.InputTokens != 12 || got.OutputTokens != 8 || got.TotalTokens != 20 {
		t.Fatalf("tokens = %+v", got)
	}
	if tokens["missing-run"] != (run.TokenSummary{}) {
		t.Fatalf("missing run tokens = %+v", tokens["missing-run"])
	}
}

func schemaObjectExists(t *testing.T, store *Store, name string) bool {
	t.Helper()

	var exists bool
	err := store.db.QueryRow(`SELECT EXISTS (
		SELECT 1 FROM sqlite_master WHERE name = ?
	)`, name).Scan(&exists)
	if err != nil {
		t.Fatalf("query schema object %q: %v", name, err)
	}
	return exists
}

func createRun(t *testing.T, ctx context.Context, store *Store, issueID int64) run.Run {
	t.Helper()

	storedRun, err := store.CreateRun(ctx, CreateRunInput{
		IssueID:        issueID,
		Workspace:      "/tmp/workspace",
		Attempt:        1,
		OrchestratorID: "orchestrator",
	})
	if err != nil {
		t.Fatalf("create run: %v", err)
	}
	return storedRun
}
