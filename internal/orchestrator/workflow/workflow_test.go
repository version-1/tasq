package workflow

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadParsesFrontMatter(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "WORKFLOW.md")
	if err := os.WriteFile(path, []byte(`---
polling:
  interval_ms: 1200
workspace:
  root: .workspaces
  source: .
agent:
  max_concurrent_agents: 2
  max_turns: 3
  continuation_turns_enabled: true
  max_retry_attempts: 4
  max_retry_backoff_ms: 4000
codex:
  command: codex app-server --debug
  read_timeout_ms: 6000
  turn_timeout_ms: 7000
  stall_timeout_ms: 5000
server:
  port: 8081
hooks:
  after_create: touch after-create
  before_run: |
    echo before
    echo run
  after_run: echo after
  before_remove: echo remove
  timeout_ms: 9000
---
# Prompt

Work on {{ issue.title }}.
`), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	workflow, err := Load(path)
	if err != nil {
		t.Fatalf("load workflow: %v", err)
	}

	if workflow.Config.PollInterval != 1200*time.Millisecond {
		t.Fatalf("poll interval = %s", workflow.Config.PollInterval)
	}
	if workflow.Config.WorkspaceRoot != filepath.Join(dir, ".workspaces") {
		t.Fatalf("workspace root = %q", workflow.Config.WorkspaceRoot)
	}
	if workflow.Config.WorkspaceSource != dir {
		t.Fatalf("workspace source = %q", workflow.Config.WorkspaceSource)
	}
	if workflow.Config.MaxConcurrentRuns != 2 {
		t.Fatalf("max concurrent = %d", workflow.Config.MaxConcurrentRuns)
	}
	if workflow.Config.MaxTurns != 3 {
		t.Fatalf("max turns = %d", workflow.Config.MaxTurns)
	}
	if !workflow.Config.ContinuationTurns {
		t.Fatal("continuation turns were not enabled")
	}
	if workflow.Config.MaxRetryAttempts != 4 {
		t.Fatalf("max retry attempts = %d", workflow.Config.MaxRetryAttempts)
	}
	if workflow.Config.MaxRetryBackoff != 4*time.Second {
		t.Fatalf("max retry backoff = %s", workflow.Config.MaxRetryBackoff)
	}
	if workflow.Config.StallTimeout != 5*time.Second {
		t.Fatalf("stall timeout = %s", workflow.Config.StallTimeout)
	}
	if workflow.Config.CodexReadTimeout != 6*time.Second {
		t.Fatalf("codex read timeout = %s", workflow.Config.CodexReadTimeout)
	}
	if workflow.Config.CodexTurnTimeout != 7*time.Second {
		t.Fatalf("codex turn timeout = %s", workflow.Config.CodexTurnTimeout)
	}
	if workflow.Config.CodexCommand != "codex app-server --debug" {
		t.Fatalf("codex command = %q", workflow.Config.CodexCommand)
	}
	if workflow.Config.ServerPort != 8081 {
		t.Fatalf("server port = %d", workflow.Config.ServerPort)
	}
	if workflow.Config.HookAfterCreate != "touch after-create" {
		t.Fatalf("after_create hook = %q", workflow.Config.HookAfterCreate)
	}
	if workflow.Config.HookBeforeRun != "echo before\necho run\n" {
		t.Fatalf("before_run hook = %q", workflow.Config.HookBeforeRun)
	}
	if workflow.Config.HookAfterRun != "echo after" {
		t.Fatalf("after_run hook = %q", workflow.Config.HookAfterRun)
	}
	if workflow.Config.HookBeforeRemove != "echo remove" {
		t.Fatalf("before_remove hook = %q", workflow.Config.HookBeforeRemove)
	}
	if workflow.Config.HookTimeout != 9*time.Second {
		t.Fatalf("hook timeout = %s", workflow.Config.HookTimeout)
	}
	if workflow.PromptTemplate != "# Prompt\n\nWork on {{ issue.title }}." {
		t.Fatalf("prompt = %q", workflow.PromptTemplate)
	}
}

func TestLoadParsesYAMLCommentsAndIgnoresUnknownKeys(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "WORKFLOW.md")
	if err := os.WriteFile(path, []byte(`---
polling:
  interval_ms: 1200 # comment
unknown:
  future_key: true
agent:
  unknown_key: ignored
---
Prompt
`), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	workflow, err := Load(path)
	if err != nil {
		t.Fatalf("load workflow: %v", err)
	}

	if workflow.Config.PollInterval != 1200*time.Millisecond {
		t.Fatalf("poll interval = %s", workflow.Config.PollInterval)
	}
	if workflow.PromptTemplate != "Prompt" {
		t.Fatalf("prompt = %q", workflow.PromptTemplate)
	}
}

func TestLoadDefaultsHookTimeout(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "WORKFLOW.md")
	if err := os.WriteFile(path, []byte(`---
hooks:
  after_create: echo created
---
Prompt
`), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	workflow, err := Load(path)
	if err != nil {
		t.Fatalf("load workflow: %v", err)
	}

	if workflow.Config.HookTimeout != time.Minute {
		t.Fatalf("hook timeout = %s", workflow.Config.HookTimeout)
	}
}

func TestLoadRejectsInvalidHookTimeout(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "WORKFLOW.md")
	if err := os.WriteFile(path, []byte(`---
hooks:
  timeout_ms: 0
---
Prompt
`), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	if _, err := Load(path); err == nil {
		t.Fatal("expected invalid hook timeout error")
	}
}

func TestLoadKeepsIndentedDocumentMarkerInMultilineHook(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "WORKFLOW.md")
	if err := os.WriteFile(path, []byte(`---
hooks:
  before_run: |
    echo before
    ---
    echo after
---
Prompt
`), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	workflow, err := Load(path)
	if err != nil {
		t.Fatalf("load workflow: %v", err)
	}
	if workflow.Config.HookBeforeRun != "echo before\n---\necho after" {
		t.Fatalf("before_run hook = %q", workflow.Config.HookBeforeRun)
	}
	if workflow.PromptTemplate != "Prompt" {
		t.Fatalf("prompt = %q", workflow.PromptTemplate)
	}
}
