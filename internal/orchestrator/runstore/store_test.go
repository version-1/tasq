package runstore

import (
	"context"
	"path/filepath"
	"testing"
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
		"runs_work_item_idx",
		"runner_events",
		"runner_events_run_idx",
		"workspace_metadata",
		"workspace_setup_failures",
		"workspace_setup_failures_issue_idx",
		"outbox_events",
		"outbox_events_unsent_idx",
	} {
		if !schemaObjectExists(t, store, name) {
			t.Fatalf("schema object %q does not exist", name)
		}
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
	outboxCount, err := store.UnsentOutboxCount(ctx)
	if err != nil {
		t.Fatalf("count unsent outbox: %v", err)
	}
	if outboxCount != 0 {
		t.Fatalf("unsent outbox count = %d", outboxCount)
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
