package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadParsesFrontMatter(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "WORKFLOW.md")
	if err := os.WriteFile(path, []byte(`---
polling:
  interval_ms: 1200
tracker:
  kind: linear
  endpoint: https://linear.example/graphql
  api_key: $TEST_LINEAR_API_KEY
  project_slug: TASQ
  active_states:
    - Todo
    - In Progress
  terminal_states:
    - Done
    - Canceled
workspace:
  root: .workspaces
  source: .
agent:
  max_concurrent_agents: 2
  max_concurrent_agents_by_state:
    Todo: 1
    IN_PROGRESS: 3
    invalid: 0
    not_numeric: nope
  max_turns: 3
  continuation_turns_enabled: true
  max_retry_attempts: 4
  max_retry_backoff_ms: 4000
codex:
  command: codex app-server --debug
  approval_policy: never
  thread_sandbox: workspace-write
  turn_sandbox_policy: danger-full-access
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
	t.Setenv("TEST_LINEAR_API_KEY", "linear-token")

	workflow, err := Load(path)
	if err != nil {
		t.Fatalf("load workflow: %v", err)
	}

	if workflow.Config.PollInterval != 1200*time.Millisecond {
		t.Fatalf("poll interval = %s", workflow.Config.PollInterval)
	}
	if workflow.Config.TrackerKind != "linear" {
		t.Fatalf("tracker kind = %q", workflow.Config.TrackerKind)
	}
	if workflow.Config.TrackerEndpoint != "https://linear.example/graphql" {
		t.Fatalf("tracker endpoint = %q", workflow.Config.TrackerEndpoint)
	}
	if workflow.Config.TrackerAPIKey != "linear-token" {
		t.Fatalf("tracker api key = %q", workflow.Config.TrackerAPIKey)
	}
	if workflow.Config.TrackerProjectSlug != "TASQ" {
		t.Fatalf("tracker project slug = %q", workflow.Config.TrackerProjectSlug)
	}
	if strings.Join(workflow.Config.TrackerActiveStates, ",") != "Todo,In Progress" {
		t.Fatalf("tracker active states = %+v", workflow.Config.TrackerActiveStates)
	}
	if strings.Join(workflow.Config.TrackerTerminalStates, ",") != "Done,Canceled" {
		t.Fatalf("tracker terminal states = %+v", workflow.Config.TrackerTerminalStates)
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
	if workflow.Config.MaxConcurrentRunsByState["todo"] != 1 || workflow.Config.MaxConcurrentRunsByState["in_progress"] != 3 {
		t.Fatalf("max concurrent by state = %+v", workflow.Config.MaxConcurrentRunsByState)
	}
	if _, ok := workflow.Config.MaxConcurrentRunsByState["invalid"]; ok {
		t.Fatalf("invalid state entry was kept: %+v", workflow.Config.MaxConcurrentRunsByState)
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
	if workflow.Config.CodexApprovalPolicy != "never" {
		t.Fatalf("codex approval policy = %q", workflow.Config.CodexApprovalPolicy)
	}
	if workflow.Config.CodexThreadSandbox != "workspace-write" {
		t.Fatalf("codex thread sandbox = %q", workflow.Config.CodexThreadSandbox)
	}
	if workflow.Config.CodexTurnSandboxPolicy != "danger-full-access" {
		t.Fatalf("codex turn sandbox policy = %q", workflow.Config.CodexTurnSandboxPolicy)
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

func TestLoadDefaultsTrackerAndCodexExtensionFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "WORKFLOW.md")
	if err := os.WriteFile(path, []byte("Prompt\n"), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	workflow, err := Load(path)
	if err != nil {
		t.Fatalf("load workflow: %v", err)
	}

	if workflow.Config.TrackerKind != "" {
		t.Fatalf("tracker kind = %q", workflow.Config.TrackerKind)
	}
	if workflow.Config.TrackerEndpoint != "https://api.linear.app/graphql" {
		t.Fatalf("tracker endpoint = %q", workflow.Config.TrackerEndpoint)
	}
	if strings.Join(workflow.Config.TrackerActiveStates, ",") != "Todo,In Progress" {
		t.Fatalf("tracker active states = %+v", workflow.Config.TrackerActiveStates)
	}
	if strings.Join(workflow.Config.TrackerTerminalStates, ",") != "Closed,Cancelled,Canceled,Duplicate,Done" {
		t.Fatalf("tracker terminal states = %+v", workflow.Config.TrackerTerminalStates)
	}
	if len(workflow.Config.MaxConcurrentRunsByState) != 0 {
		t.Fatalf("max concurrent by state = %+v", workflow.Config.MaxConcurrentRunsByState)
	}
	if workflow.Config.CodexApprovalPolicy != "" || workflow.Config.CodexThreadSandbox != "" || workflow.Config.CodexTurnSandboxPolicy != "" {
		t.Fatalf("codex pass-through fields = %q/%q/%q", workflow.Config.CodexApprovalPolicy, workflow.Config.CodexThreadSandbox, workflow.Config.CodexTurnSandboxPolicy)
	}
}

func TestLoadRejectsTrackerKindWithoutProjectSlug(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "WORKFLOW.md")
	if err := os.WriteFile(path, []byte(`---
tracker:
  kind: linear
---
Prompt
`), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "workflow_parse_error: tracker.project_slug") {
		t.Fatalf("error = %v", err)
	}
}

func TestLoadRejectsNonMapFrontMatter(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "WORKFLOW.md")
	if err := os.WriteFile(path, []byte(`---
- invalid
---
Prompt
`), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	_, err := Load(path)
	if err == nil || !strings.Contains(err.Error(), "workflow_front_matter_not_a_map") {
		t.Fatalf("error = %v", err)
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
