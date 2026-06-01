package coordinator

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/orchestrator/run"
	"github.com/version-1/tasq/internal/orchestrator/runner"
	"github.com/version-1/tasq/internal/orchestrator/workflow"
	"github.com/version-1/tasq/internal/orchestrator/workspace"
)

const dispatcherShutdownTimeout = 30 * time.Second

type IssueReader interface {
	Issue(ctx context.Context, id int64) (entity.Issue, error)
	IssueStatesByIDs(ctx context.Context, ids []int64) ([]entity.IssueState, error)
}

type DispatchStore interface {
	UpdateRunStatus(ctx context.Context, runID string, status run.Status, errText string) (run.Run, error)
	CompleteRunningRun(ctx context.Context, runID string, status run.Status, errText string) (run.Run, error)
	ScheduleRetry(ctx context.Context, runID string, errText string, retryAfter time.Time) (run.Run, error)
	LastEventTimes(ctx context.Context, runIDs []string) (map[string]time.Time, error)
	RecordRunnerEvent(ctx context.Context, runID string, eventType string, message string, payloadJSON string) error
}

type Dispatcher struct {
	tracker           IssueReader
	store             DispatchStore
	runner            runner.Runner
	workflowConfig    workflow.Config
	promptTemplate    string
	maxConcurrentRuns int

	mu      sync.Mutex
	claimed map[string]struct{}
	cancels map[string]context.CancelFunc
	runCtx  context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

type DispatcherConfig struct {
	Tracker           IssueReader
	Store             DispatchStore
	Runner            runner.Runner
	WorkflowConfig    workflow.Config
	PromptTemplate    string
	MaxConcurrentRuns int
}

func NewDispatcher(config DispatcherConfig) (*Dispatcher, error) {
	if config.Tracker == nil {
		return nil, errors.New("tracker is required")
	}
	if config.Store == nil {
		return nil, errors.New("store is required")
	}
	if config.Runner == nil {
		return nil, errors.New("runner is required")
	}
	if config.MaxConcurrentRuns <= 0 {
		return nil, errors.New("max concurrent runs must be positive")
	}
	runCtx, cancel := context.WithCancel(context.Background())
	return &Dispatcher{
		tracker:           config.Tracker,
		store:             config.Store,
		runner:            config.Runner,
		workflowConfig:    config.WorkflowConfig,
		promptTemplate:    config.PromptTemplate,
		maxConcurrentRuns: config.MaxConcurrentRuns,
		claimed:           make(map[string]struct{}),
		cancels:           make(map[string]context.CancelFunc),
		runCtx:            runCtx,
		cancel:            cancel,
	}, nil
}

func (d *Dispatcher) Dispatch(ctx context.Context, activeRuns []run.Run) error {
	if len(activeRuns) == 0 {
		return nil
	}
	runningCount := 0
	var queuedRuns []run.Run
	for _, activeRun := range activeRuns {
		switch activeRun.Status {
		case run.StatusRunning:
			runningCount++
		case run.StatusQueued:
			queuedRuns = append(queuedRuns, activeRun)
		}
	}
	if len(queuedRuns) == 0 {
		return nil
	}
	sort.SliceStable(queuedRuns, func(i, j int) bool {
		if queuedRuns[i].CreatedAt.Equal(queuedRuns[j].CreatedAt) {
			return queuedRuns[i].ID < queuedRuns[j].ID
		}
		return queuedRuns[i].CreatedAt.Before(queuedRuns[j].CreatedAt)
	})

	availableSlots := d.maxConcurrentRuns - runningCount - d.claimedCount()
	if availableSlots <= 0 {
		return nil
	}
	for _, queuedRun := range queuedRuns {
		if availableSlots <= 0 {
			break
		}
		if !d.claim(queuedRun.RunID) {
			continue
		}
		issue, err := d.tracker.Issue(ctx, queuedRun.IssueID)
		if err != nil {
			d.release(queuedRun.RunID)
			return fmt.Errorf("fetch issue %d for run %s: %w", queuedRun.IssueID, queuedRun.RunID, err)
		}
		task := d.taskForRun(queuedRun, issue)
		d.wg.Add(1)
		runCtx, cancel := context.WithCancel(d.runCtx)
		d.registerCancel(queuedRun.RunID, cancel)
		go d.startRun(runCtx, queuedRun, task)
		availableSlots--
	}
	return nil
}

func (d *Dispatcher) Reconcile(ctx context.Context, activeRuns []run.Run) error {
	if len(activeRuns) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(activeRuns))
	for _, activeRun := range activeRuns {
		ids = append(ids, activeRun.IssueID)
	}
	states, err := d.tracker.IssueStatesByIDs(ctx, ids)
	if err != nil {
		log.Printf("orchestrator reconcile skipped: %v", err)
		return nil
	}
	stateByID := make(map[int64]entity.Status, len(states))
	for _, state := range states {
		stateByID[state.ID] = state.Status
	}
	for _, activeRun := range activeRuns {
		state, ok := stateByID[activeRun.IssueID]
		if !ok || !isTerminalIssueState(state) {
			continue
		}
		reason := fmt.Sprintf("issue %d is terminal: %s", activeRun.IssueID, state)
		switch activeRun.Status {
		case run.StatusQueued:
			if _, err := d.store.UpdateRunStatus(ctx, activeRun.RunID, run.StatusCancelled, reason); err != nil {
				return fmt.Errorf("cancel queued run %s: %w", activeRun.RunID, err)
			}
			d.recordEvent(activeRun.RunID, "reconcile_cancelled", reason, "")
		case run.StatusRunning:
			if cancel, ok := d.cancelForRun(activeRun.RunID); ok {
				cancel()
			}
			if _, err := d.store.UpdateRunStatus(ctx, activeRun.RunID, run.StatusCancelled, reason); err != nil {
				return fmt.Errorf("cancel running run %s: %w", activeRun.RunID, err)
			}
			d.recordEvent(activeRun.RunID, "reconcile_cancelled", reason, "")
		}
	}
	return nil
}

func (d *Dispatcher) DetectStalls(ctx context.Context, runningRuns []run.Run) error {
	if len(runningRuns) == 0 || d.workflowConfig.StallTimeout <= 0 {
		return nil
	}
	runIDs := make([]string, 0, len(runningRuns))
	for _, runningRun := range runningRuns {
		runIDs = append(runIDs, runningRun.RunID)
	}
	lastEventTimes, err := d.store.LastEventTimes(ctx, runIDs)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	for _, runningRun := range runningRuns {
		lastEvent, ok := lastEventTimes[runningRun.RunID]
		if !ok {
			lastEvent = runningRun.CreatedAt
		}
		if now.Sub(lastEvent) <= d.workflowConfig.StallTimeout {
			continue
		}
		reason := fmt.Sprintf("stall timeout after %s", d.workflowConfig.StallTimeout)
		if cancel, ok := d.cancelForRun(runningRun.RunID); ok {
			cancel()
		}
		if _, err := d.store.UpdateRunStatus(ctx, runningRun.RunID, run.StatusFailed, reason); err != nil {
			return fmt.Errorf("fail stalled run %s: %w", runningRun.RunID, err)
		}
		d.recordEvent(runningRun.RunID, "stall_timeout", reason, "")
	}
	return nil
}

func (d *Dispatcher) Shutdown(ctx context.Context) error {
	d.cancel()
	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()
	timeout := dispatcherShutdownTimeout
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
		if timeout <= 0 {
			timeout = dispatcherShutdownTimeout
		}
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return fmt.Errorf("dispatcher shutdown timed out after %s", timeout)
	}
}

func (d *Dispatcher) startRun(runCtx context.Context, storedRun run.Run, task runner.Task) {
	defer d.wg.Done()
	defer d.release(storedRun.RunID)
	defer d.unregisterCancel(storedRun.RunID)
	defer func() {
		if value := recover(); value != nil {
			message := fmt.Sprintf("runner panic: %v", value)
			d.handleFailedRun(storedRun, message)
		}
	}()

	if _, err := d.store.UpdateRunStatus(context.Background(), storedRun.RunID, run.StatusRunning, ""); err != nil {
		log.Printf("orchestrator dispatch update running failed run=%s: %v", storedRun.RunID, err)
		return
	}
	d.release(storedRun.RunID)
	d.recordEvent(storedRun.RunID, "running", "runner started", "")

	task.OnEvent = func(event runner.Event) {
		d.recordEvent(storedRun.RunID, event.EventType, event.Message, event.PayloadJSON)
	}
	result := d.runner.Run(runCtx, task)
	if result.Status == "" {
		result.Status = run.StatusFailed
		if result.Error == "" {
			result.Error = "runner returned empty status"
		}
	}
	if result.Status == run.StatusFailed {
		d.handleFailedRun(storedRun, result.Error)
		return
	}
	completedRun, err := d.store.CompleteRunningRun(context.Background(), storedRun.RunID, result.Status, result.Error)
	if err != nil {
		log.Printf("orchestrator dispatch update terminal failed run=%s status=%s: %v", storedRun.RunID, result.Status, err)
		return
	}
	if completedRun.Status == result.Status {
		d.recordEvent(storedRun.RunID, string(result.Status), result.Error, "")
	}
}

func (d *Dispatcher) taskForRun(storedRun run.Run, issue entity.Issue) runner.Task {
	return runner.Task{
		Issue:          issue,
		Attempt:        storedRun.Attempt,
		RunID:          storedRun.RunID,
		Workspace:      workspace.Workspace{Path: storedRun.Workspace, WorkspaceKey: issueIdentifier(storedRun.IssueID)},
		PromptTemplate: d.promptTemplate,
		MaxTurns:       d.workflowConfig.MaxTurns,
		ContinueTurns:  d.workflowConfig.ContinuationTurns,
		Command:        d.workflowConfig.CodexCommand,
		ReadTimeout:    d.workflowConfig.CodexReadTimeout,
		TurnTimeout:    d.workflowConfig.CodexTurnTimeout,
	}
}

func (d *Dispatcher) handleFailedRun(storedRun run.Run, message string) {
	if storedRun.Attempt >= d.workflowConfig.MaxRetryAttempts {
		completedRun, err := d.store.CompleteRunningRun(context.Background(), storedRun.RunID, run.StatusFailed, message)
		if err != nil {
			log.Printf("orchestrator dispatch update failed run=%s: %v", storedRun.RunID, err)
			return
		}
		if completedRun.Status != run.StatusFailed || completedRun.Error != message {
			return
		}
		d.recordEvent(storedRun.RunID, "retry_exhausted", message, "")
		d.recordEvent(storedRun.RunID, string(run.StatusFailed), message, "")
		return
	}
	backoff := calculateBackoff(storedRun.Attempt, d.workflowConfig.MaxRetryBackoff)
	retryAfter := time.Now().UTC().Add(backoff)
	scheduledRun, err := d.store.ScheduleRetry(context.Background(), storedRun.RunID, message, retryAfter)
	if err != nil {
		log.Printf("orchestrator dispatch schedule retry failed run=%s: %v", storedRun.RunID, err)
		return
	}
	if scheduledRun.Status != run.StatusFailed || scheduledRun.RetryAfter == nil || scheduledRun.Error != message {
		return
	}
	d.recordEvent(storedRun.RunID, "retry_scheduled", fmt.Sprintf("retry after %s", backoff), "")
	d.recordEvent(storedRun.RunID, string(run.StatusFailed), message, "")
}

func (d *Dispatcher) recordEvent(runID string, eventType string, message string, payloadJSON string) {
	if err := d.store.RecordRunnerEvent(context.Background(), runID, eventType, message, payloadJSON); err != nil {
		log.Printf("orchestrator dispatch record event failed run=%s event=%s: %v", runID, eventType, err)
	}
}

func (d *Dispatcher) claim(runID string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.claimed[runID]; ok {
		return false
	}
	d.claimed[runID] = struct{}{}
	return true
}

func (d *Dispatcher) release(runID string) {
	d.mu.Lock()
	delete(d.claimed, runID)
	d.mu.Unlock()
}

func (d *Dispatcher) registerCancel(runID string, cancel context.CancelFunc) {
	d.mu.Lock()
	d.cancels[runID] = cancel
	d.mu.Unlock()
}

func (d *Dispatcher) unregisterCancel(runID string) {
	d.mu.Lock()
	delete(d.cancels, runID)
	d.mu.Unlock()
}

func (d *Dispatcher) cancelForRun(runID string) (context.CancelFunc, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	cancel, ok := d.cancels[runID]
	return cancel, ok
}

func (d *Dispatcher) claimedCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.claimed)
}

func calculateBackoff(attempt int, maxBackoff time.Duration) time.Duration {
	if attempt <= 0 {
		attempt = 1
	}
	backoff := 10 * time.Second
	for i := 1; i < attempt; i++ {
		backoff *= 2
		if maxBackoff > 0 && backoff >= maxBackoff {
			return maxBackoff
		}
	}
	if maxBackoff > 0 && backoff > maxBackoff {
		return maxBackoff
	}
	return backoff
}

func isTerminalIssueState(status entity.Status) bool {
	return status == entity.StatusDone || status == entity.StatusFailed
}
