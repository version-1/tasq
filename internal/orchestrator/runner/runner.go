package runner

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
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
	prompt, err := renderPrompt(task)
	if err != nil {
		return Result{Status: run.StatusFailed, Error: err.Error()}
	}
	session, err := startSession(ctx, task)
	if err != nil {
		emit(task, "startup_failed", err.Error(), "")
		return Result{Status: run.StatusFailed, Error: err.Error()}
	}
	defer session.close()
	emit(task, "process_started", fmt.Sprintf("pid=%d", session.pid()), "")

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
	var threadStart struct {
		Thread struct {
			ID string `json:"id"`
		} `json:"thread"`
	}
	if err := session.request(ctx, task.ReadTimeout, "thread/start", map[string]any{
		"cwd":          task.Workspace.Path,
		"ephemeral":    true,
		"serviceName":  "tasq-orchestrator",
		"threadSource": "user",
	}, &threadStart); err != nil {
		return Result{Status: run.StatusFailed, Error: err.Error()}
	}
	if threadStart.Thread.ID == "" {
		return Result{Status: run.StatusFailed, Error: "thread/start returned empty thread id"}
	}
	emit(task, "session_started", "thread_id="+threadStart.Thread.ID, "")

	maxTurns := 1
	if task.ContinueTurns && task.MaxTurns > 1 {
		maxTurns = task.MaxTurns
	}
	for turnNumber := 1; turnNumber <= maxTurns; turnNumber++ {
		turnPrompt := prompt
		if turnNumber > 1 {
			turnPrompt = "Continue the same task in this live thread. Do not repeat completed work. Stop when the workflow is ready for handoff."
		}
		turnID, err := session.startTurn(ctx, task, threadStart.Thread.ID, turnPrompt)
		if err != nil {
			return Result{Status: run.StatusFailed, Error: err.Error()}
		}
		emit(task, "turn_started", fmt.Sprintf("turn_id=%s turn_number=%d", turnID, turnNumber), "")
		if err := session.waitTurn(ctx, task.TurnTimeout, task, threadStart.Thread.ID, turnID); err != nil {
			return Result{Status: run.StatusFailed, Error: err.Error()}
		}
	}
	return Result{Status: run.StatusSucceeded}
}

type session struct {
	cmd      *exec.Cmd
	stdin    io.WriteCloser
	mu       sync.Mutex
	nextID   int64
	messages chan rpcMessage
	done     chan error
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
	cmd := exec.CommandContext(ctx, "bash", "-lc", task.Command)
	cmd.Dir = task.Workspace.Path
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("create stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start codex app-server: %w", err)
	}
	s := &session{
		cmd:      cmd,
		stdin:    stdin,
		messages: make(chan rpcMessage, 64),
		done:     make(chan error, 1),
	}
	go s.readStdout(stdout)
	go readStderr(stderr, task)
	go func() {
		s.done <- cmd.Wait()
		close(s.done)
	}()
	return s, nil
}

func (s *session) pid() int {
	if s.cmd.Process == nil {
		return 0
	}
	return s.cmd.Process.Pid
}

func (s *session) close() {
	_ = s.stdin.Close()
	if s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	<-s.done
}

func (s *session) readStdout(stdout io.Reader) {
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		var message rpcMessage
		if err := json.Unmarshal(scanner.Bytes(), &message); err != nil {
			s.messages <- rpcMessage{Method: "malformed", Params: append([]byte(nil), scanner.Bytes()...)}
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
		case err := <-s.done:
			if err == nil {
				return errors.New("codex app-server exited")
			}
			return fmt.Errorf("codex app-server exited: %w", err)
		case <-timer.C:
			return fmt.Errorf("%s response_timeout", method)
		case message := <-s.messages:
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
		case err := <-s.done:
			if err == nil {
				return errors.New("codex app-server exited before turn completion")
			}
			return fmt.Errorf("codex app-server exited before turn completion: %w", err)
		case <-timer.C:
			return errors.New("turn_timeout")
		case message := <-s.messages:
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
	if _, err := s.stdin.Write(append(raw, '\n')); err != nil {
		return fmt.Errorf("write app-server request: %w", err)
	}
	return nil
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

func renderPrompt(task Task) (string, error) {
	prompt := task.PromptTemplate
	if prompt == "" {
		prompt = "Work on issue {{ issue.id }}: {{ issue.title }}\n\n{{ issue.description }}"
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
