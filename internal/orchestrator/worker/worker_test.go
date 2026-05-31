package worker

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/orchestrator/run"
	"github.com/version-1/tasq/internal/orchestrator/runner"
	"github.com/version-1/tasq/internal/orchestrator/runstore"
	"github.com/version-1/tasq/internal/orchestrator/workflow"
	"github.com/version-1/tasq/internal/orchestrator/workspace"
)

func TestRequestRefreshCoalescesWithoutBlocking(t *testing.T) {
	t.Parallel()

	worker := &Worker{refreshCh: make(chan struct{}, 1)}
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 10; i++ {
			worker.RequestRefresh()
		}
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("RequestRefresh blocked")
	}
	if got := len(worker.refreshCh); got != 1 {
		t.Fatalf("queued refreshes = %d, want 1", got)
	}
}

func TestRunProcessesRefreshBeforeNextPoll(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	store, err := runstore.Open(ctx, filepath.Join(t.TempDir(), "orchestrator.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	tracker := &fakeTrackerClient{claims: make(chan struct{}, 4)}
	worker, err := NewWithConfig(store, tracker, "orchestrator", 30, workflow.Definition{
		Config: workflow.Config{
			PollInterval:      time.Hour,
			WorkspaceRoot:     t.TempDir(),
			WorkspaceSource:   t.TempDir(),
			MaxConcurrentRuns: 1,
			MaxTurns:          1,
			MaxRetryAttempts:  1,
			MaxRetryBackoff:   time.Second,
			CodexCommand:      "codex app-server",
			CodexReadTimeout:  time.Second,
			CodexTurnTimeout:  time.Second,
		},
	}, runner.SimulatedRunner{})
	if err != nil {
		t.Fatalf("create worker: %v", err)
	}
	go worker.Run(ctx)
	waitClaim(t, tracker.claims)

	worker.RequestRefresh()
	waitClaim(t, tracker.claims)
}

func TestRunWithRetriesRunsBeforeAndAfterHooks(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	worker := &Worker{
		maxRetryAttempts: 1,
		hookConfig: workspace.HookConfig{
			BeforeRun: `echo before >> hooks.out`,
			AfterRun:  `echo after >> hooks.out`,
			Timeout:   time.Second,
		},
		runner: recordingRunner{result: runner.Result{Status: run.StatusSucceeded}},
	}

	result := worker.runWithRetries(context.Background(), "run-1", entity.WorkItem{IssueID: 1}, workspace.Workspace{Path: dir})
	if result.Status != run.StatusSucceeded {
		t.Fatalf("status = %q error=%q", result.Status, result.Error)
	}

	content, err := os.ReadFile(filepath.Join(dir, "hooks.out"))
	if err != nil {
		t.Fatalf("read hook output: %v", err)
	}
	if string(content) != "before\nafter\n" {
		t.Fatalf("hook output = %q", content)
	}
}

func TestRunWithRetriesBeforeRunFailureAbortsAttempt(t *testing.T) {
	t.Parallel()

	agentRunner := &countingRunner{result: runner.Result{Status: run.StatusSucceeded}}
	worker := &Worker{
		maxRetryAttempts: 1,
		hookConfig: workspace.HookConfig{
			BeforeRun: `echo no >&2; exit 1`,
			Timeout:   time.Second,
		},
		runner: agentRunner,
	}

	result := worker.runWithRetries(context.Background(), "run-1", entity.WorkItem{IssueID: 1}, workspace.Workspace{Path: t.TempDir()})
	if result.Status != run.StatusFailed {
		t.Fatalf("status = %q", result.Status)
	}
	if agentRunner.count != 0 {
		t.Fatalf("runner calls = %d, want 0", agentRunner.count)
	}
}

func waitClaim(t *testing.T, claims <-chan struct{}) {
	t.Helper()
	select {
	case <-claims:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for claim")
	}
}

type fakeTrackerClient struct {
	mu     sync.Mutex
	claims chan struct{}
}

type recordingRunner struct {
	result runner.Result
}

func (r recordingRunner) Run(ctx context.Context, task runner.Task) runner.Result {
	return r.result
}

type countingRunner struct {
	count  int
	result runner.Result
}

func (r *countingRunner) Run(ctx context.Context, task runner.Task) runner.Result {
	r.count++
	return r.result
}

func (f *fakeTrackerClient) ClaimWorkItem(ctx context.Context, orchestratorID string, leaseSeconds int) (*entity.WorkItem, error) {
	select {
	case f.claims <- struct{}{}:
	default:
	}
	return nil, nil
}

func (f *fakeTrackerClient) RenewWorkItemLease(ctx context.Context, input entity.RenewWorkItemLeaseInput) error {
	return nil
}

func (f *fakeTrackerClient) SendRunEvent(ctx context.Context, event run.OutboxEvent) error {
	return nil
}

func (f *fakeTrackerClient) Issue(ctx context.Context, id int64) (entity.Issue, error) {
	return entity.Issue{}, nil
}

func (f *fakeTrackerClient) Issues(ctx context.Context) ([]entity.Issue, error) {
	return nil, nil
}
