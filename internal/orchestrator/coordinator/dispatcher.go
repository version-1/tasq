package coordinator

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/orchestrator/run"
	"github.com/version-1/tasq/internal/orchestrator/runner"
	"github.com/version-1/tasq/internal/orchestrator/workflow"
	"github.com/version-1/tasq/internal/orchestrator/workspace"
)

const dispatcherShutdownTimeout = 30 * time.Second
const maxRunnerEventLogPayloadLength = 2000
const ignoredRunnerEventTypeAgentMessageDelta = "item/agentMessage/delta"
const maxContinuationChangeRequests = 20

type IssueReader interface {
	Issue(ctx context.Context, id int64) (entity.Issue, error)
}

type IssueUpdater interface {
	UpdateIssue(ctx context.Context, id int64, input entity.UpdateIssueInput) (entity.Issue, error)
	CreateComment(ctx context.Context, issueID int64, input entity.CreateCommentInput) (entity.Comment, error)
	ChangeRequestsByIssueID(ctx context.Context, issueID int64, status entity.ChangeRequestStatus, limit int) ([]entity.ChangeRequest, error)
	UpdateChangeRequest(ctx context.Context, id int64, input entity.UpdateChangeRequestInput) (entity.ChangeRequest, error)
}

type IssueTracker interface {
	IssueReader
	IssueUpdater
}

type DispatchStore interface {
	UpdateRunStatus(ctx context.Context, runID string, status run.Status, errText string) (run.Run, error)
	UpdateRunThreadID(ctx context.Context, runID string, threadID string) (run.Run, error)
	RecordRunnerEvent(ctx context.Context, runID string, eventType string, message string, payloadJSON string) error
	LatestResumeThreadIDByIssueID(ctx context.Context, issueID int64) (string, error)
	RunsByIssueID(ctx context.Context, issueID int64) ([]run.Run, error)
}

type WorkflowResolver interface {
	Resolve(ctx context.Context, projectID int64) (workflow.Definition, error)
}

type Dispatcher struct {
	tracker           IssueTracker
	store             DispatchStore
	runner            runner.Runner
	workflowResolver  WorkflowResolver
	maxConcurrentRuns int

	mu      sync.Mutex
	claimed map[string]struct{}
	runCtx  context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

type DispatcherConfig struct {
	Tracker           IssueTracker
	Store             DispatchStore
	Runner            runner.Runner
	WorkflowResolver  WorkflowResolver
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
	if config.WorkflowResolver == nil {
		return nil, errors.New("workflow resolver is required")
	}
	if config.MaxConcurrentRuns <= 0 {
		return nil, errors.New("max concurrent runs must be positive")
	}
	runCtx, cancel := context.WithCancel(context.Background())
	return &Dispatcher{
		tracker:           config.Tracker,
		store:             config.Store,
		runner:            config.Runner,
		workflowResolver:  config.WorkflowResolver,
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
		definition, err := d.workflowResolver.Resolve(ctx, issue.ProjectID)
		if err != nil {
			d.release(queuedRun.RunID)
			return fmt.Errorf("resolve workflow for project %d run %s: %w", issue.ProjectID, queuedRun.RunID, err)
		}
		task, err := d.taskForRun(ctx, queuedRun, issue, definition)
		if err != nil {
			d.release(queuedRun.RunID)
			return err
		}
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
			d.failRun(storedRun, message)
		}
	}()

	if _, err := d.store.UpdateRunStatus(context.Background(), storedRun.RunID, run.StatusRunning, ""); err != nil {
		log.Printf("orchestrator dispatch update running failed run=%s: %v", storedRun.RunID, err)
		return
	}
	d.recordEvent(storedRun.RunID, "running", "runner started", "")

	task.OnEvent = func(event runner.Event) {
		if !shouldKeepRunnerEvent(event) {
			return
		}
		logRunnerEvent(storedRun.RunID, event)
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
	if result.Status == run.StatusFailed {
		d.blockRunnableIssue(context.Background(), storedRun, result.Error)
	}
}

func (d *Dispatcher) taskForRun(ctx context.Context, storedRun run.Run, issue entity.Issue, definition workflow.Definition) (runner.Task, error) {
	resumeThreadID, err := d.store.LatestResumeThreadIDByIssueID(ctx, storedRun.IssueID)
	if errors.Is(err, sql.ErrNoRows) {
		resumeThreadID = ""
	} else if err != nil {
		return runner.Task{}, fmt.Errorf("read resume thread id for issue %d run %s: %w", storedRun.IssueID, storedRun.RunID, err)
	}
	changeRequests, err := d.claimContinuationChangeRequests(ctx, storedRun)
	if err != nil {
		return runner.Task{}, err
	}
	return runner.Task{
		Issue:          issue,
		Attempt:        storedRun.Attempt,
		RunID:          storedRun.RunID,
		Workspace:      workspace.Workspace{Path: storedRun.Workspace, WorkspaceKey: issueIdentifier(storedRun.IssueID)},
		PromptTemplate: definition.PromptTemplate,
		TaskWorkPrompt: definition.Config.Tasq.TaskWorkPrompt,
		ResumeThreadID: resumeThreadID,
		ChangeRequests: changeRequests,
		MaxTurns:       definition.Config.MaxTurns,
		ContinueTurns:  definition.Config.ContinuationTurns,
		Command:        definition.Config.CodexCommand,
		ReadTimeout:    definition.Config.CodexReadTimeout,
		TurnTimeout:    definition.Config.CodexTurnTimeout,
	}, nil
}

func (d *Dispatcher) claimContinuationChangeRequests(ctx context.Context, storedRun run.Run) ([]entity.ChangeRequest, error) {
	runs, err := d.store.RunsByIssueID(ctx, storedRun.IssueID)
	if err != nil {
		return nil, fmt.Errorf("list runs for issue %d run %s: %w", storedRun.IssueID, storedRun.RunID, err)
	}
	if !hasPriorRun(runs, storedRun.RunID) {
		return nil, nil
	}
	openRequests, err := d.tracker.ChangeRequestsByIssueID(ctx, storedRun.IssueID, entity.ChangeRequestOpen, maxContinuationChangeRequests)
	if err != nil {
		return nil, fmt.Errorf("list open change requests for issue %d run %s: %w", storedRun.IssueID, storedRun.RunID, err)
	}
	claimed := make([]entity.ChangeRequest, 0, len(openRequests))
	for _, request := range openRequests {
		status := entity.ChangeRequestInProgress
		updated, err := d.tracker.UpdateChangeRequest(ctx, request.ID, entity.UpdateChangeRequestInput{Status: &status})
		if err != nil {
			return nil, fmt.Errorf("claim change request %d for run %s: %w", request.ID, storedRun.RunID, err)
		}
		claimed = append(claimed, updated)
	}
	return claimed, nil
}

func hasPriorRun(runs []run.Run, currentRunID string) bool {
	for _, storedRun := range runs {
		if storedRun.RunID != currentRunID {
			return true
		}
	}
	return false
}

func (d *Dispatcher) failRun(storedRun run.Run, message string) {
	if _, err := d.store.UpdateRunStatus(context.Background(), storedRun.RunID, run.StatusFailed, message); err != nil {
		log.Printf("orchestrator dispatch update panic failed run=%s: %v", storedRun.RunID, err)
		return
	}
	d.recordEvent(storedRun.RunID, string(run.StatusFailed), message, "")
	d.blockRunnableIssue(context.Background(), storedRun, message)
}

func (d *Dispatcher) blockRunnableIssue(ctx context.Context, storedRun run.Run, errText string) {
	issue, err := d.tracker.Issue(ctx, storedRun.IssueID)
	if err != nil {
		log.Printf("orchestrator dispatch fetch issue failed run=%s issue=%d: %v", storedRun.RunID, storedRun.IssueID, err)
		return
	}
	if !shouldBlockFailedRunIssue(issue.Status) {
		return
	}
	blocked := entity.StatusBlocked
	if _, err := d.tracker.UpdateIssue(ctx, storedRun.IssueID, entity.UpdateIssueInput{Status: &blocked}); err != nil {
		log.Printf("orchestrator dispatch block issue failed run=%s issue=%d: %v", storedRun.RunID, storedRun.IssueID, err)
		return
	}
	body := failureCommentBody(storedRun.RunID, errText)
	if _, err := d.tracker.CreateComment(ctx, storedRun.IssueID, entity.CreateCommentInput{
		Author: "tasq-orchestrator",
		Type:   entity.CommentBlocker,
		Body:   body,
	}); err != nil {
		log.Printf("orchestrator dispatch comment issue failure failed run=%s issue=%d: %v", storedRun.RunID, storedRun.IssueID, err)
	}
}

func shouldBlockFailedRunIssue(status entity.Status) bool {
	return status == entity.StatusReady || status == entity.StatusInProgress
}

func failureCommentBody(runID string, errText string) string {
	if errText == "" {
		errText = "runner failed without an error message"
	}
	const maxErrorLength = 4000
	if len(errText) > maxErrorLength {
		errText = errText[:maxErrorLength] + "... truncated"
	}
	return fmt.Sprintf("Orchestrator marked this issue blocked because runner run %s failed.\n\nError: %s", runID, errText)
}

func (d *Dispatcher) recordEvent(runID string, eventType string, message string, payloadJSON string) {
	if err := d.store.RecordRunnerEvent(context.Background(), runID, eventType, message, payloadJSON); err != nil {
		log.Printf("orchestrator dispatch record event failed run=%s event=%s: %v", runID, eventType, err)
	}
	if eventType != "session_started" {
		return
	}
	threadID := strings.TrimPrefix(message, "thread_id=")
	if threadID == "" || threadID == message {
		return
	}
	if _, err := d.store.UpdateRunThreadID(context.Background(), runID, threadID); err != nil {
		log.Printf("orchestrator dispatch update thread id failed run=%s: %v", runID, err)
	}
}

func logRunnerEvent(runID string, event runner.Event) {
	if !shouldKeepRunnerEvent(event) {
		return
	}
	log.Print(formatRunnerEventLog(runID, event))
}

func shouldKeepRunnerEvent(event runner.Event) bool {
	return event.EventType != ignoredRunnerEventTypeAgentMessageDelta
}

func formatRunnerEventLog(runID string, event runner.Event) string {
	if event.PayloadJSON == "" {
		return fmt.Sprintf("orchestrator runner event run=%s event=%s message=%q", runID, event.EventType, event.Message)
	}
	return fmt.Sprintf(
		"orchestrator runner event run=%s event=%s message=%q payload=%s",
		runID,
		event.EventType,
		event.Message,
		truncateLogPayload(event.PayloadJSON),
	)
}

func truncateLogPayload(payload string) string {
	if len(payload) <= maxRunnerEventLogPayloadLength {
		return payload
	}
	return payload[:maxRunnerEventLogPayloadLength] + "... truncated"
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
