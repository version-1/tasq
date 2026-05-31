package runner

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
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

	prompt := renderPrompt(Task{
		PromptTemplate: "{{ issue.id }} {{ issue.title }} {{ issue.description }} {{ attempt }}",
		Attempt:        2,
		Issue: entity.Issue{
			ID:          7,
			Title:       "Title",
			Description: "Description",
		},
	})
	if prompt != "7 Title Description 2" {
		t.Fatalf("prompt = %q", prompt)
	}
}
