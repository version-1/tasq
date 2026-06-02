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
	tracker := newFakeTracker([]entity.Issue{{ID: 42, Status: entity.StatusReady}})
	dispatcher := newTestDispatcherWithTracker(t, store, testRunner, tracker, 2)

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
	issue := tracker.issue(t, 42)
	if issue.Status != entity.StatusBlocked {
		t.Fatalf("issue status = %s", issue.Status)
	}
	comments := tracker.commentsForIssue(t, 42)
	if len(comments) != 1 || comments[0].Type != entity.CommentBlocker || !strings.Contains(comments[0].Body, "startup failed") {
		t.Fatalf("comments = %+v", comments)
	}
}

func TestDispatcherLeavesNonReadyIssueWhenRunFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	storedRun := createQueuedRun(t, store, 42)
	testRunner := &recordingRunner{result: runner.Result{Status: run.StatusFailed, Error: "startup failed"}}
	tracker := newFakeTracker([]entity.Issue{{ID: 42, Status: entity.StatusReview}})
	dispatcher := newTestDispatcherWithTracker(t, store, testRunner, tracker, 2)

	if err := dispatcher.Dispatch(ctx, []run.Run{storedRun}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	shutdownDispatcher(t, dispatcher)

	issue := tracker.issue(t, 42)
	if issue.Status != entity.StatusReview {
		t.Fatalf("issue status = %s", issue.Status)
	}
	if comments := tracker.commentsForIssue(t, 42); len(comments) != 0 {
		t.Fatalf("comments = %+v", comments)
	}
}

func TestDispatcherRecoversPanickingRunner(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	storedRun := createQueuedRun(t, store, 42)
	testRunner := &recordingRunner{panicValue: "boom"}
	tracker := newFakeTracker([]entity.Issue{{ID: 42, Status: entity.StatusReady}})
	dispatcher := newTestDispatcherWithTracker(t, store, testRunner, tracker, 2)

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
	issue := tracker.issue(t, 42)
	if issue.Status != entity.StatusBlocked {
		t.Fatalf("issue status = %s", issue.Status)
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

func TestFormatRunnerEventLog(t *testing.T) {
	t.Parallel()

	line := formatRunnerEventLog("run-1", runner.Event{
		EventType:   "turn/completed",
		Message:     "done",
		PayloadJSON: `{"ok":true}`,
	})

	want := `orchestrator runner event run=run-1 event=turn/completed message="done" payload={"ok":true}`
	if line != want {
		t.Fatalf("line = %q, want %q", line, want)
	}
}

func TestFormatRunnerEventLogTruncatesPayload(t *testing.T) {
	t.Parallel()

	line := formatRunnerEventLog("run-1", runner.Event{
		EventType:   "notification",
		Message:     "large payload",
		PayloadJSON: strings.Repeat("a", maxRunnerEventLogPayloadLength+1),
	})

	if !strings.Contains(line, strings.Repeat("a", maxRunnerEventLogPayloadLength)+"... truncated") {
		t.Fatalf("line did not contain truncated payload: %q", line)
	}
}

func TestShouldLogRunnerEventIgnoresAgentMessageDelta(t *testing.T) {
	t.Parallel()

	if shouldLogRunnerEvent(runner.Event{EventType: "item/agentMessage/delta"}) {
		t.Fatal("agent message delta event should not be logged")
	}
	if !shouldLogRunnerEvent(runner.Event{EventType: "turn/completed"}) {
		t.Fatal("turn completed event should be logged")
	}
}

func createQueuedRun(t *testing.T, store *runstore.Store, issueID int64) run.Run {
	t.Helper()
	storedRun, err := store.CreateRun(context.Background(), runstore.CreateRunInput{
		IssueID:        issueID,
		Workspace:      filepath.Join(t.TempDir(), issueIdentifier(issueID)),
		Attempt:        1,
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
	return newTestDispatcherWithTracker(t, store, testRunner, newFakeTracker(issues), maxConcurrentRuns)
}

func newTestDispatcherWithTracker(t *testing.T, store *runstore.Store, testRunner runner.Runner, tracker *fakeTracker, maxConcurrentRuns int) *Dispatcher {
	t.Helper()
	dispatcher, err := NewDispatcher(DispatcherConfig{
		Tracker:           tracker,
		Store:             store,
		Runner:            testRunner,
		WorkflowConfig:    workflow.Config{MaxTurns: 2, CodexCommand: "codex app-server", CodexReadTimeout: time.Second, CodexTurnTimeout: time.Second},
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
