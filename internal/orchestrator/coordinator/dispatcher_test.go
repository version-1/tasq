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
			{EventType: "item/agentMessage/delta", Message: "token", PayloadJSON: `{"delta":"token"}`},
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
	if task.ResumeThreadID != "" {
		t.Fatalf("resume thread id = %q, want empty", task.ResumeThreadID)
	}
}

func TestDispatcherPersistsSessionStartedThreadID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	storedRun := createQueuedRun(t, store, 42)
	testRunner := &recordingRunner{
		result: runner.Result{Status: run.StatusSucceeded},
		events: []runner.Event{
			{EventType: "session_started", Message: "thread_id=thread-42"},
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
	if updated.ThreadID != "thread-42" {
		t.Fatalf("thread id = %q, want thread-42", updated.ThreadID)
	}
}

func TestDispatcherPassesLatestResumeThreadIDToRunner(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	previousRun, err := store.CreateRun(ctx, runstore.CreateRunInput{
		IssueID:        42,
		Workspace:      filepath.Join(t.TempDir(), "issue-42"),
		ThreadID:       "thread-previous",
		Attempt:        1,
		OrchestratorID: "test-orchestrator",
	})
	if err != nil {
		t.Fatalf("create previous run: %v", err)
	}
	if _, err := store.UpdateRunStatus(ctx, previousRun.RunID, run.StatusFailed, "blocked"); err != nil {
		t.Fatalf("mark previous failed: %v", err)
	}
	storedRun, err := store.CreateRun(ctx, runstore.CreateRunInput{
		IssueID:        42,
		Workspace:      filepath.Join(t.TempDir(), "issue-42"),
		Attempt:        2,
		OrchestratorID: "test-orchestrator",
	})
	if err != nil {
		t.Fatalf("create retry run: %v", err)
	}
	testRunner := &recordingRunner{result: runner.Result{Status: run.StatusSucceeded}}
	dispatcher := newTestDispatcher(t, store, testRunner, []entity.Issue{{ID: 42, Status: entity.StatusReady, Title: "Run task"}})

	if err := dispatcher.Dispatch(ctx, []run.Run{storedRun}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	shutdownDispatcher(t, dispatcher)

	task := testRunner.firstTask(t)
	if task.ResumeThreadID != "thread-previous" {
		t.Fatalf("resume thread id = %q, want thread-previous", task.ResumeThreadID)
	}
}

func TestDispatcherResolvesWorkflowForEachRunProject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	firstRun := createQueuedRun(t, store, 42)
	secondRun := createQueuedRun(t, store, 43)
	testRunner := &recordingRunner{result: runner.Result{Status: run.StatusSucceeded}}
	tracker := newFakeTracker([]entity.Issue{
		{ID: 42, ProjectID: 7, Status: entity.StatusReady, Title: "First"},
		{ID: 43, ProjectID: 8, Status: entity.StatusReady, Title: "Second"},
	})
	falseTaskWorkPrompt := false
	resolver := workflowResolverFunc(func(ctx context.Context, projectID int64) (workflow.Definition, error) {
		switch projectID {
		case 7:
			return workflow.Definition{
				Config: workflow.Config{
					Tasq:              workflow.TasqConfig{TaskWorkPrompt: &falseTaskWorkPrompt},
					MaxTurns:          3,
					ContinuationTurns: true,
					CodexCommand:      "codex app-server --project-a",
					CodexReadTimeout:  2 * time.Second,
					CodexTurnTimeout:  3 * time.Second,
				},
				PromptTemplate: "Project A prompt",
			}, nil
		case 8:
			return workflow.Definition{
				Config: workflow.Config{
					MaxTurns:         5,
					CodexCommand:     "codex app-server --project-b",
					CodexReadTimeout: 4 * time.Second,
					CodexTurnTimeout: 5 * time.Second,
				},
				PromptTemplate: "Project B prompt",
			}, nil
		default:
			t.Fatalf("unexpected projectID = %d", projectID)
			return workflow.Definition{}, nil
		}
	})
	dispatcher, err := NewDispatcher(DispatcherConfig{
		Tracker:           tracker,
		Store:             store,
		Runner:            testRunner,
		WorkflowResolver:  resolver,
		MaxConcurrentRuns: 2,
	})
	if err != nil {
		t.Fatalf("new dispatcher: %v", err)
	}

	if err := dispatcher.Dispatch(ctx, []run.Run{firstRun, secondRun}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	shutdownDispatcher(t, dispatcher)

	tasks := testRunner.tasksByIssueID()
	firstTask := tasks[42]
	if firstTask.PromptTemplate != "Project A prompt" || firstTask.MaxTurns != 3 || !firstTask.ContinueTurns || firstTask.Command != "codex app-server --project-a" {
		t.Fatalf("first task workflow = %+v", firstTask)
	}
	if firstTask.TaskWorkPrompt == nil || *firstTask.TaskWorkPrompt {
		t.Fatalf("first task task_work_prompt = %v", firstTask.TaskWorkPrompt)
	}
	secondTask := tasks[43]
	if secondTask.PromptTemplate != "Project B prompt" || secondTask.MaxTurns != 5 || secondTask.ContinueTurns || secondTask.Command != "codex app-server --project-b" {
		t.Fatalf("second task workflow = %+v", secondTask)
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

func TestDispatcherIncludesApprovalRequestDetailsInBlockerComment(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	storedRun := createQueuedRun(t, store, 42)
	errText := `approval_required: tasq denied app-server approval request by policy

method: item/commandExecution/requestApproval
payload: {"command":"make test","approvalId":"approval-1","reason":"needs command approval"}`
	testRunner := &recordingRunner{result: runner.Result{Status: run.StatusFailed, Error: errText}}
	tracker := newFakeTracker([]entity.Issue{{ID: 42, Status: entity.StatusReady}})
	dispatcher := newTestDispatcherWithTracker(t, store, testRunner, tracker, 2)

	if err := dispatcher.Dispatch(ctx, []run.Run{storedRun}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	shutdownDispatcher(t, dispatcher)

	issue := tracker.issue(t, 42)
	if issue.Status != entity.StatusBlocked {
		t.Fatalf("issue status = %s", issue.Status)
	}
	comments := tracker.commentsForIssue(t, 42)
	if len(comments) != 1 {
		t.Fatalf("comments = %+v", comments)
	}
	for _, want := range []string{"approval_required", "item/commandExecution/requestApproval", "make test", "approval-1", "needs command approval"} {
		if !strings.Contains(comments[0].Body, want) {
			t.Fatalf("comment body missing %q: %s", want, comments[0].Body)
		}
	}
}

func TestDispatcherBlocksInProgressIssueWhenRunFails(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	storedRun := createQueuedRun(t, store, 42)
	testRunner := &recordingRunner{result: runner.Result{Status: run.StatusFailed, Error: "approval required"}}
	tracker := newFakeTracker([]entity.Issue{{ID: 42, Status: entity.StatusInProgress}})
	dispatcher := newTestDispatcherWithTracker(t, store, testRunner, tracker, 2)

	if err := dispatcher.Dispatch(ctx, []run.Run{storedRun}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	shutdownDispatcher(t, dispatcher)

	issue := tracker.issue(t, 42)
	if issue.Status != entity.StatusBlocked {
		t.Fatalf("issue status = %s", issue.Status)
	}
	comments := tracker.commentsForIssue(t, 42)
	if len(comments) != 1 || comments[0].Type != entity.CommentBlocker || !strings.Contains(comments[0].Body, "approval required") {
		t.Fatalf("comments = %+v", comments)
	}
}

func TestDispatcherLeavesReviewIssueWhenRunFails(t *testing.T) {
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

func TestShouldKeepRunnerEventIgnoresAgentMessageDelta(t *testing.T) {
	t.Parallel()

	if shouldKeepRunnerEvent(runner.Event{EventType: "item/agentMessage/delta"}) {
		t.Fatal("agent message delta event should not be kept")
	}
	if !shouldKeepRunnerEvent(runner.Event{EventType: "turn/completed"}) {
		t.Fatal("turn completed event should be kept")
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
		WorkflowResolver:  staticWorkflowResolver(testWorkflowDefinition()),
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

func (r *recordingRunner) tasksByIssueID() map[int64]runner.Task {
	r.mu.Lock()
	defer r.mu.Unlock()
	output := make(map[int64]runner.Task, len(r.tasks))
	for _, task := range r.tasks {
		output[task.Issue.ID] = task
	}
	return output
}

type workflowResolverFunc func(ctx context.Context, projectID int64) (workflow.Definition, error)

func (f workflowResolverFunc) Resolve(ctx context.Context, projectID int64) (workflow.Definition, error) {
	return f(ctx, projectID)
}

func staticWorkflowResolver(definition workflow.Definition) WorkflowResolver {
	return workflowResolverFunc(func(ctx context.Context, projectID int64) (workflow.Definition, error) {
		return definition, nil
	})
}

func testWorkflowDefinition() workflow.Definition {
	return workflow.Definition{
		Config: workflow.Config{
			MaxTurns:         2,
			CodexCommand:     "codex app-server",
			CodexReadTimeout: time.Second,
			CodexTurnTimeout: time.Second,
		},
		PromptTemplate: "Do the task",
	}
}
