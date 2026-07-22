package runner

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/orchestrator/run"
	"github.com/version-1/tasq/internal/orchestrator/runner/transport"
	"github.com/version-1/tasq/internal/orchestrator/runner/transport/stdio"
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
	ResumeThreadID string
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

type session struct {
	conn     transport.Connection
	mu       sync.Mutex
	nextID   int64
	closing  atomic.Bool
	messages chan rpcMessage
}

type rpcMessage struct {
	ID     any             `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int64           `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

func startSession(ctx context.Context, task Task) (*session, error) {
	conn, err := stdio.Start(ctx, task.Command, task.Workspace.Path)
	if err != nil {
		return nil, err
	}
	s := &session{
		conn:     conn,
		messages: make(chan rpcMessage, 64),
	}
	go s.readFrames(task)
	go readStderr(conn.Stderr(), task)
	return s, nil
}

func (s *session) identifier() string {
	return s.conn.Identifier()
}

func (s *session) close() {
	s.closing.Store(true)
	_ = s.conn.Close()
}

func (s *session) readFrames(task Task) {
	defer close(s.messages)

	for {
		frame, err := s.conn.Receive(context.Background())
		if err != nil {
			if s.closing.Load() {
				return
			}
			message := "EOF"
			if !errors.Is(err, io.EOF) {
				message = err.Error()
			}
			emit(task, s.conn.FrameSource()+"_closed", message, "")
			return
		}
		var message rpcMessage
		if err := json.Unmarshal(frame, &message); err != nil {
			line := string(frame)
			emitMalformedFrame(task, s.conn.FrameSource(), line, err)
			s.messages <- rpcMessage{Method: "malformed", Params: append([]byte(nil), frame...)}
			continue
		}
		s.messages <- message
	}
}

func readStderr(stderr io.Reader, task Task) {
	scanner := bufio.NewScanner(stderr)
	scanner.Buffer(make([]byte, 0, 16*1024), 1024*1024)
	for scanner.Scan() {
		emit(task, "stderr", scanner.Text(), "")
	}
}

func emitMalformedFrame(task Task, source string, line string, err error) {
	payload, marshalErr := json.Marshal(map[string]string{
		"error": err.Error(),
	})
	payloadJSON := ""
	if marshalErr == nil {
		payloadJSON = string(payload)
	}
	emit(task, source+"_malformed", truncateText(line, 10000), payloadJSON)
}

func (s *session) request(ctx context.Context, timeout time.Duration, method string, params any, output any) error {
	id := s.nextRequestID()
	if err := s.write(map[string]any{"id": id, "method": method, "params": params}); err != nil {
		return err
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-s.conn.Done():
			return appServerExitedError("codex app-server exited", err)
		case <-timer.C:
			return fmt.Errorf("%s response_timeout", method)
		case message, ok := <-s.messages:
			if !ok {
				return fmt.Errorf("%s %s_closed_before_response", method, s.conn.FrameSource())
			}
			if message.Method != "" && message.ID != nil {
				_ = s.write(map[string]any{
					"id": message.ID,
					"error": map[string]any{
						"code":    -32601,
						"message": "server request is not supported by tasq orchestrator",
					},
				})
				continue
			}
			if message.Method != "" {
				continue
			}
			if fmt.Sprint(message.ID) != strconv.FormatInt(id, 10) {
				continue
			}
			if message.Error != nil {
				return fmt.Errorf("%s response_error: %s", method, message.Error.Message)
			}
			if output != nil {
				if err := json.Unmarshal(message.Result, output); err != nil {
					return fmt.Errorf("decode %s response: %w", method, err)
				}
			}
			return nil
		}
	}
}

func (s *session) notify(method string, params any) error {
	return s.write(map[string]any{"method": method, "params": params})
}

func (s *session) createOrResumeThread(ctx context.Context, task Task) (string, error) {
	method := "thread/start"
	params := map[string]any{
		"cwd":          task.Workspace.Path,
		"ephemeral":    false,
		"serviceName":  "tasq-orchestrator",
		"threadSource": "user",
	}
	if task.ResumeThreadID != "" {
		method = "thread/resume"
		params = map[string]any{
			"cwd":      task.Workspace.Path,
			"threadId": task.ResumeThreadID,
		}
	}

	var response struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	if err := s.request(ctx, task.ReadTimeout, method, params, &response); err != nil {
		return "", err
	}
	if response.Thread.ID == "" {
		return "", fmt.Errorf("%s returned empty thread id", method)
	}
	return response.Thread.ID, nil
}

func (s *session) startTurn(ctx context.Context, task Task, threadID string, prompt string) (string, error) {
	var turnStart struct {
		Turn struct {
			ID string `json:"id"`
		} `json:"turn"`
	}
	if err := s.request(ctx, task.ReadTimeout, "turn/start", map[string]any{
		"threadId": threadID,
		"cwd":      task.Workspace.Path,
		"input": []map[string]any{
			{"type": "text", "text": prompt},
		},
	}, &turnStart); err != nil {
		return "", err
	}
	if turnStart.Turn.ID == "" {
		return "", errors.New("turn/start returned empty turn id")
	}
	return turnStart.Turn.ID, nil
}

func (s *session) waitTurn(ctx context.Context, timeout time.Duration, task Task, threadID string, turnID string) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-s.conn.Done():
			return appServerExitedError("codex app-server exited before turn completion", err)
		case <-timer.C:
			return errors.New("turn_timeout")
		case message, ok := <-s.messages:
			if !ok {
				return fmt.Errorf("%s_closed_before_turn_completion", s.conn.FrameSource())
			}
			if message.Method == "" {
				continue
			}
			payload := string(message.Params)
			emit(task, message.Method, "", payload)
			if message.ID != nil {
				if isApprovalRequest(message.Method) {
					if err := s.write(map[string]any{
						"id":     message.ID,
						"result": map[string]any{"decision": "cancel"},
					}); err != nil {
						return err
					}
					return approvalRequestDeniedError(message.Method, payload)
				}
				_ = s.write(map[string]any{
					"id": message.ID,
					"error": map[string]any{
						"code":    -32601,
						"message": "server request is not supported by tasq orchestrator",
					},
				})
				continue
			}
			if message.Method == "turn/completed" && notificationMatches(message.Params, threadID, turnID) {
				emit(task, "turn_completed", "turn_id="+turnID, payload)
				return nil
			}
			if message.Method == "error" && notificationMatches(message.Params, threadID, turnID) {
				return fmt.Errorf("turn_failed: %s", payload)
			}
		}
	}
}

func approvalRequestDeniedError(method string, payload string) error {
	return fmt.Errorf(
		"approval_required: tasq denied app-server approval request by policy\n\nmethod: %s\npayload: %s",
		method,
		truncateText(payload, 4000),
	)
}

func isApprovalRequest(method string) bool {
	return method == "item/commandExecution/requestApproval" || method == "item/fileChange/requestApproval"
}

func truncateText(value string, maxLength int) string {
	if maxLength <= 0 || len(value) <= maxLength {
		return value
	}
	return value[:maxLength] + "... truncated"
}

func (s *session) nextRequestID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextID++
	return s.nextID
}

func (s *session) write(message any) error {
	raw, err := json.Marshal(message)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conn.Send(context.Background(), raw)
}

func appServerExitedError(message string, err error) error {
	if err == nil {
		return errors.New(message)
	}
	return fmt.Errorf("%s: %w", message, err)
}

func notificationMatches(raw json.RawMessage, threadID string, turnID string) bool {
	var payload struct {
		ThreadID string `json:"threadId"`
		TurnID   string `json:"turnId"`
		Turn     struct {
			ID string `json:"id"`
		} `json:"turn"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return false
	}
	if payload.ThreadID != "" && payload.ThreadID != threadID {
		return false
	}
	if payload.TurnID != "" {
		return payload.TurnID == turnID
	}
	return payload.Turn.ID == turnID
}

var templateVariablePattern = regexp.MustCompile(`{{\s*([^{}]+?)\s*}}`)
var templateNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*(\.[A-Za-z_][A-Za-z0-9_]*)*$`)

const defaultTaskWorkPrompt = "Use `tq` to keep the issue tracker synchronized:\n" +
	"\n" +
	"If the `tasq-cli` skill is available, use it as the preferred guidance for tracker operations.\n" +
	"\n" +
	"- When work starts, move the issue to `in_progress` and leave a progress comment.\n" +
	"- Add progress comments at meaningful milestones during the work.\n" +
	"- If work is blocked, leave a blocker comment that explains why, then move the issue to `blocked`.\n" +
	"- When the pull request is ready for review, leave a handoff comment with the PR and verification summary, then move the issue to `review`.\n" +
	"- Always pass `--author codex` when posting comments.\n" +
	"- Run only the commands for the current lifecycle stage; the examples below are not a single script.\n" +
	"\n" +
	"```sh\n" +
	"# Start\n" +
	"tq issue update {{ issue.id }} --status in_progress\n" +
	"tq comment add {{ issue.id }} --author codex --type progress --body \"Started work.\"\n" +
	"\n" +
	"# Meaningful progress milestone\n" +
	"tq comment add {{ issue.id }} --author codex --type progress --body \"Implemented the change; running verification.\"\n" +
	"\n" +
	"# Blocked (use instead of the review handoff)\n" +
	"tq comment add {{ issue.id }} --author codex --type blocker --body \"Blocked: explain the blocker and what is needed.\"\n" +
	"tq issue update {{ issue.id }} --status blocked\n" +
	"\n" +
	"# Ready for review\n" +
	"tq comment add {{ issue.id }} --author codex --type handoff --body \"PR: <url>; verification: <summary>.\"\n" +
	"tq issue update {{ issue.id }} --status review\n" +
	"```\n" +
	"\n" +
	"Run the installed `tq` binary from `PATH`. Do not use `go run ./cmd/tq` for\n" +
	"tracker synchronization."

const continuationGuidance = "First run `tq issue update %d --status in_progress` to keep the issue tracker synchronized. Then continue the same task in this live thread. Do not repeat completed work. Stop when the workflow is ready for handoff."

func continuationPrompt(task Task) string {
	return fmt.Sprintf(continuationGuidance, task.Issue.ID)
}

func renderPrompt(task Task) (string, error) {
	prompt := task.PromptTemplate
	if prompt == "" {
		prompt = "Work on issue {{ issue.id }}: {{ issue.title }}\n\n{{ issue.description }}"
	}
	if shouldInjectTaskWorkPrompt(task.TaskWorkPrompt) {
		prompt = defaultTaskWorkPrompt + "\n\n" + prompt
	}
	if strings.Count(prompt, "{{") != strings.Count(prompt, "}}") {
		return "", errors.New("template_parse_error: unbalanced template delimiters")
	}
	var renderErr error
	vars := templateVariables(task)
	rendered := templateVariablePattern.ReplaceAllStringFunc(prompt, func(match string) string {
		if renderErr != nil {
			return match
		}
		parts := templateVariablePattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			renderErr = fmt.Errorf("template_parse_error: malformed template expression %q", match)
			return match
		}
		expression := strings.TrimSpace(parts[1])
		if strings.Contains(expression, "|") {
			name, _, _ := strings.Cut(expression, "|")
			if templateNamePattern.MatchString(strings.TrimSpace(name)) {
				renderErr = fmt.Errorf("template_render_error: unknown filter in %q", expression)
			}
			return match
		}
		if !templateNamePattern.MatchString(expression) {
			return match
		}
		value, ok := vars[expression]
		if !ok {
			renderErr = fmt.Errorf("template_render_error: unknown variable %q", expression)
			return match
		}
		return value
	})
	if renderErr != nil {
		return "", renderErr
	}
	return rendered, nil
}

func shouldInjectTaskWorkPrompt(taskWorkPrompt *bool) bool {
	return taskWorkPrompt == nil || *taskWorkPrompt
}

func templateVariables(task Task) map[string]string {
	attempt := 0
	if task.Attempt > 1 {
		attempt = task.Attempt
	}
	return map[string]string{
		"issue.id":          strconv.FormatInt(task.Issue.ID, 10),
		"issue.title":       task.Issue.Title,
		"issue.description": task.Issue.Description,
		"issue.status":      string(task.Issue.Status),
		"issue.priority":    string(task.Issue.Priority),
		"issue.assignee":    task.Issue.Assignee,
		"issue.created_at":  task.Issue.CreatedAt.Format(time.RFC3339),
		"issue.updated_at":  task.Issue.UpdatedAt.Format(time.RFC3339),
		"attempt":           strconv.Itoa(attempt),
	}
}

func emit(task Task, eventType string, message string, payloadJSON string) {
	if task.OnEvent != nil {
		task.OnEvent(Event{EventType: eventType, Message: message, PayloadJSON: payloadJSON})
	}
}
