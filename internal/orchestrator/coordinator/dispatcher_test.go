package coordinator

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/orchestrator/run"
	"github.com/version-1/tasq/internal/orchestrator/runner"
	"github.com/version-1/tasq/internal/orchestrator/runstore"
	"github.com/version-1/tasq/internal/orchestrator/workflow"
)

func TestDispatcherDoesNothingWithoutQueuedRuns(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	testRunner := &recordingRunner{result: runner.Result{Status: run.StatusSucceeded}}
	dispatcher := newTestDispatcher(t, store, testRunner, []entity.Issue{{ID: 42, Status: entity.StatusReady}})

	if err := dispatcher.Dispatch(ctx, nil); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if got := testRunner.runCount(); got != 0 {
		t.Fatalf("run count = %d", got)
	}
}

func TestDispatcherCompletesSuccessfulRun(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	storedRun := createQueuedRun(t, store, 42)
	testRunner := &recordingRunner{
		result: runner.Result{Status: run.StatusSucceeded},
		events: []runner.Event{
			{EventType: "turn_completed", Message: "done", PayloadJSON: `{"ok":true}`},
		},
	}
	dispatcher := newTestDispatcher(t, store, testRunner, []entity.Issue{{ID: 42, Status: entity.StatusReady, Title: "Run task"}})

	if err := dispatcher.Dispatch(ctx, []run.Run{storedRun}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	shutdownDispatcher(t, dispatcher)

	updated, err := store.RunByRunID(ctx, storedRun.RunID)
	if err != nil {
		t.Fatalf("run by id: %v", err)
	}
	if updated.Status != run.StatusSucceeded || updated.Error != "" {
		t.Fatalf("updated run = %+v", updated)
	}
	events, err := store.RunnerEvents(ctx, storedRun.RunID, 10)
	if err != nil {
		t.Fatalf("runner events: %v", err)
	}
	eventTypes := eventTypes(events)
	if strings.Join(eventTypes, ",") != "running,turn_completed,succeeded" {
		t.Fatalf("event types = %v", eventTypes)
	}
	task := testRunner.firstTask(t)
	if task.Issue.ID != 42 || task.RunID != storedRun.RunID || task.Workspace.Path == "" {
		t.Fatalf("task = %+v", task)
	}
}

func TestDispatcherRecordsFailedRun(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	storedRun := createQueuedRun(t, store, 42)
	testRunner := &recordingRunner{result: runner.Result{Status: run.StatusFailed, Error: "startup failed"}}
	dispatcher := newTestDispatcher(t, store, testRunner, []entity.Issue{{ID: 42, Status: entity.StatusReady}})

	if err := dispatcher.Dispatch(ctx, []run.Run{storedRun}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	shutdownDispatcher(t, dispatcher)

	updated, err := store.RunByRunID(ctx, storedRun.RunID)
	if err != nil {
		t.Fatalf("run by id: %v", err)
	}
	if updated.Status != run.StatusFailed || updated.Error != "startup failed" {
		t.Fatalf("updated run = %+v", updated)
	}
	if updated.RetryAfter == nil {
		t.Fatalf("retry after was not set: %+v", updated)
	}
	events, err := store.RunnerEvents(ctx, storedRun.RunID, 10)
	if err != nil {
		t.Fatalf("runner events: %v", err)
	}
	if !containsEvent(events, "retry_scheduled") {
		t.Fatalf("events = %+v", eventTypes(events))
	}
}

func TestDispatcherRecordsRetryExhaustedAtMaxAttempts(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	storedRun := createQueuedRunWithAttempt(t, store, 42, 3)
	testRunner := &recordingRunner{result: runner.Result{Status: run.StatusFailed, Error: "startup failed"}}
	dispatcher := newTestDispatcher(t, store, testRunner, []entity.Issue{{ID: 42, Status: entity.StatusReady}})

	if err := dispatcher.Dispatch(ctx, []run.Run{storedRun}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	shutdownDispatcher(t, dispatcher)

	updated, err := store.RunByRunID(ctx, storedRun.RunID)
	if err != nil {
		t.Fatalf("run by id: %v", err)
	}
	if updated.Status != run.StatusFailed || updated.RetryAfter != nil {
		t.Fatalf("updated run = %+v", updated)
	}
	events, err := store.RunnerEvents(ctx, storedRun.RunID, 10)
	if err != nil {
		t.Fatalf("runner events: %v", err)
	}
	if !containsEvent(events, "retry_exhausted") {
		t.Fatalf("events = %+v", eventTypes(events))
	}
}

func TestDispatcherRecoversPanickingRunner(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	storedRun := createQueuedRun(t, store, 42)
	testRunner := &recordingRunner{panicValue: "boom"}
	dispatcher := newTestDispatcher(t, store, testRunner, []entity.Issue{{ID: 42, Status: entity.StatusReady}})

	if err := dispatcher.Dispatch(ctx, []run.Run{storedRun}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	shutdownDispatcher(t, dispatcher)

	updated, err := store.RunByRunID(ctx, storedRun.RunID)
	if err != nil {
		t.Fatalf("run by id: %v", err)
	}
	if updated.Status != run.StatusFailed || !strings.Contains(updated.Error, "runner panic: boom") {
		t.Fatalf("updated run = %+v", updated)
	}
}

func TestDispatcherRespectsConcurrencyLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	runningRun := createQueuedRun(t, store, 1)
	var err error
	runningRun, err = store.UpdateRunStatus(ctx, runningRun.RunID, run.StatusRunning, "")
	if err != nil {
		t.Fatalf("mark running: %v", err)
	}
	queuedRun := createQueuedRun(t, store, 2)
	testRunner := &recordingRunner{result: runner.Result{Status: run.StatusSucceeded}}
	dispatcher := newTestDispatcherWithMax(t, store, testRunner, []entity.Issue{
		{ID: 1, Status: entity.StatusReady},
		{ID: 2, Status: entity.StatusReady},
	}, 1)

	if err := dispatcher.Dispatch(ctx, []run.Run{runningRun, queuedRun}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	shutdownDispatcher(t, dispatcher)

	if got := testRunner.runCount(); got != 0 {
		t.Fatalf("run count = %d", got)
	}
	updated, err := store.RunByRunID(ctx, queuedRun.RunID)
	if err != nil {
		t.Fatalf("run by id: %v", err)
	}
	if updated.Status != run.StatusQueued {
		t.Fatalf("queued run status = %s", updated.Status)
	}
}

func TestCalculateBackoff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{attempt: 1, want: 10 * time.Second},
		{attempt: 2, want: 20 * time.Second},
		{attempt: 3, want: 40 * time.Second},
		{attempt: 4, want: 80 * time.Second},
		{attempt: 5, want: 160 * time.Second},
	}
	for _, tt := range tests {
		if got := calculateBackoff(tt.attempt, 5*time.Minute); got != tt.want {
			t.Fatalf("attempt %d backoff = %s, want %s", tt.attempt, got, tt.want)
		}
	}
	if got := calculateBackoff(10, 30*time.Second); got != 30*time.Second {
		t.Fatalf("capped backoff = %s", got)
	}
}

func TestDispatcherReconcileCancelsQueuedRunForTerminalIssue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	storedRun := createQueuedRun(t, store, 42)
	dispatcher := newTestDispatcher(t, store, &recordingRunner{}, []entity.Issue{{ID: 42, Status: entity.StatusDone}})

	if err := dispatcher.Reconcile(ctx, []run.Run{storedRun}); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	updated, err := store.RunByRunID(ctx, storedRun.RunID)
	if err != nil {
		t.Fatalf("run by id: %v", err)
	}
	if updated.Status != run.StatusCancelled || updated.Error == "" {
		t.Fatalf("updated run = %+v", updated)
	}
}

func TestDispatcherReconcileCancelsRunningRunForTerminalIssue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	storedRun := createQueuedRun(t, store, 42)
	storedRun, err := store.UpdateRunStatus(ctx, storedRun.RunID, run.StatusRunning, "")
	if err != nil {
		t.Fatalf("mark running: %v", err)
	}
	dispatcher := newTestDispatcher(t, store, &recordingRunner{}, []entity.Issue{{ID: 42, Status: entity.StatusFailed}})
	cancelled := false
	dispatcher.registerCancel(storedRun.RunID, func() { cancelled = true })

	if err := dispatcher.Reconcile(ctx, []run.Run{storedRun}); err != nil {
		t.Fatalf("reconcile: %v", err)
	}

	if !cancelled {
		t.Fatal("cancel func was not called")
	}
	updated, err := store.RunByRunID(ctx, storedRun.RunID)
	if err != nil {
		t.Fatalf("run by id: %v", err)
	}
	if updated.Status != run.StatusCancelled {
		t.Fatalf("updated run = %+v", updated)
	}
}

func TestDispatcherDetectStallsFailsSilentRunner(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	storedRun := createQueuedRun(t, store, 42)
	storedRun, err := store.UpdateRunStatus(ctx, storedRun.RunID, run.StatusRunning, "")
	if err != nil {
		t.Fatalf("mark running: %v", err)
	}
	dispatcher := newTestDispatcher(t, store, &recordingRunner{}, []entity.Issue{{ID: 42, Status: entity.StatusReady}})
	dispatcher.workflowConfig.StallTimeout = time.Nanosecond

	if err := dispatcher.DetectStalls(ctx, []run.Run{storedRun}); err != nil {
		t.Fatalf("detect stalls: %v", err)
	}

	updated, err := store.RunByRunID(ctx, storedRun.RunID)
	if err != nil {
		t.Fatalf("run by id: %v", err)
	}
	if updated.Status != run.StatusFailed || !strings.Contains(updated.Error, "stall timeout") {
		t.Fatalf("updated run = %+v", updated)
	}
}

func TestDispatcherDetectStallsLeavesRecentRunnerRunning(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	storedRun := createQueuedRun(t, store, 42)
	storedRun, err := store.UpdateRunStatus(ctx, storedRun.RunID, run.StatusRunning, "")
	if err != nil {
		t.Fatalf("mark running: %v", err)
	}
	if err := store.RecordRunnerEvent(ctx, storedRun.RunID, "progress", "working", ""); err != nil {
		t.Fatalf("record event: %v", err)
	}
	dispatcher := newTestDispatcher(t, store, &recordingRunner{}, []entity.Issue{{ID: 42, Status: entity.StatusReady}})
	dispatcher.workflowConfig.StallTimeout = time.Hour

	if err := dispatcher.DetectStalls(ctx, []run.Run{storedRun}); err != nil {
		t.Fatalf("detect stalls: %v", err)
	}

	updated, err := store.RunByRunID(ctx, storedRun.RunID)
	if err != nil {
		t.Fatalf("run by id: %v", err)
	}
	if updated.Status != run.StatusRunning {
		t.Fatalf("updated run = %+v", updated)
	}
}

func createQueuedRun(t *testing.T, store *runstore.Store, issueID int64) run.Run {
	t.Helper()
	return createQueuedRunWithAttempt(t, store, issueID, 1)
}

func createQueuedRunWithAttempt(t *testing.T, store *runstore.Store, issueID int64, attempt int) run.Run {
	t.Helper()
	storedRun, err := store.CreateRun(context.Background(), runstore.CreateRunInput{
		IssueID:        issueID,
		Workspace:      filepath.Join(t.TempDir(), issueIdentifier(issueID)),
		Attempt:        attempt,
		OrchestratorID: "test-orchestrator",
	})
	if err != nil {
		t.Fatalf("create run: %v", err)
	}
	return storedRun
}

func newTestDispatcher(t *testing.T, store *runstore.Store, testRunner runner.Runner, issues []entity.Issue) *Dispatcher {
	t.Helper()
	return newTestDispatcherWithMax(t, store, testRunner, issues, 2)
}

func newTestDispatcherWithMax(t *testing.T, store *runstore.Store, testRunner runner.Runner, issues []entity.Issue, maxConcurrentRuns int) *Dispatcher {
	t.Helper()
	dispatcher, err := NewDispatcher(DispatcherConfig{
		Tracker:           fakeTracker{issues: issues},
		Store:             store,
		Runner:            testRunner,
		WorkflowConfig:    workflow.Config{MaxTurns: 2, MaxRetryAttempts: 3, MaxRetryBackoff: 5 * time.Minute, StallTimeout: 5 * time.Minute, CodexCommand: "codex app-server", CodexReadTimeout: time.Second, CodexTurnTimeout: time.Second},
		PromptTemplate:    "Do the task",
		MaxConcurrentRuns: maxConcurrentRuns,
	})
	if err != nil {
		t.Fatalf("new dispatcher: %v", err)
	}
	return dispatcher
}

func shutdownDispatcher(t *testing.T, dispatcher *Dispatcher) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := dispatcher.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown dispatcher: %v", err)
	}
}

func eventTypes(events []run.RunnerEvent) []string {
	output := make([]string, 0, len(events))
	for _, event := range events {
		output = append(output, event.EventType)
	}
	return output
}

func containsEvent(events []run.RunnerEvent, eventType string) bool {
	for _, event := range events {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}

type recordingRunner struct {
	mu         sync.Mutex
	tasks      []runner.Task
	result     runner.Result
	events     []runner.Event
	panicValue any
}

func (r *recordingRunner) Run(ctx context.Context, task runner.Task) runner.Result {
	r.mu.Lock()
	r.tasks = append(r.tasks, task)
	r.mu.Unlock()
	if r.panicValue != nil {
		panic(r.panicValue)
	}
	for _, event := range r.events {
		task.OnEvent(event)
	}
	return r.result
}

func (r *recordingRunner) runCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.tasks)
}

func (r *recordingRunner) firstTask(t *testing.T) runner.Task {
	t.Helper()
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.tasks) == 0 {
		t.Fatal("runner was not called")
	}
	return r.tasks[0]
}
