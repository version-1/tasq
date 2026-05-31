package workflow

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
	}
}

func parse(raw []byte, workflowDir string) (Config, string, error) {
	config := DefaultConfig(workflowDir)
	text := string(raw)
	if !strings.HasPrefix(text, "---\n") && strings.TrimSpace(text) != "---" {
		return config, text, nil
	}

	scanner := bufio.NewScanner(strings.NewReader(text))
	if !scanner.Scan() {
		return Config{}, "", errors.New("workflow_parse_error: empty workflow")
	}
	var frontMatter []string
	var body []string
	inFrontMatter := true
	for scanner.Scan() {
		line := scanner.Text()
		if inFrontMatter {
			if strings.TrimSpace(line) == "---" {
				inFrontMatter = false
				continue
			}
			frontMatter = append(frontMatter, line)
			continue
		}
		body = append(body, line)
	}
	if err := scanner.Err(); err != nil {
		return Config{}, "", fmt.Errorf("workflow_parse_error: %w", err)
	}
	if inFrontMatter {
		return Config{}, "", errors.New("workflow_parse_error: unterminated front matter")
	}
	if err := applyFrontMatter(&config, frontMatter, workflowDir); err != nil {
		return Config{}, "", err
	}
	return config, strings.Join(body, "\n"), nil
}

func applyFrontMatter(config *Config, lines []string, workflowDir string) error {
	section := ""
	for _, raw := range lines {
		withoutComment := strings.SplitN(raw, "#", 2)[0]
		if strings.TrimSpace(withoutComment) == "" {
			continue
		}
		indent := len(withoutComment) - len(strings.TrimLeft(withoutComment, " "))
		line := strings.TrimSpace(withoutComment)
		if indent == 0 {
			key, value, ok := splitConfigLine(line)
			if !ok {
				return fmt.Errorf("workflow_parse_error: invalid front matter line %q", raw)
			}
			if value != "" {
				return fmt.Errorf("workflow_front_matter_not_a_map: %s must be an object", key)
			}
			section = key
			continue
		}
		key, value, ok := splitConfigLine(line)
		if !ok {
			return fmt.Errorf("workflow_parse_error: invalid front matter line %q", raw)
		}
		if err := applyValue(config, section, key, value); err != nil {
			return err
		}
	}
	return normalizeConfig(config, workflowDir)
}

func splitConfigLine(line string) (string, string, bool) {
	key, value, ok := strings.Cut(line, ":")
	if !ok {
		return "", "", false
	}
	return strings.TrimSpace(key), trimScalar(value), true
}

func trimScalar(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, `"'`)
	return value
}

func applyValue(config *Config, section string, key string, value string) error {
	switch section {
	case "polling":
		if key == "interval_ms" {
			parsed, err := parseMillis(value)
			if err != nil {
				return fmt.Errorf("workflow_parse_error: polling.interval_ms: %w", err)
			}
			config.PollInterval = parsed
		}
	case "workspace":
		switch key {
		case "root":
			config.WorkspaceRoot = value
		case "source":
			config.WorkspaceSource = value
		}
	case "agent":
		switch key {
		case "max_concurrent_agents":
			parsed, err := parsePositiveInt(value)
			if err != nil {
				return fmt.Errorf("workflow_parse_error: agent.max_concurrent_agents: %w", err)
			}
			config.MaxConcurrentRuns = parsed
		case "max_turns":
			parsed, err := parsePositiveInt(value)
			if err != nil {
				return fmt.Errorf("workflow_parse_error: agent.max_turns: %w", err)
			}
			config.MaxTurns = parsed
		case "continuation_turns_enabled":
			parsed, err := parseBool(value)
			if err != nil {
				return fmt.Errorf("workflow_parse_error: agent.continuation_turns_enabled: %w", err)
			}
			config.ContinuationTurns = parsed
		case "max_retry_attempts":
			parsed, err := parsePositiveInt(value)
			if err != nil {
				return fmt.Errorf("workflow_parse_error: agent.max_retry_attempts: %w", err)
			}
			config.MaxRetryAttempts = parsed
		case "max_retry_backoff_ms":
			parsed, err := parseMillis(value)
			if err != nil {
				return fmt.Errorf("workflow_parse_error: agent.max_retry_backoff_ms: %w", err)
			}
			config.MaxRetryBackoff = parsed
		}
	case "codex":
		switch key {
		case "command":
			config.CodexCommand = value
		case "stall_timeout_ms":
			parsed, err := parseMillis(value)
			if err != nil {
				return fmt.Errorf("workflow_parse_error: codex.stall_timeout_ms: %w", err)
			}
			config.StallTimeout = parsed
		case "read_timeout_ms":
			parsed, err := parseMillis(value)
			if err != nil {
				return fmt.Errorf("workflow_parse_error: codex.read_timeout_ms: %w", err)
			}
			config.CodexReadTimeout = parsed
		case "turn_timeout_ms":
			parsed, err := parseMillis(value)
			if err != nil {
				return fmt.Errorf("workflow_parse_error: codex.turn_timeout_ms: %w", err)
			}
			config.CodexTurnTimeout = parsed
		}
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
