package coordinator

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/orchestrator/run"
	"github.com/version-1/tasq/internal/orchestrator/runstore"
	"github.com/version-1/tasq/internal/orchestrator/workspace"
)

func TestPollQueuesReadyIssues(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	manager := newTestWorkspaceManager(t)
	poller := newTestPoller(t, store, manager, []entity.Issue{
		{ID: 42, Status: entity.StatusReady, Title: "Build polling"},
	})

	if err := poller.Poll(ctx); err != nil {
		t.Fatalf("poll: %v", err)
	}

	runs, err := store.ActiveRuns(ctx)
	if err != nil {
		t.Fatalf("active runs: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("runs = %+v", runs)
	}
	if runs[0].IssueID != 42 || runs[0].Status != run.StatusQueued || runs[0].Attempt != 1 {
		t.Fatalf("run = %+v", runs[0])
	}
	if runs[0].Workspace != filepath.Join(manager.Root(), "issue-42") {
		t.Fatalf("workspace = %q", runs[0].Workspace)
	}
	events, err := store.RunnerEvents(ctx, runs[0].RunID, 10)
	if err != nil {
		t.Fatalf("runner events: %v", err)
	}
	if len(events) != 1 || events[0].EventType != "queued" {
		t.Fatalf("events = %+v", events)
	}
}

func TestPollSkipsIssuesWithActiveRuns(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	manager := newTestWorkspaceManager(t)
	if _, err := store.CreateRun(ctx, runstore.CreateRunInput{
		IssueID:        42,
		Workspace:      filepath.Join(manager.Root(), "issue-42"),
		Attempt:        1,
		OrchestratorID: "test-orchestrator",
	}); err != nil {
		t.Fatalf("create existing run: %v", err)
	}
	poller := newTestPoller(t, store, manager, []entity.Issue{
		{ID: 42, Status: entity.StatusReady, Title: "Build polling"},
	})

	if err := poller.Poll(ctx); err != nil {
		t.Fatalf("poll: %v", err)
	}

	runs, err := store.ActiveRuns(ctx)
	if err != nil {
		t.Fatalf("active runs: %v", err)
	}
	if len(runs) != 1 {
		t.Fatalf("runs = %+v", runs)
	}
}

func TestPollRespectsMaxActiveRuns(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	manager := newTestWorkspaceManager(t)
	if _, err := store.CreateRun(ctx, runstore.CreateRunInput{
		IssueID:        1,
		Workspace:      filepath.Join(manager.Root(), "issue-1"),
		Attempt:        1,
		OrchestratorID: "test-orchestrator",
	}); err != nil {
		t.Fatalf("create existing run: %v", err)
	}
	poller := newTestPoller(t, store, manager, []entity.Issue{
		{ID: 2, Status: entity.StatusReady, Title: "Second task"},
		{ID: 3, Status: entity.StatusReady, Title: "Third task"},
	})

	if err := poller.Poll(ctx); err != nil {
		t.Fatalf("poll: %v", err)
	}

	runs, err := store.ActiveRuns(ctx)
	if err != nil {
		t.Fatalf("active runs: %v", err)
	}
	if len(runs) != 2 {
		t.Fatalf("runs = %+v", runs)
	}
}

type fakeTracker struct {
	issues []entity.Issue
}

func (t fakeTracker) IssuesByStates(ctx context.Context, states []string) ([]entity.Issue, error) {
	return t.issues, nil
}

func openTestStore(t *testing.T) *runstore.Store {
	t.Helper()
	store, err := runstore.Open(context.Background(), filepath.Join(t.TempDir(), "orchestrator.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("close store: %v", err)
		}
	})
	return store
}

func newTestWorkspaceManager(t *testing.T) *workspace.Manager {
	t.Helper()
	manager, err := workspace.NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("new workspace manager: %v", err)
	}
	return manager
}

func newTestPoller(t *testing.T, store *runstore.Store, manager *workspace.Manager, issues []entity.Issue) *Poller {
	t.Helper()
	poller, err := NewPoller(PollerConfig{
		Tracker:        fakeTracker{issues: issues},
		Store:          store,
		Workspaces:     manager,
		Interval:       time.Minute,
		MaxActiveRuns:  2,
		OrchestratorID: "test-orchestrator",
	})
	if err != nil {
		t.Fatalf("new poller: %v", err)
	}
	return poller
}
