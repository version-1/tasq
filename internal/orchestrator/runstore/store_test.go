package runstore

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/version-1/tasq/internal/orchestrator/run"
)

func TestOpenAppliesOrchestratorSchema(t *testing.T) {
	t.Parallel()

	store, err := OpenMigrated(context.Background(), filepath.Join(t.TempDir(), "orchestrator.sqlite"))
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

func TestOpenRejectsPendingOrchestratorMigrations(t *testing.T) {
	t.Parallel()

	_, err := Open(context.Background(), filepath.Join(t.TempDir(), "orchestrator.sqlite"))
	if err == nil {
		t.Fatal("open store succeeded, want pending migration error")
	}
	if !strings.Contains(err.Error(), "run `tq migrate`") {
		t.Fatalf("error = %v, want tq migrate guidance", err)
	}
}

func TestStoreQueriesActiveRunsAndLatestRunByIssueID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "orchestrator.sqlite"))
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

	runsForIssue, err := store.RunsByIssueID(ctx, 2)
	if err != nil {
		t.Fatalf("runs by issue id: %v", err)
	}
	if len(runsForIssue) != 2 {
		t.Fatalf("runs for issue = %+v", runsForIssue)
	}
	if runsForIssue[0].RunID != latestForIssue.RunID || runsForIssue[1].RunID != running.RunID {
		t.Fatalf("runs order = %+v", runsForIssue)
	}
}

func TestStoreRecordsRunnerEventAndWorkspaceMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "orchestrator.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	if err := store.RecordRunnerEvent(ctx, "run-1", "turn_completed", "done", `{"ok":true}`); err != nil {
		t.Fatalf("record runner event: %v", err)
	}
	if err := store.RecordRunnerEvent(ctx, "run-1", "running", "runner started", ""); err != nil {
		t.Fatalf("record running event: %v", err)
	}
	if err := store.RecordRunnerEvent(ctx, "run-1", "debug", "ignore", ""); err != nil {
		t.Fatalf("record debug event: %v", err)
	}
	events, err := store.RunnerEvents(ctx, "run-1", 10)
	if err != nil {
		t.Fatalf("list runner events: %v", err)
	}
	if len(events) != 3 || events[0].EventType != "turn_completed" || events[0].PayloadJSON != `{"ok":true}` {
		t.Fatalf("events = %+v", events)
	}
	conversationEvents, err := store.ConversationEvents(ctx, "run-1")
	if err != nil {
		t.Fatalf("conversation events: %v", err)
	}
	if len(conversationEvents) != 2 {
		t.Fatalf("conversation events = %+v", conversationEvents)
	}
	if conversationEvents[0].EventType != "turn_completed" || conversationEvents[1].EventType != "running" {
		t.Fatalf("conversation event order = %+v", conversationEvents)
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
