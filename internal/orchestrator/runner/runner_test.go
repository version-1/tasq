package runner

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/orchestrator/run"
	"github.com/version-1/tasq/internal/orchestrator/workspace"
)

func TestCodexRunnerCompletesTurn(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(dir, "fake-app-server.sh")
	if err := os.WriteFile(script, []byte(`#!/bin/sh
count=0
while IFS= read -r line; do
  count=$((count + 1))
  case "$count" in
    1)
      echo '{"id":1,"result":{"codexHome":"/tmp/codex","platformFamily":"unix","platformOs":"linux","userAgent":"fake"}}'
      ;;
    3)
      echo '{"id":2,"result":{"thread":{"id":"thread-1"},"approvalPolicy":"never","approvalsReviewer":"user","cwd":"'"$PWD"'","model":"fake","modelProvider":"fake","sandbox":{"type":"readOnly"}}}'
      ;;
    4)
      echo '{"id":3,"result":{"turn":{"id":"turn-1"}}}'
      echo '{"method":"turn/completed","params":{"threadId":"thread-1","turn":{"id":"turn-1"}}}'
      ;;
  esac
done
`), 0o755); err != nil {
		t.Fatalf("write fake app-server: %v", err)
	}
	var events []Event
	result := CodexRunner{}.Run(context.Background(), Task{
		Attempt: 1,
		Issue: entity.Issue{
			ID:          123,
			Title:       "Runner task",
			Description: "Wire Codex.",
		},
		RunID:          "run-1",
		Workspace:      workspace.Workspace{Path: dir, WorkspaceKey: "ISSUE-123"},
		PromptTemplate: "Work on {{ issue.id }}: {{ issue.title }}",
		Command:        "sh " + strconv.Quote(script),
		ReadTimeout:    5 * time.Second,
		TurnTimeout:    5 * time.Second,
		OnEvent: func(event Event) {
			events = append(events, event)
		},
	})
	if result.Status != run.StatusSucceeded {
		t.Fatalf("status = %q error = %q", result.Status, result.Error)
	}
	if len(events) == 0 {
		t.Fatal("expected runner events")
	}
}

func TestCodexRunnerEmitsMalformedStdout(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(dir, "fake-app-server.sh")
	if err := os.WriteFile(script, []byte(`#!/bin/sh
count=0
while IFS= read -r line; do
  count=$((count + 1))
  case "$count" in
    1)
      echo 'shell startup noise'
      echo '{"id":1,"result":{"codexHome":"/tmp/codex","platformFamily":"unix","platformOs":"linux","userAgent":"fake"}}'
      ;;
    3)
      echo '{"id":2,"result":{"thread":{"id":"thread-1"},"approvalPolicy":"never","approvalsReviewer":"user","cwd":"'"$PWD"'","model":"fake","modelProvider":"fake","sandbox":{"type":"readOnly"}}}'
      ;;
    4)
      echo '{"id":3,"result":{"turn":{"id":"turn-1"}}}'
      echo '{"method":"turn/completed","params":{"threadId":"thread-1","turn":{"id":"turn-1"}}}'
      ;;
  esac
done
`), 0o755); err != nil {
		t.Fatalf("write fake app-server: %v", err)
	}
	var events []Event
	result := CodexRunner{}.Run(context.Background(), Task{
		Attempt: 1,
		Issue: entity.Issue{
			ID:          123,
			Title:       "Runner task",
			Description: "Wire Codex.",
		},
		RunID:          "run-1",
		Workspace:      workspace.Workspace{Path: dir, WorkspaceKey: "ISSUE-123"},
		PromptTemplate: "Work on {{ issue.id }}: {{ issue.title }}",
		Command:        "sh " + strconv.Quote(script),
		ReadTimeout:    5 * time.Second,
		TurnTimeout:    5 * time.Second,
		OnEvent: func(event Event) {
			events = append(events, event)
		},
	})
	if result.Status != run.StatusSucceeded {
		t.Fatalf("status = %q error = %q", result.Status, result.Error)
	}
	event, ok := findEvent(events, "stdout_malformed")
	if !ok {
		t.Fatalf("events = %+v", events)
	}
	if event.Message != "shell startup noise" {
		t.Fatalf("malformed stdout message = %q", event.Message)
	}
	if !strings.Contains(event.PayloadJSON, "invalid character") {
		t.Fatalf("malformed stdout payload = %q", event.PayloadJSON)
	}
}

func TestCodexRunnerFailsTurnWhenApprovalRequestIsDenied(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(dir, "fake-app-server.sh")
	if err := os.WriteFile(script, []byte(`#!/bin/sh
count=0
while IFS= read -r line; do
  count=$((count + 1))
  case "$count" in
    1)
      echo '{"id":1,"result":{"codexHome":"/tmp/codex","platformFamily":"unix","platformOs":"linux","userAgent":"fake"}}'
      ;;
    3)
      echo '{"id":2,"result":{"thread":{"id":"thread-1"},"approvalPolicy":"never","approvalsReviewer":"user","cwd":"'"$PWD"'","model":"fake","modelProvider":"fake","sandbox":{"type":"readOnly"}}}'
      ;;
    4)
      echo '{"id":3,"result":{"turn":{"id":"turn-1"}}}'
      echo '{"id":99,"method":"item/fileChange/requestApproval","params":{"threadId":"thread-1","turnId":"turn-1","itemId":"change-1","reason":"needs approval","startedAtMs":1}}'
      ;;
    5)
      case "$line" in
        *'"result":{"decision":"cancel"}'*)
          ;;
        *)
          echo "unexpected approval response: $line" >&2
          exit 2
          ;;
      esac
      echo '{"method":"turn/completed","params":{"threadId":"thread-1","turn":{"id":"turn-1"}}}'
      ;;
  esac
done
`), 0o755); err != nil {
		t.Fatalf("write fake app-server: %v", err)
	}
	var events []Event
	result := CodexRunner{}.Run(context.Background(), Task{
		Attempt: 1,
		Issue: entity.Issue{
			ID:          123,
			Title:       "Runner task",
			Description: "Wire Codex.",
		},
		RunID:          "run-1",
		Workspace:      workspace.Workspace{Path: dir, WorkspaceKey: "ISSUE-123"},
		PromptTemplate: "Work on {{ issue.id }}: {{ issue.title }}",
		Command:        "sh " + strconv.Quote(script),
		ReadTimeout:    5 * time.Second,
		TurnTimeout:    5 * time.Second,
		OnEvent: func(event Event) {
			events = append(events, event)
		},
	})
	if result.Status != run.StatusFailed {
		t.Fatalf("status = %q error = %q", result.Status, result.Error)
	}
	if !strings.Contains(result.Error, "approval_required") || !strings.Contains(result.Error, "item/fileChange/requestApproval") || !strings.Contains(result.Error, "needs approval") {
		t.Fatalf("error = %q", result.Error)
	}
	if !eventTypesContain(events, "item/fileChange/requestApproval") {
		t.Fatalf("events = %+v", events)
	}
}

func TestRenderPromptSubstitutesIssueFields(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 3, 3, 4, 5, 0, time.UTC)
	prompt, err := renderPrompt(Task{
		PromptTemplate: "{{ issue.id }} {{ issue.title }} {{ issue.description }} {{ issue.status }} {{ issue.priority }} {{ issue.assignee }} {{ issue.created_at }} {{ issue.updated_at }} {{ attempt }}",
		TaskWorkPrompt: boolPtr(false),
		Attempt:        2,
		Issue: entity.Issue{
			ID:          7,
			Title:       "Title",
			Description: "Description",
			Status:      entity.StatusReady,
			Priority:    entity.PriorityHigh,
			Assignee:    "agent",
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		},
	})
	if err != nil {
		t.Fatalf("render prompt: %v", err)
	}
	if prompt != "7 Title Description ready high agent 2026-01-02T03:04:05Z 2026-01-03T03:04:05Z 2" {
		t.Fatalf("prompt = %q", prompt)
	}
}

func eventTypesContain(events []Event, eventType string) bool {
	for _, event := range events {
		if event.EventType == eventType {
			return true
		}
	}
	return false
}

func findEvent(events []Event, eventType string) (Event, bool) {
	for _, event := range events {
		if event.EventType == eventType {
			return event, true
		}
	}
	return Event{}, false
}

func TestRenderPromptUsesZeroAttemptForFirstAttempt(t *testing.T) {
	t.Parallel()

	prompt, err := renderPrompt(Task{
		PromptTemplate: "{{ attempt }}",
		TaskWorkPrompt: boolPtr(false),
		Attempt:        1,
	})
	if err != nil {
		t.Fatalf("render prompt: %v", err)
	}
	if prompt != "0" {
		t.Fatalf("prompt = %q", prompt)
	}
}

func TestRenderPromptRejectsUnknownVariableAndFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		template string
		want     string
	}{
		{name: "unknown variable", template: "{{ issue.labels }}", want: "template_render_error"},
		{name: "unknown filter", template: "{{ issue.title | upcase }}", want: "template_render_error"},
		{name: "unbalanced", template: "{{ issue.title", want: "template_parse_error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := renderPrompt(Task{PromptTemplate: tt.template})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %s", err, tt.want)
			}
		})
	}
}

func TestRenderPromptInjectsTaskWorkPromptByDefault(t *testing.T) {
	t.Parallel()

	prompt, err := renderPrompt(Task{
		PromptTemplate: "Work on {{ issue.title }}.",
		Issue: entity.Issue{
			ID:    7,
			Title: "Runner task",
		},
	})
	if err != nil {
		t.Fatalf("render prompt: %v", err)
	}
	if !strings.HasPrefix(prompt, "Use `tq` to keep the issue tracker synchronized:") {
		t.Fatalf("prompt did not start with task work prompt: %q", prompt)
	}
	if !strings.Contains(prompt, "tq issue update 7 --status in_progress") {
		t.Fatalf("prompt did not render injected issue ID: %q", prompt)
	}
	if !strings.Contains(prompt, "\n\nWork on Runner task.") {
		t.Fatalf("prompt missing workflow template: %q", prompt)
	}
}

func TestRenderPromptSkipsTaskWorkPromptWhenDisabled(t *testing.T) {
	t.Parallel()

	prompt, err := renderPrompt(Task{
		PromptTemplate: "Work on {{ issue.id }}.",
		TaskWorkPrompt: boolPtr(false),
		Issue:          entity.Issue{ID: 7},
	})
	if err != nil {
		t.Fatalf("render prompt: %v", err)
	}
	if prompt != "Work on 7." {
		t.Fatalf("prompt = %q", prompt)
	}
}

func boolPtr(value bool) *bool {
	return &value
}
