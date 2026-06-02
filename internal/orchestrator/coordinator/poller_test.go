package coordinator

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/orchestrator/run"
	"github.com/version-1/tasq/internal/orchestrator/runner"
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

func TestPollQueuesAndDispatchesInSameCycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	manager := newTestWorkspaceManager(t)
	testRunner := &recordingRunner{result: runner.Result{Status: run.StatusSucceeded}}
	tracker := newFakeTracker([]entity.Issue{{ID: 42, Status: entity.StatusReady, Title: "Build dispatch"}})
	dispatcher := newTestDispatcherWithTracker(t, store, testRunner, tracker, 2)
	poller, err := NewPoller(PollerConfig{
		Tracker:        tracker,
		Store:          store,
		Workspaces:     manager,
		Dispatcher:     dispatcher,
		Interval:       time.Minute,
		MaxActiveRuns:  2,
		OrchestratorID: "test-orchestrator",
	})
	if err != nil {
		t.Fatalf("new poller: %v", err)
	}

	if err := poller.Poll(ctx); err != nil {
		t.Fatalf("poll: %v", err)
	}
	shutdownDispatcher(t, dispatcher)

	storedRun, err := store.RunByIssueID(ctx, 42)
	if err != nil {
		t.Fatalf("run by issue id: %v", err)
	}
	if storedRun.Status != run.StatusSucceeded {
		t.Fatalf("run status = %s", storedRun.Status)
	}
	if got := testRunner.runCount(); got != 1 {
		t.Fatalf("run count = %d", got)
	}
}

func TestPollFailedDispatchBlocksIssueAndSkipsNextPoll(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	manager := newTestWorkspaceManager(t)
	testRunner := &recordingRunner{result: runner.Result{Status: run.StatusFailed, Error: "codex command not found"}}
	tracker := newFakeTracker([]entity.Issue{{ID: 42, Status: entity.StatusReady, Title: "Build dispatch"}})
	dispatcher := newTestDispatcherWithTracker(t, store, testRunner, tracker, 2)
	poller, err := NewPoller(PollerConfig{
		Tracker:        tracker,
		Store:          store,
		Workspaces:     manager,
		Dispatcher:     dispatcher,
		Interval:       time.Minute,
		MaxActiveRuns:  2,
		OrchestratorID: "test-orchestrator",
	})
	if err != nil {
		t.Fatalf("new poller: %v", err)
	}

	if err := poller.Poll(ctx); err != nil {
		t.Fatalf("poll: %v", err)
	}
	shutdownDispatcher(t, dispatcher)
	if err := poller.Poll(ctx); err != nil {
		t.Fatalf("poll again: %v", err)
	}

	if issue := tracker.issue(t, 42); issue.Status != entity.StatusBlocked {
		t.Fatalf("issue status = %s", issue.Status)
	}
	storedRun, err := store.RunByIssueID(ctx, 42)
	if err != nil {
		t.Fatalf("run by issue id: %v", err)
	}
	if storedRun.Status != run.StatusFailed {
		t.Fatalf("run status = %s", storedRun.Status)
	}
	if got := testRunner.runCount(); got != 1 {
		t.Fatalf("run count = %d", got)
	}
}

type fakeTracker struct {
	mu       sync.Mutex
	issues   map[int64]entity.Issue
	comments map[int64][]entity.Comment
}

func newFakeTracker(issues []entity.Issue) *fakeTracker {
	tracker := &fakeTracker{
		issues:   make(map[int64]entity.Issue, len(issues)),
		comments: make(map[int64][]entity.Comment),
	}
	for _, issue := range issues {
		tracker.issues[issue.ID] = issue
	}
	return tracker
}

func (t *fakeTracker) Issue(ctx context.Context, id int64) (entity.Issue, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	issue, ok := t.issues[id]
	if ok {
		return issue, nil
	}
	return entity.Issue{}, sql.ErrNoRows
}

func (t *fakeTracker) IssuesByStates(ctx context.Context, states []string) ([]entity.Issue, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	stateSet := make(map[entity.Status]struct{}, len(states))
	for _, state := range states {
		stateSet[entity.Status(state)] = struct{}{}
	}
	output := make([]entity.Issue, 0, len(t.issues))
	for _, issue := range t.issues {
		if len(stateSet) == 0 {
			output = append(output, issue)
			continue
		}
		if _, ok := stateSet[issue.Status]; ok {
			output = append(output, issue)
		}
	}
	return output, nil
}

func (t *fakeTracker) UpdateIssue(ctx context.Context, id int64, input entity.UpdateIssueInput) (entity.Issue, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	issue, ok := t.issues[id]
	if !ok {
		return entity.Issue{}, sql.ErrNoRows
	}
	if input.Status != nil {
		issue.Status = *input.Status
	}
	if input.Assignee != nil {
		issue.Assignee = *input.Assignee
	}
	t.issues[id] = issue
	return issue, nil
}

func (t *fakeTracker) CreateComment(ctx context.Context, issueID int64, input entity.CreateCommentInput) (entity.Comment, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.issues[issueID]; !ok {
		return entity.Comment{}, sql.ErrNoRows
	}
	comment := entity.Comment{
		ID:      int64(len(t.comments[issueID]) + 1),
		IssueID: issueID,
		Author:  input.Author,
		Type:    input.Type,
		Body:    input.Body,
	}
	t.comments[issueID] = append(t.comments[issueID], comment)
	return comment, nil
}

func (t *fakeTracker) issue(tb testing.TB, id int64) entity.Issue {
	tb.Helper()
	issue, err := t.Issue(context.Background(), id)
	if err != nil {
		tb.Fatalf("issue %d: %v", id, err)
	}
	return issue
}

func (t *fakeTracker) commentsForIssue(tb testing.TB, id int64) []entity.Comment {
	tb.Helper()
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]entity.Comment(nil), t.comments[id]...)
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
	repoRoot := t.TempDir()
	gitCommand(t, repoRoot, "init")
	gitCommand(t, repoRoot, "config", "user.email", "test@example.com")
	gitCommand(t, repoRoot, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(repoRoot, "README.md"), []byte("test repo\n"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}
	gitCommand(t, repoRoot, "add", "README.md")
	gitCommand(t, repoRoot, "commit", "-m", "initial commit")
	manager, err := workspace.NewManager(filepath.Join(repoRoot, ".worktrees"))
	if err != nil {
		t.Fatalf("new workspace manager: %v", err)
	}
	return manager
}

func gitCommand(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	return string(output)
}

func newTestPoller(t *testing.T, store *runstore.Store, manager *workspace.Manager, issues []entity.Issue) *Poller {
	t.Helper()
	poller, err := NewPoller(PollerConfig{
		Tracker:        newFakeTracker(issues),
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
