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
}

type DispatchStore interface {
	UpdateRunStatus(ctx context.Context, runID string, status run.Status, errText string) (run.Run, error)
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
		go d.startRun(queuedRun, task)
		availableSlots--
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

func (d *Dispatcher) startRun(storedRun run.Run, task runner.Task) {
	defer d.wg.Done()
	defer d.release(storedRun.RunID)
	defer func() {
		if value := recover(); value != nil {
			message := fmt.Sprintf("runner panic: %v", value)
			d.failRun(storedRun.RunID, message)
		}
	}()

	if _, err := d.store.UpdateRunStatus(context.Background(), storedRun.RunID, run.StatusRunning, ""); err != nil {
		log.Printf("orchestrator dispatch update running failed run=%s: %v", storedRun.RunID, err)
		return
	}
	d.recordEvent(storedRun.RunID, "running", "runner started", "")

	task.OnEvent = func(event runner.Event) {
		d.recordEvent(storedRun.RunID, event.EventType, event.Message, event.PayloadJSON)
	}
	result := d.runner.Run(d.runCtx, task)
	if result.Status == "" {
		result.Status = run.StatusFailed
		if result.Error == "" {
			result.Error = "runner returned empty status"
		}
	}
	if _, err := d.store.UpdateRunStatus(context.Background(), storedRun.RunID, result.Status, result.Error); err != nil {
		log.Printf("orchestrator dispatch update terminal failed run=%s status=%s: %v", storedRun.RunID, result.Status, err)
		return
	}
	d.recordEvent(storedRun.RunID, string(result.Status), result.Error, "")
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

func (d *Dispatcher) failRun(runID string, message string) {
	if _, err := d.store.UpdateRunStatus(context.Background(), runID, run.StatusFailed, message); err != nil {
		log.Printf("orchestrator dispatch update panic failed run=%s: %v", runID, err)
		return
	}
	d.recordEvent(runID, string(run.StatusFailed), message, "")
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

func (d *Dispatcher) claimedCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.claimed)
}
