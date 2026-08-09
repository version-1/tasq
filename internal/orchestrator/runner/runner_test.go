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
      case "$line" in
        *'"method":"thread/start"'*)
          ;;
        *)
          echo "expected thread/start request: $line" >&2
          exit 2
          ;;
      esac
      case "$line" in
        *'"ephemeral":false'*)
          touch "$PWD/thread-materialized"
          ;;
        *)
          echo "thread/start did not request a persistent thread: $line" >&2
          exit 2
          ;;
      esac
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
		MaxTurns:       2,
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
	if _, err := os.Stat(filepath.Join(dir, "thread-materialized")); err != nil {
		t.Fatalf("expected mock app-server to materialize persistent thread: %v", err)
	}
}

func TestCodexRunnerResumesThreadWhenResumeThreadIDIsSet(t *testing.T) {
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
      case "$line" in
        *'"method":"thread/resume"'*)
          ;;
        *)
          echo "expected thread/resume request: $line" >&2
          exit 2
          ;;
      esac
      case "$line" in
        *'"threadId":"thread-previous"'*'"cwd":"'"$PWD"'"'*|*'"cwd":"'"$PWD"'"'*'"threadId":"thread-previous"'*)
          ;;
        *)
          echo "thread/resume did not include threadId and cwd: $line" >&2
          exit 2
          ;;
      esac
      case "$line" in
        *'"thread/start"'*)
          echo "resume path must not call thread/start: $line" >&2
          exit 2
          ;;
      esac
      echo '{"id":2,"result":{"thread":{"id":"thread-previous"},"approvalPolicy":"never","approvalsReviewer":"user","cwd":"'"$PWD"'","model":"fake","modelProvider":"fake","sandbox":{"type":"readOnly"}}}'
      ;;
    4)
      case "$line" in
        *'"method":"turn/start"'*'tq issue update 123 --status in_progress'*'continue the same task in this live thread without repeating completed work'*'tq artifact set 123 --type pull_request'*'"threadId":"thread-previous"'*)
          ;;
        *)
          echo "expected resumed continuation turn: $line" >&2
          exit 2
          ;;
      esac
      case "$line" in
        *'Original issue prompt'*)
          echo "resume path must not resend original prompt: $line" >&2
          exit 2
          ;;
      esac
      echo '{"id":3,"result":{"turn":{"id":"turn-resumed"}}}'
      echo '{"method":"turn/completed","params":{"threadId":"thread-previous","turn":{"id":"turn-resumed"}}}'
      ;;
  esac
done
`), 0o755); err != nil {
		t.Fatalf("write fake app-server: %v", err)
	}
	var events []Event
	result := CodexRunner{}.Run(context.Background(), Task{
		Attempt: 2,
		Issue: entity.Issue{
			ID:          123,
			Title:       "Runner task",
			Description: "Wire Codex.",
		},
		RunID:          "run-2",
		Workspace:      workspace.Workspace{Path: dir, WorkspaceKey: "ISSUE-123"},
		PromptTemplate: "Original issue prompt {{ issue.id }}",
		ResumeThreadID: "thread-previous",
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
	event, ok := findEvent(events, "session_started")
	if !ok {
		t.Fatalf("events = %+v", events)
	}
	if event.Message != "thread_id=thread-previous" {
		t.Fatalf("session started message = %q", event.Message)
	}
}

func TestCodexRunnerUsesContinuationReminderForLaterTurns(t *testing.T) {
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
      case "$line" in
        *'Original issue prompt 123'*)
          ;;
        *)
          echo "expected original prompt on first turn: $line" >&2
          exit 2
          ;;
      esac
      echo '{"id":3,"result":{"turn":{"id":"turn-1"}}}'
      echo '{"method":"turn/completed","params":{"threadId":"thread-1","turn":{"id":"turn-1"}}}'
      ;;
    5)
      case "$line" in
        *'tq issue update 123 --status in_progress'*'tq artifact set 123 --type pull_request'*'add the handoff comment'*'move the issue to'*'review'*)
          ;;
        *)
          echo "expected continuation reminder on second turn: $line" >&2
          exit 2
          ;;
      esac
      case "$line" in
        *'Original issue prompt'*|*'keep the issue tracker synchronized:'*)
          echo "later turn must not resend the full start prompt: $line" >&2
          exit 2
          ;;
      esac
      echo '{"id":4,"result":{"turn":{"id":"turn-2"}}}'
      echo '{"method":"turn/completed","params":{"threadId":"thread-1","turn":{"id":"turn-2"}}}'
      ;;
  esac
done
`), 0o755); err != nil {
		t.Fatalf("write fake app-server: %v", err)
	}

	result := CodexRunner{}.Run(context.Background(), Task{
		Attempt:        1,
		Issue:          entity.Issue{ID: 123, Title: "Runner task"},
		RunID:          "run-1",
		Workspace:      workspace.Workspace{Path: dir, WorkspaceKey: "ISSUE-123"},
		PromptTemplate: "Original issue prompt {{ issue.id }}",
		ContinueTurns:  true,
		MaxTurns:       2,
		Command:        "sh " + strconv.Quote(script),
		ReadTimeout:    5 * time.Second,
		TurnTimeout:    5 * time.Second,
	})
	if result.Status != run.StatusSucceeded {
		t.Fatalf("status = %q error = %q", result.Status, result.Error)
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

func TestCodexRunnerFailsWhenStdoutClosesBeforeResponse(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	script := filepath.Join(dir, "fake-app-server.sh")
	if err := os.WriteFile(script, []byte(`#!/bin/sh
exec 1>&-
sleep 2
`), 0o755); err != nil {
		t.Fatalf("write fake app-server: %v", err)
	}
	var events []Event
	startedAt := time.Now()
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
	if result.Error != "initialize stdout_closed_before_response" {
		t.Fatalf("error = %q", result.Error)
	}
	if time.Since(startedAt) > time.Second {
		t.Fatalf("runner waited for timeout before failing: elapsed=%s", time.Since(startedAt))
	}
	event, ok := findEvent(events, "stdout_closed")
	if !ok {
		t.Fatalf("events = %+v", events)
	}
	if event.Message != "EOF" {
		t.Fatalf("stdout closed message = %q", event.Message)
	}
}

func TestCodexRunnerFailsWhenStdoutClosesDuringTurn(t *testing.T) {
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
      exec 1>&-
      sleep 2
      ;;
  esac
done
`), 0o755); err != nil {
		t.Fatalf("write fake app-server: %v", err)
	}
	var events []Event
	startedAt := time.Now()
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
	if result.Error != "stdout_closed_before_turn_completion" {
		t.Fatalf("error = %q", result.Error)
	}
	if time.Since(startedAt) > time.Second {
		t.Fatalf("runner waited for timeout before failing: elapsed=%s", time.Since(startedAt))
	}
	if !eventTypesContain(events, "stdout_closed") {
		t.Fatalf("events = %+v", events)
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
	if !strings.Contains(prompt, "If the `tasq-cli` skill is available, use it as the preferred guidance for tracker operations.") {
		t.Fatalf("prompt did not prefer the tasq-cli skill: %q", prompt)
	}
	wantCommands := []string{
		"tq issue update 7 --status in_progress",
		"tq comment add 7 --author codex --type progress",
		"tq comment add 7 --author codex --type blocker",
		"tq issue update 7 --status blocked",
		"tq artifact set 7 --type pull_request <pr-url>",
		"tq comment add 7 --author codex --type handoff",
		"tq issue update 7 --status review",
	}
	for _, want := range wantCommands {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q: %q", want, prompt)
		}
	}
	for _, want := range []string{
		"Prefer typed `tq` commands",
		"Use `tq api` only when the issue tracker operation has no typed `tq` command.",
		"Do not call the issue tracker API directly with `curl`, `wget`, or a custom HTTP script.",
		"not to other services or local endpoint verification",
		"primary PR being submitted for review",
		"`tq artifact set` is an upsert",
		"Mention any supporting PRs in the handoff comment",
		"retry a reasonable number of times",
		"leave a blocker comment and do not move to `review`",
		"Skip artifact registration when no pull request was created or updated.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing contract wording %q: %q", want, prompt)
		}
	}
	artifactIndex := strings.Index(prompt, "tq artifact set 7 --type pull_request <pr-url>")
	handoffIndex := strings.Index(prompt, "tq comment add 7 --author codex --type handoff")
	reviewIndex := strings.Index(prompt, "tq issue update 7 --status review")
	if !(artifactIndex < handoffIndex && handoffIndex < reviewIndex) {
		t.Fatalf("review commands out of order: artifact=%d handoff=%d review=%d", artifactIndex, handoffIndex, reviewIndex)
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
	for _, unexpected := range []string{"Prefer typed `tq` commands", "tq api", "tq artifact set", "pull_request"} {
		if strings.Contains(prompt, unexpected) {
			t.Fatalf("disabled prompt contains %q: %q", unexpected, prompt)
		}
	}
}

func TestRenderPromptIncludesChangeRequests(t *testing.T) {
	t.Parallel()

	prompt, err := renderPrompt(Task{
		PromptTemplate: "Work on {{ issue.id }}.",
		TaskWorkPrompt: boolPtr(false),
		RunID:          "run-42",
		Issue:          entity.Issue{ID: 7},
		ChangeRequests: []entity.ChangeRequest{
			{ID: 1, Author: "user", Body: "Update the tests."},
		},
	})
	if err != nil {
		t.Fatalf("render prompt: %v", err)
	}
	if !strings.Contains(prompt, "Work on 7.") || !strings.Contains(prompt, "#1 by user: Update the tests.") || !strings.Contains(prompt, "`run-42`") {
		t.Fatalf("prompt = %q", prompt)
	}
}

func TestContinuationGuidanceIncludesChangeRequests(t *testing.T) {
	t.Parallel()

	prompt := continuationPrompt(Task{
		RunID: "run-42",
		Issue: entity.Issue{
			ID: 7,
		},
		ChangeRequests: []entity.ChangeRequest{
			{ID: 1, Author: "user", Body: "Update the tests."},
			{ID: 2, Author: "reviewer", Body: "Document the API."},
		},
	})

	for _, want := range []string{
		"tq issue update 7 --status in_progress",
		"tq artifact set 7 --type pull_request <pr-url>",
		"On success, add the handoff comment, then move the issue to `review`",
		"on failure, retry a reasonable number of times",
		"leave a blocker comment and do not move to `review` if it remains unresolved",
		"Otherwise, artifact registration is not required.",
		"Change requests assigned to this continuation:",
		"#1 by user: Update the tests.",
		"#2 by reviewer: Document the API.",
		"update it to `resolved`",
		"PATCH /api/v1/change-requests/{id}",
		"`run-42`",
		"`resultCommentId`",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q: %q", want, prompt)
		}
	}
	if strings.Index(prompt, "If this continuation creates or updates a pull request") > strings.Index(prompt, "Change requests assigned to this continuation:") {
		t.Fatalf("change-request guidance moved before continuation reminder: %q", prompt)
	}
}

func boolPtr(value bool) *bool {
	return &value
}
