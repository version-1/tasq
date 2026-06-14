package config

const (
	DefaultIssueTrackerPort = 37651
	DefaultOrchestratorPort = 37652
	DefaultWebPort          = 37653
)

func DefaultWorkflowTemplate() string {
	return `---
polling:
  interval_ms: 30000
workspace:
  root: .worktrees
agent:
  max_concurrent_agents: 1
  max_turns: 20
  continuation_turns_enabled: false
  max_retry_attempts: 3
  max_retry_backoff_ms: 300000
codex:
  command: codex app-server
  read_timeout_ms: 5000
  turn_timeout_ms: 3600000
  stall_timeout_ms: 300000
---
`
}
