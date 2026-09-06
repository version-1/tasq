package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ProcessIdentity identifies a running process independently of PID reuse.
type ProcessIdentity struct {
	StartedAt  time.Time
	Executable string
}

var readProcessIdentity = processIdentityForPID

func CurrentProcessIdentity() (ProcessIdentity, error) {
	return ProcessIdentityForPID(os.Getpid())
}

func ProcessIdentityForPID(pid int) (ProcessIdentity, error) {
	return readProcessIdentity(pid)
}

func processIdentityForPID(pid int) (ProcessIdentity, error) {
	startedAt, err := processStartTime(pid)
	if err != nil {
		return ProcessIdentity{}, err
	}
	if pid == os.Getpid() {
		executable, err := os.Executable()
		if err != nil {
			return ProcessIdentity{}, fmt.Errorf("resolve current executable: %w", err)
		}
		executable, err = canonicalExecutable(executable)
		if err != nil {
			return ProcessIdentity{}, err
		}
		return ProcessIdentity{StartedAt: startedAt, Executable: executable}, nil
	}
	return ProcessIdentity{StartedAt: startedAt}, nil
}

func (state ServiceState) MatchesProcessIdentity() (bool, error) {
	if state.PID <= 0 || state.ProcessStartedAt.IsZero() {
		return false, nil
	}
	identity, err := ProcessIdentityForPID(state.PID)
	if err != nil {
		return false, err
	}
	return state.ProcessStartedAt.Equal(identity.StartedAt), nil
}

func processStartTime(pid int) (time.Time, error) {
	output, err := processAttribute(pid, "lstart=")
	if err != nil {
		return time.Time{}, err
	}
	startedAt, err := time.ParseInLocation("Mon Jan _2 15:04:05 2006", strings.TrimSpace(output), time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse process %d start time: %w", pid, err)
	}
	return startedAt.UTC(), nil
}

func processAttribute(pid int, attribute string) (string, error) {
	command := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", attribute)
	command.Env = append(os.Environ(), "LC_ALL=C", "LANG=C")
	output, err := command.Output()
	if err != nil {
		return "", fmt.Errorf("inspect process %d: %w", pid, err)
	}
	return string(output), nil
}

func canonicalExecutable(path string) (string, error) {
	path = strings.TrimSpace(path)
	if !filepath.IsAbs(path) {
		return path, nil
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve executable %s: %w", path, err)
	}
	if resolved, err := filepath.EvalSymlinks(absPath); err == nil {
		return filepath.Clean(resolved), nil
	}
	return filepath.Clean(absPath), nil
}
