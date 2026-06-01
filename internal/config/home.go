package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	EnvHome = "TQ_HOME"
)

func Home() (string, error) {
	if value := strings.TrimSpace(os.Getenv(EnvHome)); value != "" {
		return filepath.Abs(value)
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home: %w", err)
	}
	return filepath.Join(userHome, ".tasq"), nil
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
	} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return "", fmt.Errorf("create %s: %w", path, err)
		}
	}
	return home, nil
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

func ConfigPath(home string) string {
	return filepath.Join(ConfigDir(home), "config.yaml")
}

func StatePath(home string) string {
	return filepath.Join(SystemDir(home), "state.json")
}

func StateLockPath(home string) string {
	return filepath.Join(SystemDir(home), "state.json.lock")
}

func IssueTrackerDBPath(home string) string {
	return filepath.Join(DataDir(home), "issues.sqlite")
}

func OrchestratorDBPath(home string) string {
	return filepath.Join(DataDir(home), "orchestrator.sqlite")
}
