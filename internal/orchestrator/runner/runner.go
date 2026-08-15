package runner

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/orchestrator/run"
	"github.com/version-1/tasq/internal/orchestrator/workspace"
)

type Runner interface {
	Run(ctx context.Context, task Task) Result
}

type Task struct {
	Issue          entity.Issue
	Attempt        int
	RunID          string
	Workspace      workspace.Workspace
	PromptTemplate string
	TaskWorkPrompt *bool
	TasqCommand    string
	ResumeThreadID string
	ChangeRequests []entity.ChangeRequest
	MaxTurns       int
	ContinueTurns  bool
	Command        string
	ReadTimeout    time.Duration
	TurnTimeout    time.Duration
	OnEvent        func(Event)
}

type Result struct {
	Status run.Status
	Error  string
}

type Event struct {
	EventType   string
	Message     string
	PayloadJSON string
}

type SimulatedRunner struct {
	Duration time.Duration
}

func (r SimulatedRunner) Run(ctx context.Context, task Task) Result {
	duration := r.Duration
	if duration <= 0 {
		duration = 10 * time.Millisecond
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return Result{Status: run.StatusCancelled, Error: ctx.Err().Error()}
	case <-timer.C:
		return Result{Status: run.StatusSucceeded}
	}
}

type CodexRunner struct{}

func (r CodexRunner) Run(ctx context.Context, task Task) Result {
	if task.Command == "" {
		return Result{Status: run.StatusFailed, Error: "codex command is required"}
	}
	if task.ReadTimeout <= 0 {
		task.ReadTimeout = 5 * time.Second
	}
	if task.TurnTimeout <= 0 {
		task.TurnTimeout = time.Hour
	}
	if task.Workspace.Path == "" {
		return Result{Status: run.StatusFailed, Error: "workspace path is required"}
	}
	prompt := continuationPrompt(task)
	if task.ResumeThreadID == "" {
		var err error
		prompt, err = renderPrompt(task)
		if err != nil {
			return Result{Status: run.StatusFailed, Error: err.Error()}
		}
	}
	session, err := startSession(ctx, task)
	if err != nil {
		emit(task, "startup_failed", err.Error(), "")
		return Result{Status: run.StatusFailed, Error: err.Error()}
	}
	defer session.close()
	emit(task, "process_started", session.identifier(), "")

	if err := session.request(ctx, task.ReadTimeout, "initialize", map[string]any{
		"clientInfo": map[string]any{
			"name":    "tasq-orchestrator",
			"title":   "Tasq Orchestrator",
			"version": "0.1.0",
		},
		"capabilities": map[string]any{
			"experimentalApi": true,
		},
	}, nil); err != nil {
		return Result{Status: run.StatusFailed, Error: err.Error()}
	}
	if err := session.notify("initialized", map[string]any{}); err != nil {
		return Result{Status: run.StatusFailed, Error: err.Error()}
	}
	threadID, err := session.createOrResumeThread(ctx, task)
	if err != nil {
		return Result{Status: run.StatusFailed, Error: err.Error()}
	}
	emit(task, "session_started", "thread_id="+threadID, "")

	maxTurns := 1
	if task.ContinueTurns && task.MaxTurns > 1 {
		maxTurns = task.MaxTurns
	}
	for turnNumber := 1; turnNumber <= maxTurns; turnNumber++ {
		turnPrompt := prompt
		if turnNumber > 1 {
			turnPrompt = continuationPrompt(task)
		}
		turnID, err := session.startTurn(ctx, task, threadID, turnPrompt)
		if err != nil {
			return Result{Status: run.StatusFailed, Error: err.Error()}
		}
		emit(task, "turn_started", fmt.Sprintf("turn_id=%s turn_number=%d", turnID, turnNumber), "")
		if err := session.waitTurn(ctx, task.TurnTimeout, task, threadID, turnID); err != nil {
			return Result{Status: run.StatusFailed, Error: err.Error()}
		}
	}
	return Result{Status: run.StatusSucceeded}
}

var templateVariablePattern = regexp.MustCompile(`{{\s*([^{}]+?)\s*}}`)
var templateNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*$`)
