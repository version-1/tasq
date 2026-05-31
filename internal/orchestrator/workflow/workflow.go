package workflow

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Definition struct {
	Path           string
	Config         Config
	PromptTemplate string
}

type Config struct {
	PollInterval      time.Duration
	WorkspaceRoot     string
	WorkspaceSource   string
	MaxConcurrentRuns int
	MaxTurns          int
	ContinuationTurns bool
	MaxRetryAttempts  int
	MaxRetryBackoff   time.Duration
	StallTimeout      time.Duration
	CodexCommand      string
	CodexReadTimeout  time.Duration
	CodexTurnTimeout  time.Duration
	ServerPort        int
	HookAfterCreate   string
	HookBeforeRun     string
	HookAfterRun      string
	HookBeforeRemove  string
	HookTimeout       time.Duration
}

func Load(path string) (Definition, error) {
	if path == "" {
		path = "WORKFLOW.md"
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return Definition{}, fmt.Errorf("missing_workflow_file: %w", err)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return Definition{}, fmt.Errorf("resolve workflow path: %w", err)
	}
	config, body, err := parse(raw, filepath.Dir(absPath))
	if err != nil {
		return Definition{}, err
	}
	return Definition{
		Path:           absPath,
		Config:         config,
		PromptTemplate: strings.TrimSpace(body),
	}, nil
}

func DefaultConfig(workflowDir string) Config {
	return Config{
		PollInterval:      30 * time.Second,
		WorkspaceRoot:     filepath.Join(os.TempDir(), "symphony_workspaces"),
		WorkspaceSource:   workflowDir,
		MaxConcurrentRuns: 10,
		MaxTurns:          20,
		ContinuationTurns: false,
		MaxRetryAttempts:  3,
		MaxRetryBackoff:   5 * time.Minute,
		StallTimeout:      5 * time.Minute,
		CodexCommand:      "codex app-server",
		CodexReadTimeout:  5 * time.Second,
		CodexTurnTimeout:  time.Hour,
		ServerPort:        -1,
		HookTimeout:       60 * time.Second,
	}
}

func parse(raw []byte, workflowDir string) (Config, string, error) {
	config := DefaultConfig(workflowDir)
	text := string(raw)
	if !strings.HasPrefix(text, "---\n") && strings.TrimSpace(text) != "---" {
		return config, text, nil
	}

	rest := strings.TrimPrefix(text, "---\n")
	frontMatter, body, ok := splitFrontMatter(rest)
	if !ok {
		return Config{}, "", errors.New("workflow_parse_error: unterminated front matter")
	}
	if err := applyFrontMatter(&config, []byte(frontMatter), workflowDir); err != nil {
		return Config{}, "", err
	}
	return config, body, nil
}

func splitFrontMatter(rest string) (string, string, bool) {
	if strings.HasSuffix(rest, "\n---") {
		return strings.TrimSuffix(rest, "\n---"), "", true
	}
	frontMatter, body, ok := strings.Cut(rest, "\n---\n")
	return frontMatter, body, ok
}

type frontMatterConfig struct {
	Polling struct {
		IntervalMs configScalar `yaml:"interval_ms"`
	} `yaml:"polling"`
	Workspace struct {
		Root   configScalar `yaml:"root"`
		Source configScalar `yaml:"source"`
	} `yaml:"workspace"`
	Agent struct {
		MaxConcurrentAgents      configScalar `yaml:"max_concurrent_agents"`
		MaxTurns                 configScalar `yaml:"max_turns"`
		ContinuationTurnsEnabled configScalar `yaml:"continuation_turns_enabled"`
		MaxRetryAttempts         configScalar `yaml:"max_retry_attempts"`
		MaxRetryBackoffMs        configScalar `yaml:"max_retry_backoff_ms"`
	} `yaml:"agent"`
	Codex struct {
		Command        configScalar `yaml:"command"`
		StallTimeoutMs configScalar `yaml:"stall_timeout_ms"`
		ReadTimeoutMs  configScalar `yaml:"read_timeout_ms"`
		TurnTimeoutMs  configScalar `yaml:"turn_timeout_ms"`
	} `yaml:"codex"`
	Server struct {
		Port configScalar `yaml:"port"`
	} `yaml:"server"`
	Hooks struct {
		AfterCreate  configScalar `yaml:"after_create"`
		BeforeRun    configScalar `yaml:"before_run"`
		AfterRun     configScalar `yaml:"after_run"`
		BeforeRemove configScalar `yaml:"before_remove"`
		TimeoutMs    configScalar `yaml:"timeout_ms"`
	} `yaml:"hooks"`
}

type configScalar struct {
	Value string
	Set   bool
}

func (s *configScalar) UnmarshalYAML(node *yaml.Node) error {
	s.Set = true
	s.Value = node.Value
	return nil
}

func applyFrontMatter(config *Config, raw []byte, workflowDir string) error {
	var frontMatter frontMatterConfig
	if err := yaml.Unmarshal(raw, &frontMatter); err != nil {
		return fmt.Errorf("workflow_parse_error: %w", err)
	}
	if err := applyWorkflowYAML(config, frontMatter); err != nil {
		return err
	}
	return normalizeConfig(config, workflowDir)
}

func applyWorkflowYAML(config *Config, frontMatter frontMatterConfig) error {
	if frontMatter.Polling.IntervalMs.Set {
		parsed, err := parseMillis(frontMatter.Polling.IntervalMs.Value)
		if err != nil {
			return fmt.Errorf("workflow_parse_error: polling.interval_ms: %w", err)
		}
		config.PollInterval = parsed
	}
	if frontMatter.Workspace.Root.Set {
		config.WorkspaceRoot = frontMatter.Workspace.Root.Value
	}
	if frontMatter.Workspace.Source.Set {
		config.WorkspaceSource = frontMatter.Workspace.Source.Value
	}
	if frontMatter.Agent.MaxConcurrentAgents.Set {
		parsed, err := parsePositiveInt(frontMatter.Agent.MaxConcurrentAgents.Value)
		if err != nil {
			return fmt.Errorf("workflow_parse_error: agent.max_concurrent_agents: %w", err)
		}
		config.MaxConcurrentRuns = parsed
	}
	if frontMatter.Agent.MaxTurns.Set {
		parsed, err := parsePositiveInt(frontMatter.Agent.MaxTurns.Value)
		if err != nil {
			return fmt.Errorf("workflow_parse_error: agent.max_turns: %w", err)
		}
		config.MaxTurns = parsed
	}
	if frontMatter.Agent.ContinuationTurnsEnabled.Set {
		parsed, err := parseBool(frontMatter.Agent.ContinuationTurnsEnabled.Value)
		if err != nil {
			return fmt.Errorf("workflow_parse_error: agent.continuation_turns_enabled: %w", err)
		}
		config.ContinuationTurns = parsed
	}
	if frontMatter.Agent.MaxRetryAttempts.Set {
		parsed, err := parsePositiveInt(frontMatter.Agent.MaxRetryAttempts.Value)
		if err != nil {
			return fmt.Errorf("workflow_parse_error: agent.max_retry_attempts: %w", err)
		}
		config.MaxRetryAttempts = parsed
	}
	if frontMatter.Agent.MaxRetryBackoffMs.Set {
		parsed, err := parseMillis(frontMatter.Agent.MaxRetryBackoffMs.Value)
		if err != nil {
			return fmt.Errorf("workflow_parse_error: agent.max_retry_backoff_ms: %w", err)
		}
		config.MaxRetryBackoff = parsed
	}
	if frontMatter.Codex.Command.Set {
		config.CodexCommand = frontMatter.Codex.Command.Value
	}
	if frontMatter.Codex.StallTimeoutMs.Set {
		parsed, err := parseMillis(frontMatter.Codex.StallTimeoutMs.Value)
		if err != nil {
			return fmt.Errorf("workflow_parse_error: codex.stall_timeout_ms: %w", err)
		}
		config.StallTimeout = parsed
	}
	if frontMatter.Codex.ReadTimeoutMs.Set {
		parsed, err := parseMillis(frontMatter.Codex.ReadTimeoutMs.Value)
		if err != nil {
			return fmt.Errorf("workflow_parse_error: codex.read_timeout_ms: %w", err)
		}
		config.CodexReadTimeout = parsed
	}
	if frontMatter.Codex.TurnTimeoutMs.Set {
		parsed, err := parseMillis(frontMatter.Codex.TurnTimeoutMs.Value)
		if err != nil {
			return fmt.Errorf("workflow_parse_error: codex.turn_timeout_ms: %w", err)
		}
		config.CodexTurnTimeout = parsed
	}
	if frontMatter.Server.Port.Set {
		parsed, err := parseNonNegativeInt(frontMatter.Server.Port.Value)
		if err != nil {
			return fmt.Errorf("workflow_parse_error: server.port: %w", err)
		}
		config.ServerPort = parsed
	}
	if frontMatter.Hooks.AfterCreate.Set {
		config.HookAfterCreate = frontMatter.Hooks.AfterCreate.Value
	}
	if frontMatter.Hooks.BeforeRun.Set {
		config.HookBeforeRun = frontMatter.Hooks.BeforeRun.Value
	}
	if frontMatter.Hooks.AfterRun.Set {
		config.HookAfterRun = frontMatter.Hooks.AfterRun.Value
	}
	if frontMatter.Hooks.BeforeRemove.Set {
		config.HookBeforeRemove = frontMatter.Hooks.BeforeRemove.Value
	}
	if frontMatter.Hooks.TimeoutMs.Set {
		parsed, err := parseMillis(frontMatter.Hooks.TimeoutMs.Value)
		if err != nil {
			return fmt.Errorf("workflow_parse_error: hooks.timeout_ms: %w", err)
		}
		config.HookTimeout = parsed
	}
	return nil
}

func normalizeConfig(config *Config, workflowDir string) error {
	if config.PollInterval <= 0 {
		return errors.New("workflow_parse_error: polling.interval_ms must be positive")
	}
	if config.MaxConcurrentRuns <= 0 {
		return errors.New("workflow_parse_error: agent.max_concurrent_agents must be positive")
	}
	if config.MaxTurns <= 0 {
		return errors.New("workflow_parse_error: agent.max_turns must be positive")
	}
	if config.MaxRetryAttempts <= 0 {
		return errors.New("workflow_parse_error: agent.max_retry_attempts must be positive")
	}
	if config.MaxRetryBackoff <= 0 {
		return errors.New("workflow_parse_error: agent.max_retry_backoff_ms must be positive")
	}
	if config.CodexCommand == "" {
		return errors.New("workflow_parse_error: codex.command must be present")
	}
	if config.CodexReadTimeout <= 0 {
		return errors.New("workflow_parse_error: codex.read_timeout_ms must be positive")
	}
	if config.CodexTurnTimeout <= 0 {
		return errors.New("workflow_parse_error: codex.turn_timeout_ms must be positive")
	}
	root, err := resolveConfigPath(config.WorkspaceRoot, workflowDir)
	if err != nil {
		return err
	}
	config.WorkspaceRoot = root
	source, err := resolveConfigPath(config.WorkspaceSource, workflowDir)
	if err != nil {
		return err
	}
	config.WorkspaceSource = source
	return nil
}

func resolveConfigPath(value string, workflowDir string) (string, error) {
	if value == "" {
		return "", errors.New("workflow_parse_error: workspace.root must be present")
	}
	if strings.HasPrefix(value, "$") {
		value = os.Getenv(strings.TrimPrefix(value, "$"))
		if value == "" {
			return "", errors.New("workflow_parse_error: workspace.root env value is empty")
		}
	}
	if strings.HasPrefix(value, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("workflow_parse_error: resolve home: %w", err)
		}
		value = filepath.Join(home, strings.TrimPrefix(value, "~"))
	}
	if !filepath.IsAbs(value) {
		value = filepath.Join(workflowDir, value)
	}
	abs, err := filepath.Abs(value)
	if err != nil {
		return "", fmt.Errorf("workflow_parse_error: resolve path: %w", err)
	}
	return filepath.Clean(abs), nil
}

func parseMillis(value string) (time.Duration, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if parsed <= 0 {
		return 0, errors.New("must be positive")
	}
	return time.Duration(parsed) * time.Millisecond, nil
}

func parsePositiveInt(value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if parsed <= 0 {
		return 0, errors.New("must be positive")
	}
	return parsed, nil
}

func parseNonNegativeInt(value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if parsed < 0 {
		return 0, errors.New("must be non-negative")
	}
	return parsed, nil
}

func parseBool(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true", "yes", "on":
		return true, nil
	case "false", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid bool %q", value)
	}
}
