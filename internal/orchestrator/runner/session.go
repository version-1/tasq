package runner

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/version-1/tasq/internal/orchestrator/runner/transport"
	"github.com/version-1/tasq/internal/orchestrator/runner/transport/stdio"
)

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
