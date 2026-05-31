package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/orchestrator/run"
	"github.com/version-1/tasq/internal/orchestrator/runner"
	"github.com/version-1/tasq/internal/orchestrator/runstore"
	"github.com/version-1/tasq/internal/orchestrator/tracker"
	"github.com/version-1/tasq/internal/orchestrator/workflow"
	"github.com/version-1/tasq/internal/orchestrator/workspace"
)

type Worker struct {
	store            *runstore.Store
	client           *tracker.Client
	orchestratorID   string
	pollInterval     time.Duration
	leaseSeconds     int
	maxConcurrent    int
	maxTurns         int
	promptTemplate   string
	codexCommand     string
	codexReadTimeout time.Duration
	codexTurnTimeout time.Duration
	runner           runner.Runner
	workspaceManager *workspace.Manager
	running          map[int64]runningRun
	claimed          map[int64]struct{}
}

type runningRun struct {
	run       run.Run
	workItem  entity.WorkItem
	startedAt time.Time
}

func NewWithConfig(store *runstore.Store, client *tracker.Client, orchestratorID string, leaseSeconds int, definition workflow.Definition, agentRunner runner.Runner) (*Worker, error) {
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
	workspaceManager, err := workspace.NewManager(config.WorkspaceRoot)
	if err != nil {
		return nil, err
	}
	return &Worker{
		store:            store,
		client:           client,
		orchestratorID:   orchestratorID,
		pollInterval:     config.PollInterval,
		leaseSeconds:     leaseSeconds,
		maxConcurrent:    config.MaxConcurrentRuns,
		maxTurns:         config.MaxTurns,
		promptTemplate:   definition.PromptTemplate,
		codexCommand:     config.CodexCommand,
		codexReadTimeout: config.CodexReadTimeout,
		codexTurnTimeout: config.CodexTurnTimeout,
		runner:           agentRunner,
		workspaceManager: workspaceManager,
		running:          map[int64]runningRun{},
		claimed:          map[int64]struct{}{},
	}, nil
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	w.tick(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.tick(ctx)
		}
	}
}

func (w *Worker) tick(ctx context.Context) {
	if err := w.flushOutbox(ctx); err != nil {
		log.Printf("flush outbox: %v", err)
	}
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
		return
	}
	if err := w.store.UpsertWorkspaceMetadata(ctx, runstore.WorkspaceMetadataInput{
		WorkspaceKey: workspace.WorkspaceKey,
		IssueID:      item.IssueID,
		Path:         workspace.Path,
		CreatedNow:   workspace.CreatedNow,
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
	result := w.runWithLeaseRenewal(ctx, createdRun.RunID, *item, workspace)
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
}

func (w *Worker) runWithLeaseRenewal(ctx context.Context, runID string, item entity.WorkItem, workspace workspace.Workspace) runner.Result {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan runner.Result, 1)
	go func() {
		done <- w.runner.Run(runCtx, runner.Task{
			WorkItem:       item,
			RunID:          runID,
			Workspace:      workspace,
			PromptTemplate: w.promptTemplate,
			MaxTurns:       w.maxTurns,
			Command:        w.codexCommand,
			ReadTimeout:    w.codexReadTimeout,
			TurnTimeout:    w.codexTurnTimeout,
			OnEvent: func(event runner.Event) {
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
