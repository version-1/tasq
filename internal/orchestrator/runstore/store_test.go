package runstore

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/version-1/tasq/db/migrations"
	"github.com/version-1/tasq/internal/migration"
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
	if !schemaColumnExists(t, store, "runs", "thread_id") {
		t.Fatal("runs.thread_id column does not exist")
	}
}

func TestOrchestratorThreadIDMigrationRollsBack(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "orchestrator.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()
	if !schemaColumnExists(t, store, "runs", "thread_id") {
		t.Fatal("runs.thread_id column does not exist after migrate up")
	}

	rolledBack, err := migration.NewManager(store.db, migrations.Files, "orchestrator").Down(ctx)
	if err != nil {
		t.Fatalf("migration down: %v", err)
	}
	if rolledBack == nil || rolledBack.Version != "20260615000001" {
		t.Fatalf("rolled back = %+v, want 20260615000001", rolledBack)
	}
	if schemaColumnExists(t, store, "runs", "thread_id") {
		t.Fatal("runs.thread_id column still exists after rollback")
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

func TestStorePersistsThreadIDAndFindsLatestResumeThreadByIssueID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "orchestrator.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	withoutThread := createRun(t, ctx, store, 7)
	if withoutThread.ThreadID != "" {
		t.Fatalf("thread id = %q, want empty", withoutThread.ThreadID)
	}
	first, err := store.CreateRun(ctx, CreateRunInput{
		IssueID:        7,
		Workspace:      "/tmp/workspace",
		ThreadID:       "thread-first",
		Attempt:        2,
		OrchestratorID: "orchestrator",
	})
	if err != nil {
		t.Fatalf("create first threaded run: %v", err)
	}
	if first.ThreadID != "thread-first" {
		t.Fatalf("created thread id = %q", first.ThreadID)
	}
	updated, err := store.UpdateRunThreadID(ctx, withoutThread.RunID, "thread-latest")
	if err != nil {
		t.Fatalf("update run thread id: %v", err)
	}
	if updated.ThreadID != "thread-latest" {
		t.Fatalf("updated thread id = %q", updated.ThreadID)
	}

	threadID, err := store.LatestResumeThreadIDByIssueID(ctx, 7)
	if err != nil {
		t.Fatalf("latest resume thread id: %v", err)
	}
	if threadID != "thread-first" {
		t.Fatalf("latest thread id = %q, want thread-first", threadID)
	}
	if _, err := store.LatestResumeThreadIDByIssueID(ctx, 99); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("missing latest thread error = %v, want sql.ErrNoRows", err)
	}
}

func TestStoreInvalidatesResumeThreadIDsWhenWorkspaceCleanupIsMarked(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "orchestrator.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	storedRun, err := store.CreateRun(ctx, CreateRunInput{
		IssueID:        7,
		Workspace:      "/tmp/workspace",
		ThreadID:       "thread-stale",
		Attempt:        1,
		OrchestratorID: "orchestrator",
	})
	if err != nil {
		t.Fatalf("create threaded run: %v", err)
	}
	if err := store.UpsertWorkspaceMetadata(ctx, WorkspaceMetadataInput{
		WorkspaceKey: "issue-7",
		IssueID:      7,
		Path:         "/tmp/workspace",
		CreatedNow:   true,
		SourcePath:   "/tmp/repo",
	}); err != nil {
		t.Fatalf("upsert workspace metadata: %v", err)
	}
	if err := store.MarkWorkspaceCleanup(ctx, "issue-7", "removed", ""); err != nil {
		t.Fatalf("mark workspace cleanup: %v", err)
	}

	if _, err := store.LatestResumeThreadIDByIssueID(ctx, 7); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("latest resume thread error = %v, want sql.ErrNoRows", err)
	}
	updated, err := store.RunByRunID(ctx, storedRun.RunID)
	if err != nil {
		t.Fatalf("run by run id: %v", err)
	}
	if updated.ThreadID != "" {
		t.Fatalf("thread id after cleanup = %q, want empty", updated.ThreadID)
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

	for _, event := range []struct {
		eventType string
		message   string
		payload   string
	}{
		{eventType: "turn_completed", message: "done", payload: `{"ok":true}`},
		{eventType: "running", message: "runner started"},
		{eventType: "succeeded", message: "runner succeeded"},
		{eventType: "failed", message: "runner failed"},
		{eventType: "cancelled", message: "runner cancelled"},
		{eventType: "item/completed", message: "item completed"},
		{eventType: "item/commandExecution/requestApproval", message: "approval requested"},
		{eventType: "thread/tokenUsage/updated", message: "token usage updated", payload: `{"tokenUsage":{"total":{"totalTokens":1}}}`},
		{eventType: "account/rateLimits/updated", message: "rate limits updated", payload: `{"rateLimits":{"limitId":"codex"}}`},
	} {
		if err := store.RecordRunnerEvent(ctx, "run-1", event.eventType, event.message, event.payload); err != nil {
			t.Fatalf("record %s event: %v", event.eventType, err)
		}
	}
	if err := store.RecordRunnerEvent(ctx, "run-1", "debug", "ignore", ""); err != nil {
		t.Fatalf("record debug event: %v", err)
	}
	events, err := store.RunnerEvents(ctx, "run-1", 10)
	if err != nil {
		t.Fatalf("list runner events: %v", err)
	}
	if len(events) != 10 || events[0].EventType != "turn_completed" || events[0].PayloadJSON != `{"ok":true}` {
		t.Fatalf("events = %+v", events)
	}
	conversationEvents, err := store.ConversationEvents(ctx, "run-1")
	if err != nil {
		t.Fatalf("conversation events: %v", err)
	}
	var conversationEventTypes []string
	for _, event := range conversationEvents {
		conversationEventTypes = append(conversationEventTypes, event.EventType)
	}
	wantConversationEventTypes := []string{
		"turn_completed",
		"running",
		"succeeded",
		"failed",
		"cancelled",
		"item/completed",
		"item/commandExecution/requestApproval",
		"thread/tokenUsage/updated",
		"account/rateLimits/updated",
	}
	if strings.Join(conversationEventTypes, ",") != strings.Join(wantConversationEventTypes, ",") {
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

func TestDeleteProjectIssueDataRejectsRunningRunsWithoutDeleting(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "orchestrator.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	runningRun := createRun(t, ctx, store, 7)
	runningRun, err = store.UpdateRunStatus(ctx, runningRun.RunID, run.StatusRunning, "")
	if err != nil {
		t.Fatalf("mark running: %v", err)
	}
	if err := store.RecordRunnerEvent(ctx, runningRun.RunID, "running", "started", ""); err != nil {
		t.Fatalf("record runner event: %v", err)
	}
	if err := store.UpsertWorkspaceMetadata(ctx, WorkspaceMetadataInput{
		WorkspaceKey: "issue-7",
		IssueID:      7,
		Path:         "/tmp/workspace/issue-7",
		CreatedNow:   true,
		SourcePath:   "/tmp/repo",
	}); err != nil {
		t.Fatalf("upsert workspace metadata: %v", err)
	}
	if err := store.RecordWorkspaceSetupFailure(ctx, 7, "issue-7", "/tmp/workspace/issue-7", "failed"); err != nil {
		t.Fatalf("record workspace setup failure: %v", err)
	}

	err = store.DeleteProjectIssueData(ctx, []int64{7})

	if !errors.Is(err, ErrProjectHasRunningRuns) {
		t.Fatalf("delete error = %v, want ErrProjectHasRunningRuns", err)
	}
	if _, err := store.RunByRunID(ctx, runningRun.RunID); err != nil {
		t.Fatalf("running run was deleted: %v", err)
	}
	if count := countRows(t, store, "runner_events"); count != 1 {
		t.Fatalf("runner_events count = %d, want 1", count)
	}
	if count := countRows(t, store, "workspace_metadata"); count != 1 {
		t.Fatalf("workspace_metadata count = %d, want 1", count)
	}
	if count := countRows(t, store, "workspace_setup_failures"); count != 1 {
		t.Fatalf("workspace_setup_failures count = %d, want 1", count)
	}
}

func TestDeleteProjectIssueDataDeletesRunsEventsAndWorkspaceRecords(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "orchestrator.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	projectRun := createRun(t, ctx, store, 7)
	if _, err := store.UpdateRunStatus(ctx, projectRun.RunID, run.StatusSucceeded, ""); err != nil {
		t.Fatalf("mark project run succeeded: %v", err)
	}
	otherRun := createRun(t, ctx, store, 8)
	if _, err := store.UpdateRunStatus(ctx, otherRun.RunID, run.StatusSucceeded, ""); err != nil {
		t.Fatalf("mark other run succeeded: %v", err)
	}
	for _, runID := range []string{projectRun.RunID, otherRun.RunID} {
		if err := store.RecordRunnerEvent(ctx, runID, "succeeded", "done", ""); err != nil {
			t.Fatalf("record event for %s: %v", runID, err)
		}
	}
	for _, input := range []WorkspaceMetadataInput{
		{WorkspaceKey: "issue-7", IssueID: 7, Path: "/tmp/workspace/issue-7", CreatedNow: true, SourcePath: "/tmp/repo"},
		{WorkspaceKey: "issue-8", IssueID: 8, Path: "/tmp/workspace/issue-8", CreatedNow: true, SourcePath: "/tmp/repo"},
	} {
		if err := store.UpsertWorkspaceMetadata(ctx, input); err != nil {
			t.Fatalf("upsert workspace metadata: %v", err)
		}
	}
	if err := store.RecordWorkspaceSetupFailure(ctx, 7, "issue-7", "/tmp/workspace/issue-7", "failed"); err != nil {
		t.Fatalf("record project workspace setup failure: %v", err)
	}
	if err := store.RecordWorkspaceSetupFailure(ctx, 8, "issue-8", "/tmp/workspace/issue-8", "failed"); err != nil {
		t.Fatalf("record other workspace setup failure: %v", err)
	}

	if err := store.DeleteProjectIssueData(ctx, []int64{7}); err != nil {
		t.Fatalf("delete project issue data: %v", err)
	}
	if _, err := store.RunByRunID(ctx, projectRun.RunID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("project run error = %v, want sql.ErrNoRows", err)
	}
	if _, err := store.RunByRunID(ctx, otherRun.RunID); err != nil {
		t.Fatalf("other run error = %v", err)
	}
	if count := countRows(t, store, "runs"); count != 1 {
		t.Fatalf("runs count = %d, want 1", count)
	}
	if count := countRows(t, store, "runner_events"); count != 1 {
		t.Fatalf("runner_events count = %d, want 1", count)
	}
	if count := countRows(t, store, "workspace_metadata"); count != 1 {
		t.Fatalf("workspace_metadata count = %d, want 1", count)
	}
	if count := countRows(t, store, "workspace_setup_failures"); count != 1 {
		t.Fatalf("workspace_setup_failures count = %d, want 1", count)
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

func schemaColumnExists(t *testing.T, store *Store, table string, column string) bool {
	t.Helper()

	rows, err := store.db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		t.Fatalf("query columns for %s: %v", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			t.Fatalf("scan column for %s: %v", table, err)
		}
		if name == column {
			return true
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate columns for %s: %v", table, err)
	}
	return false
}

func countRows(t *testing.T, store *Store, table string) int {
	t.Helper()

	var count int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&count); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return count
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
