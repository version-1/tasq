package worker

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/orchestrator/run"
	"github.com/version-1/tasq/internal/orchestrator/runner"
	"github.com/version-1/tasq/internal/orchestrator/runstore"
	"github.com/version-1/tasq/internal/orchestrator/workflow"
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
