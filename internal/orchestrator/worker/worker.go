package worker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/orchestrator/run"
	"github.com/version-1/tasq/internal/orchestrator/runner"
	"github.com/version-1/tasq/internal/orchestrator/runstore"
	"github.com/version-1/tasq/internal/orchestrator/workflow"
	"github.com/version-1/tasq/internal/orchestrator/workspace"
)

type Worker struct {
	store             *runstore.Store
	client            trackerClient
	orchestratorID    string
	pollInterval      time.Duration
	leaseSeconds      int
	maxConcurrent     int
	maxTurns          int
	continuationTurns bool
	maxRetryAttempts  int
	maxRetryBackoff   time.Duration
	stallTimeout      time.Duration
	promptTemplate    string
	codexCommand      string
	codexReadTimeout  time.Duration
	codexTurnTimeout  time.Duration
	runner            runner.Runner
	workspaceManager  *workspace.Manager
	running           map[int64]runningRun
	claimed           map[int64]struct{}
	refreshCh         chan struct{}
}

type trackerClient interface {
	ClaimWorkItem(ctx context.Context, orchestratorID string, leaseSeconds int) (*entity.WorkItem, error)
	RenewWorkItemLease(ctx context.Context, input entity.RenewWorkItemLeaseInput) error
	SendRunEvent(ctx context.Context, event run.OutboxEvent) error
	Issue(ctx context.Context, id int64) (entity.Issue, error)
	Issues(ctx context.Context) ([]entity.Issue, error)
}

type runningRun struct {
	run       run.Run
	workItem  entity.WorkItem
	startedAt time.Time
}

func NewWithConfig(store *runstore.Store, client trackerClient, orchestratorID string, leaseSeconds int, definition workflow.Definition, agentRunner runner.Runner) (*Worker, error) {
	config := definition.Config
	if config.PollInterval <= 0 {
		config.PollInterval = 3 * time.Second
	}
	if leaseSeconds <= 0 {
		leaseSeconds = 30
	}
	if config.MaxConcurrentRuns <= 0 {
		config.MaxConcurrentRuns = 1
	}
	if config.MaxTurns <= 0 {
		config.MaxTurns = 1
	}
	if agentRunner == nil {
		agentRunner = runner.SimulatedRunner{}
	}
	workspaceManager, err := workspace.NewManagerWithSource(config.WorkspaceRoot, config.WorkspaceSource)
	if err != nil {
		return nil, err
	}
	return &Worker{
		store:             store,
		client:            client,
		orchestratorID:    orchestratorID,
		pollInterval:      config.PollInterval,
		leaseSeconds:      leaseSeconds,
		maxConcurrent:     config.MaxConcurrentRuns,
		maxTurns:          config.MaxTurns,
		continuationTurns: config.ContinuationTurns,
		maxRetryAttempts:  config.MaxRetryAttempts,
		maxRetryBackoff:   config.MaxRetryBackoff,
		stallTimeout:      config.StallTimeout,
		promptTemplate:    definition.PromptTemplate,
		codexCommand:      config.CodexCommand,
		codexReadTimeout:  config.CodexReadTimeout,
		codexTurnTimeout:  config.CodexTurnTimeout,
		runner:            agentRunner,
		workspaceManager:  workspaceManager,
		running:           map[int64]runningRun{},
		claimed:           map[int64]struct{}{},
		refreshCh:         make(chan struct{}, 1),
	}, nil
}

func (w *Worker) RequestRefresh() {
	select {
	case w.refreshCh <- struct{}{}:
	default:
	}
}

func (w *Worker) Run(ctx context.Context) {
	w.cleanupTerminalWorkspaces(ctx)
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	w.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.refreshCh:
			w.tick(ctx)
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	if err := w.flushOutbox(ctx); err != nil {
		log.Printf("flush outbox: %v", err)
	}
	w.logOperatorVisibility(ctx)
	item, err := w.client.ClaimWorkItem(ctx, w.orchestratorID, w.leaseSeconds)
	if err != nil {
		log.Printf("claim work item: %v", err)
		return
	}
	if item == nil {
		return
	}
	if !w.canDispatch(item.IssueID) {
		log.Printf("skip claimed work item issue_id=%d", item.IssueID)
		return
	}
	defer func() {
		if _, ok := w.running[item.IssueID]; !ok {
			delete(w.claimed, item.IssueID)
		}
	}()
	workspace, err := w.workspaceManager.CreateForIssue(fmt.Sprintf("issue-%d", item.IssueID))
	if err != nil {
		log.Printf("create workspace issue_id=%d: %v", item.IssueID, err)
		if recordErr := w.store.RecordWorkspaceSetupFailure(ctx, item.IssueID, fmt.Sprintf("issue-%d", item.IssueID), "", err.Error()); recordErr != nil {
			log.Printf("record workspace setup failure issue_id=%d: %v", item.IssueID, recordErr)
		}
		return
	}
	if err := w.store.UpsertWorkspaceMetadata(ctx, runstore.WorkspaceMetadataInput{
		WorkspaceKey: workspace.WorkspaceKey,
		IssueID:      item.IssueID,
		Path:         workspace.Path,
		CreatedNow:   workspace.CreatedNow,
		SourcePath:   w.workspaceManager.Source(),
	}); err != nil {
		log.Printf("record workspace metadata issue_id=%d workspace=%s: %v", item.IssueID, workspace.Path, err)
	}
	createdRun, err := w.store.CreateRun(ctx, runstore.CreateRunInput{
		IssueID:        item.IssueID,
		WorkItemID:     item.ID,
		ClaimToken:     item.ClaimToken,
		Workspace:      workspace.Path,
		Attempt:        item.Attempt,
		OrchestratorID: w.orchestratorID,
	})
	if err != nil {
		log.Printf("create run: %v", err)
		return
	}
	w.markRunning(createdRun, *item)
	defer w.release(item.IssueID)

	if err := w.flushOutbox(ctx); err != nil {
		log.Printf("flush outbox: %v", err)
	}
	if _, err := w.store.UpdateRunStatus(ctx, createdRun.RunID, run.StatusRunning, ""); err != nil {
		log.Printf("mark run running: %v", err)
		return
	}
	if err := w.flushOutbox(ctx); err != nil {
		log.Printf("flush outbox: %v", err)
	}
	result := w.runWithRetries(ctx, createdRun.RunID, *item, workspace)
	if result.Status == "" {
		result.Status = run.StatusSucceeded
	}
	if _, err := w.store.UpdateRunStatus(ctx, createdRun.RunID, result.Status, result.Error); err != nil {
		log.Printf("mark run finished: %v", err)
		return
	}
	if err := w.flushOutbox(ctx); err != nil {
		log.Printf("flush outbox: %v", err)
	}
	if result.Status == run.StatusFailed || result.Status == run.StatusCancelled {
		w.cleanupWorkspace(ctx, workspace.WorkspaceKey, fmt.Sprintf("issue-%d", item.IssueID))
	}
}

func (w *Worker) runWithRetries(ctx context.Context, runID string, item entity.WorkItem, workspace workspace.Workspace) runner.Result {
	maxAttempts := w.maxRetryAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	var result runner.Result
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result = w.runWithLeaseRenewal(ctx, runID, item, workspace)
		if result.Status == run.StatusSucceeded {
			return result
		}
		if attempt == maxAttempts || result.Status == run.StatusCancelled {
			break
		}
		delay := w.retryDelay(attempt)
		_ = w.store.RecordRunnerEvent(context.Background(), runID, "retrying", fmt.Sprintf("attempt=%d delay=%s error=%s", attempt+1, delay, result.Error), "")
		select {
		case <-ctx.Done():
			return runner.Result{Status: run.StatusCancelled, Error: ctx.Err().Error()}
		case <-time.After(delay):
		}
	}
	if result.Status == run.StatusFailed {
		if result.Error == "" {
			result.Error = "manual intervention required: retry attempts exhausted"
		} else {
			result.Error = "manual intervention required: retry attempts exhausted: " + result.Error
		}
	}
	return result
}

func (w *Worker) runWithLeaseRenewal(ctx context.Context, runID string, item entity.WorkItem, workspace workspace.Workspace) runner.Result {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan runner.Result, 1)
	lastEventAt := time.Now().UTC()
	var lastEventMu sync.Mutex
	go func() {
		done <- w.runner.Run(runCtx, runner.Task{
			WorkItem:       item,
			RunID:          runID,
			Workspace:      workspace,
			PromptTemplate: w.promptTemplate,
			MaxTurns:       w.maxTurns,
			ContinueTurns:  w.continuationTurns,
			Command:        w.codexCommand,
			ReadTimeout:    w.codexReadTimeout,
			TurnTimeout:    w.codexTurnTimeout,
			OnEvent: func(event runner.Event) {
				lastEventMu.Lock()
				lastEventAt = time.Now().UTC()
				lastEventMu.Unlock()
				if err := w.store.RecordRunnerEvent(context.Background(), runID, event.EventType, event.Message, event.PayloadJSON); err != nil {
					log.Printf("record runner event run_id=%s event_type=%s: %v", runID, event.EventType, err)
				}
			},
		})
	}()
	renewInterval := time.Duration(w.leaseSeconds) * time.Second / 2
	if renewInterval <= 0 {
		renewInterval = 15 * time.Second
	}
	ticker := time.NewTicker(renewInterval)
	defer ticker.Stop()
	for {
		select {
		case result := <-done:
			return result
		case <-ctx.Done():
			cancel()
			return <-done
		case <-ticker.C:
			if w.isStalled(&lastEventAt, &lastEventMu) {
				cancel()
				return runner.Result{Status: run.StatusFailed, Error: "stalled"}
			}
			if result, ok := w.reconcileActiveRun(ctx, item, workspace); ok {
				cancel()
				return result
			}
			if err := w.client.RenewWorkItemLease(ctx, entity.RenewWorkItemLeaseInput{
				WorkItemID:     item.ID,
				ClaimToken:     item.ClaimToken,
				OrchestratorID: w.orchestratorID,
				LeaseSeconds:   w.leaseSeconds,
			}); err != nil {
				log.Printf("renew work item lease issue_id=%d work_item_id=%d: %v", item.IssueID, item.ID, err)
			}
		}
	}
}

func (w *Worker) isStalled(lastEventAt *time.Time, mu *sync.Mutex) bool {
	if w.stallTimeout <= 0 {
		return false
	}
	mu.Lock()
	elapsed := time.Since(*lastEventAt)
	mu.Unlock()
	return elapsed > w.stallTimeout
}

func (w *Worker) reconcileActiveRun(ctx context.Context, item entity.WorkItem, workspace workspace.Workspace) (runner.Result, bool) {
	issue, err := w.client.Issue(ctx, item.IssueID)
	if err != nil {
		log.Printf("refresh issue state issue_id=%d: %v", item.IssueID, err)
		return runner.Result{}, false
	}
	if isTerminalStatus(issue.Status) {
		w.cleanupWorkspace(ctx, workspace.WorkspaceKey, fmt.Sprintf("issue-%d", item.IssueID))
		return runner.Result{Status: run.StatusCancelled, Error: "cancelled by terminal issue state: " + string(issue.Status)}, true
	}
	if !isActiveStatus(issue.Status) {
		return runner.Result{Status: run.StatusCancelled, Error: "cancelled by non-active issue state: " + string(issue.Status)}, true
	}
	return runner.Result{}, false
}

func (w *Worker) retryDelay(attempt int) time.Duration {
	delay := 10 * time.Second
	for i := 1; i < attempt; i++ {
		delay *= 2
	}
	if w.maxRetryBackoff > 0 && delay > w.maxRetryBackoff {
		return w.maxRetryBackoff
	}
	return delay
}

func (w *Worker) canDispatch(issueID int64) bool {
	if len(w.running) >= w.maxConcurrent {
		return false
	}
	if _, ok := w.claimed[issueID]; ok {
		return false
	}
	w.claimed[issueID] = struct{}{}
	return true
}

func (w *Worker) markRunning(run run.Run, item entity.WorkItem) {
	w.running[item.IssueID] = runningRun{run: run, workItem: item, startedAt: time.Now().UTC()}
}

func (w *Worker) release(issueID int64) {
	delete(w.running, issueID)
	delete(w.claimed, issueID)
}

func (w *Worker) flushOutbox(ctx context.Context) error {
	events, err := w.store.UnsentOutboxEvents(ctx, 20)
	if err != nil {
		return err
	}
	for _, event := range events {
		if err := w.client.SendRunEvent(ctx, event); err != nil {
			return err
		}
		if err := w.store.MarkOutboxEventSent(ctx, event.ID); err != nil {
			return err
		}
	}
	return nil
}

func (w *Worker) cleanupTerminalWorkspaces(ctx context.Context) {
	issues, err := w.client.Issues(ctx)
	if err != nil {
		log.Printf("startup terminal workspace cleanup: %v", err)
		return
	}
	for _, issue := range issues {
		if !isTerminalStatus(issue.Status) {
			continue
		}
		key := fmt.Sprintf("issue-%d", issue.ID)
		w.cleanupWorkspace(ctx, key, key)
	}
}

func (w *Worker) cleanupWorkspace(ctx context.Context, workspaceKey string, identifier string) {
	if err := w.workspaceManager.RemoveForIssue(identifier); err != nil {
		log.Printf("cleanup workspace key=%s: %v", workspaceKey, err)
		_ = w.store.MarkWorkspaceCleanup(ctx, workspaceKey, "failed", err.Error())
		return
	}
	if err := w.store.MarkWorkspaceCleanup(ctx, workspaceKey, "removed", ""); err != nil {
		log.Printf("record workspace cleanup key=%s: %v", workspaceKey, err)
	}
}

func (w *Worker) logOperatorVisibility(ctx context.Context) {
	outboxCount, err := w.store.UnsentOutboxCount(ctx)
	if err != nil {
		log.Printf("operator visibility outbox count: %v", err)
	}
	setupFailureCount, err := w.store.WorkspaceSetupFailureCount(ctx)
	if err != nil {
		log.Printf("operator visibility workspace setup failure count: %v", err)
	}
	if outboxCount > 0 || setupFailureCount > 0 {
		log.Printf("operator_visibility running=%d claimed=%d unsent_outbox=%d workspace_setup_failures=%d", len(w.running), len(w.claimed), outboxCount, setupFailureCount)
	}
}

func isActiveStatus(status entity.Status) bool {
	return status == entity.StatusReady || status == entity.StatusInProgress || status == entity.StatusReview
}

func isTerminalStatus(status entity.Status) bool {
	return status == entity.StatusDone || status == entity.StatusFailed
}
