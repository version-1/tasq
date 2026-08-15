package config

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	EnvHome       = "TQ_HOME"
	EnvManagedRun = "TQ_MANAGED_RUN"
)

var defaultHomeProfile = ""

func DefaultHomeProfile() (string, error) {
	profile := strings.TrimSpace(defaultHomeProfile)
	if profile == "" {
		return "", nil
	}
	for _, char := range profile {
		if (char < 'a' || char > 'z') && (char < '0' || char > '9') && char != '-' {
			return "", fmt.Errorf("invalid default home profile %q", profile)
		}
	}
	return profile, nil
}

func TasqCommand() (string, error) {
	profile, err := DefaultHomeProfile()
	if err != nil {
		return "", err
	}
	switch profile {
	case "":
		return "tq", nil
	case "dev":
		return "tqdev", nil
	default:
		return "", fmt.Errorf("unsupported build profile %q for Tasq CLI command", profile)
	}
}

func ValidateTasqCommand() (string, error) {
	command, err := TasqCommand()
	if err != nil {
		return "", err
	}
	if _, err := exec.LookPath(command); err != nil {
		return "", fmt.Errorf("resolve Tasq CLI command %q from PATH: %w", command, err)
	}
	return command, nil
}

func Home() (string, error) {
	if value := strings.TrimSpace(os.Getenv(EnvHome)); value != "" {
		return filepath.Abs(value)
	}
	profile, err := DefaultHomeProfile()
	if err != nil {
		return "", err
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}
	dirName := ".tasq"
	if profile != "" {
		dirName += "-" + profile
	}
	return filepath.Join(userHome, dirName), nil
}

func EnsureHome() (string, error) {
	home, err := Home()
	if err != nil {
		return "", err
	}
	for _, path := range []string{
		home,
		ConfigDir(home),
		SystemDir(home),
		DataDir(home),
		LogDir(home),
	} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return "", fmt.Errorf("create %s: %w", path, err)
		}
	}
	if err := ensureWorkflowFile(WorkflowPath(home)); err != nil {
		return "", err
	}
	return home, nil
}

func ensureWorkflowFile(path string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer file.Close()
	if _, err := file.WriteString(DefaultWorkflowTemplate()); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func ConfigDir(home string) string {
	return filepath.Join(home, "config")
}

func SystemDir(home string) string {
	return filepath.Join(home, "system")
}

func DataDir(home string) string {
	return filepath.Join(SystemDir(home), "data")
}

func LogDir(home string) string {
	return filepath.Join(SystemDir(home), "log")
}

func ConfigPath(home string) string {
	return filepath.Join(ConfigDir(home), "config.yaml")
}

func StatePath(home string) string {
	return filepath.Join(SystemDir(home), "state.json")
}

func StateLockPath(home string) string {
	return filepath.Join(SystemDir(home), "state.json.lock")
}

func WorkflowPath(home string) string {
	return filepath.Join(home, "WORKFLOW.md")
}

func IssueTrackerDBPath(home string) string {
	return filepath.Join(DataDir(home), "issues.sqlite")
}

func OrchestratorDBPath(home string) string {
	return filepath.Join(DataDir(home), "orchestrator.sqlite")
}
