package coordinator

import (
	"context"
	"database/sql"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
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

func TestPollQueuesWorkspaceUnderIssueProjectLocation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	manager := newTestWorkspaceManager(t)
	projectLocation := newTestProjectRepo(t)
	tracker := newFakeTracker([]entity.Issue{
		{ID: 42, ProjectID: 7, ProjectKey: "OTHER", Status: entity.StatusReady, Title: "Build in project repo"},
	})
	tracker.setProject(entity.Project{
		ID:       7,
		Key:      "OTHER",
		Name:     "Other project",
		Location: projectLocation,
	})
	poller := newTestPollerWithTracker(t, store, manager, tracker)

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
	wantWorkspace := filepath.Join(canonicalTestPath(t, projectLocation), ".worktrees", "issue-42")
	if runs[0].Workspace != wantWorkspace {
		t.Fatalf("workspace = %q, want %q", runs[0].Workspace, wantWorkspace)
	}
	if _, err := os.Stat(filepath.Join(runs[0].Workspace, ".git")); err != nil {
		t.Fatalf("workspace git file missing: %v", err)
	}
}

func TestPollUsesResolvedWorkflowForWorkspaceRootAndHooks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	manager := newTestWorkspaceManager(t)
	projectLocation := newTestProjectRepo(t)
	tracker := newFakeTracker([]entity.Issue{
		{ID: 42, ProjectID: 7, ProjectKey: "OTHER", Status: entity.StatusReady, Title: "Build in project repo"},
	})
	tracker.setProject(entity.Project{
		ID:       7,
		Key:      "OTHER",
		Name:     "Other project",
		Location: projectLocation,
	})
	resolver := workflowResolverFunc(func(ctx context.Context, projectID int64) (workflow.Definition, error) {
		return workflow.Definition{
			Path: filepath.Join(projectLocation, "WORKFLOW.md"),
			Config: workflow.Config{
				WorkspaceRoot:   filepath.Join(projectLocation, ".custom-worktrees"),
				HookAfterCreate: `echo project-hook > project-hook.out`,
				HookTimeout:     time.Second,
			},
			PromptTemplate: "Project prompt",
		}, nil
	})
	poller, err := NewPoller(PollerConfig{
		Tracker:        tracker,
		Store:          store,
		Workspaces:     manager,
		Workflow:       resolver,
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

	storedRun, err := store.RunByIssueID(ctx, 42)
	if err != nil {
		t.Fatalf("run by issue id: %v", err)
	}
	wantWorkspace := filepath.Join(canonicalTestPath(t, projectLocation), ".custom-worktrees", "issue-42")
	if storedRun.Workspace != wantWorkspace {
		t.Fatalf("workspace = %q, want %q", storedRun.Workspace, wantWorkspace)
	}
	content, err := os.ReadFile(filepath.Join(storedRun.Workspace, "project-hook.out"))
	if err != nil {
		t.Fatalf("read project hook output: %v", err)
	}
	if strings.TrimSpace(string(content)) != "project-hook" {
		t.Fatalf("hook output = %q", content)
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

func TestPollIgnoresPendingQueueIssues(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	manager := newTestWorkspaceManager(t)
	tracker := newFakeTracker([]entity.Issue{
		{ID: 42, Status: entity.StatusReady, Title: "Pending dependency"},
		{ID: 43, Status: entity.StatusReady, Title: "Ready dependency"},
	})
	tracker.setPendingIssueIDs(42)
	tracker.setProjectLocation(1, manager.RepoRoot())
	poller := newTestPollerWithTracker(t, store, manager, tracker)

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
	if runs[0].IssueID != 43 {
		t.Fatalf("queued issue id = %d, want 43", runs[0].IssueID)
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
	tracker.setProjectLocation(1, manager.RepoRoot())
	dispatcher := newTestDispatcherWithTracker(t, store, testRunner, tracker, 2)
	poller, err := NewPoller(PollerConfig{
		Tracker:        tracker,
		Store:          store,
		Workspaces:     manager,
		Workflow:       workflowResolverForManager(manager),
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
	tracker.setProjectLocation(1, manager.RepoRoot())
	dispatcher := newTestDispatcherWithTracker(t, store, testRunner, tracker, 2)
	poller, err := NewPoller(PollerConfig{
		Tracker:        tracker,
		Store:          store,
		Workspaces:     manager,
		Workflow:       workflowResolverForManager(manager),
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
	mu             sync.Mutex
	issues         map[int64]entity.Issue
	projects       map[int64]entity.Project
	comments       map[int64][]entity.Comment
	changeRequests map[int64]entity.ChangeRequest
	pending        map[int64]struct{}
}

func newFakeTracker(issues []entity.Issue) *fakeTracker {
	tracker := &fakeTracker{
		issues:         make(map[int64]entity.Issue, len(issues)),
		projects:       make(map[int64]entity.Project),
		comments:       make(map[int64][]entity.Comment),
		changeRequests: make(map[int64]entity.ChangeRequest),
		pending:        make(map[int64]struct{}),
	}
	for _, issue := range issues {
		if issue.ProjectID == 0 {
			issue.ProjectID = 1
		}
		if issue.ProjectKey == "" {
			issue.ProjectKey = "TEST"
		}
		tracker.issues[issue.ID] = issue
		if _, ok := tracker.projects[issue.ProjectID]; !ok {
			tracker.projects[issue.ProjectID] = entity.Project{
				ID:  issue.ProjectID,
				Key: issue.ProjectKey,
			}
		}
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

func (t *fakeTracker) Project(ctx context.Context, id int64) (entity.Project, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	project, ok := t.projects[id]
	if ok {
		return project, nil
	}
	return entity.Project{}, sql.ErrNoRows
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

func (t *fakeTracker) Queue(ctx context.Context) (entity.Queue, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	queue := entity.Queue{Queued: []entity.QueueIssue{}, Pending: []entity.QueueIssue{}}
	for _, issue := range t.issues {
		if issue.Status != entity.StatusReady {
			continue
		}
		queueIssue := entity.QueueIssue{Issue: issue}
		if _, ok := t.pending[issue.ID]; ok {
			queue.Pending = append(queue.Pending, queueIssue)
			continue
		}
		queue.Queued = append(queue.Queued, queueIssue)
	}
	return queue, nil
}

func (t *fakeTracker) setPendingIssueIDs(ids ...int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, id := range ids {
		t.pending[id] = struct{}{}
	}
}

func (t *fakeTracker) setProject(project entity.Project) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.projects[project.ID] = project
}

func (t *fakeTracker) setProjectLocation(id int64, location string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	project := t.projects[id]
	project.ID = id
	if project.Key == "" {
		project.Key = "TEST"
	}
	project.Location = location
	t.projects[id] = project
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

func (t *fakeTracker) ChangeRequestsByIssueID(ctx context.Context, issueID int64, status entity.ChangeRequestStatus, limit int) ([]entity.ChangeRequest, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if _, ok := t.issues[issueID]; !ok {
		return nil, sql.ErrNoRows
	}
	var requests []entity.ChangeRequest
	for _, request := range t.changeRequests {
		if request.IssueID != issueID {
			continue
		}
		if status != "" && request.Status != status {
			continue
		}
		requests = append(requests, request)
	}
	sort.SliceStable(requests, func(i, j int) bool {
		if requests[i].CreatedAt.Equal(requests[j].CreatedAt) {
			return requests[i].ID < requests[j].ID
		}
		return requests[i].CreatedAt.Before(requests[j].CreatedAt)
	})
	if limit > 0 && len(requests) > limit {
		requests = requests[:limit]
	}
	return requests, nil
}

func (t *fakeTracker) UpdateChangeRequest(ctx context.Context, id int64, input entity.UpdateChangeRequestInput) (entity.ChangeRequest, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	request, ok := t.changeRequests[id]
	if !ok {
		return entity.ChangeRequest{}, sql.ErrNoRows
	}
	if input.Status != nil {
		request.Status = *input.Status
	}
	if input.Body != nil {
		request.Body = *input.Body
	}
	if input.ResolvedByRunID != nil {
		request.ResolvedByRunID = input.ResolvedByRunID
	}
	if input.ResultCommentID != nil {
		request.ResultCommentID = input.ResultCommentID
	}
	t.changeRequests[id] = request
	return request, nil
}

func (t *fakeTracker) addChangeRequest(request entity.ChangeRequest) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if request.ID == 0 {
		request.ID = int64(len(t.changeRequests) + 1)
	}
	if request.Status == "" {
		request.Status = entity.ChangeRequestOpen
	}
	t.changeRequests[request.ID] = request
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
	store, err := runstore.OpenMigrated(context.Background(), filepath.Join(t.TempDir(), "orchestrator.sqlite"))
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
	repoRoot := initTestGitRepo(t)
	manager, err := workspace.NewManager(filepath.Join(repoRoot, ".worktrees"))
	if err != nil {
		t.Fatalf("new workspace manager: %v", err)
	}
	return manager
}

func newTestProjectRepo(t *testing.T) string {
	t.Helper()
	return initTestGitRepo(t)
}

func initTestGitRepo(t *testing.T) string {
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
	return repoRoot
}

func canonicalTestPath(t *testing.T, path string) string {
	t.Helper()
	cleanPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("canonicalize test path %q: %v", path, err)
	}
	return filepath.Clean(cleanPath)
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
	tracker := newFakeTracker(issues)
	tracker.setProjectLocation(1, manager.RepoRoot())
	return newTestPollerWithTracker(t, store, manager, tracker)
}

func newTestPollerWithTracker(t *testing.T, store *runstore.Store, manager *workspace.Manager, tracker *fakeTracker) *Poller {
	t.Helper()
	poller, err := NewPoller(PollerConfig{
		Tracker:        tracker,
		Store:          store,
		Workspaces:     manager,
		Workflow:       workflowResolverForManager(manager),
		Interval:       time.Minute,
		MaxActiveRuns:  2,
		OrchestratorID: "test-orchestrator",
	})
	if err != nil {
		t.Fatalf("new poller: %v", err)
	}
	return poller
}

func workflowResolverForManager(manager *workspace.Manager) WorkflowResolver {
	definition := testWorkflowDefinition()
	definition.Path = filepath.Join(manager.RepoRoot(), "WORKFLOW.md")
	definition.Config.WorkspaceRoot = manager.Root()
	return staticWorkflowResolver(definition)
}
