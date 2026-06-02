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

func TestRenderPromptSubstitutesIssueFields(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 3, 3, 4, 5, 0, time.UTC)
	prompt, err := renderPrompt(Task{
		PromptTemplate: "{{ issue.id }} {{ issue.title }} {{ issue.description }} {{ issue.status }} {{ issue.priority }} {{ issue.assignee }} {{ issue.created_at }} {{ issue.updated_at }} {{ attempt }}",
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

func TestRenderPromptUsesZeroAttemptForFirstAttempt(t *testing.T) {
	t.Parallel()

	prompt, err := renderPrompt(Task{
		PromptTemplate: "{{ attempt }}",
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
