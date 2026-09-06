package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type State struct {
	IssueTracker *ServiceState `json:"issue_tracker,omitempty"`
	Orchestrator *ServiceState `json:"orchestrator,omitempty"`
	Web          *ServiceState `json:"web,omitempty"`
}

type ServiceState struct {
	PID              int       `json:"pid"`
	Addr             string    `json:"addr"`
	DB               string    `json:"db,omitempty"`
	StartedAt        time.Time `json:"started_at"`
	ProcessStartedAt time.Time `json:"process_started_at,omitempty"`
	Executable       string    `json:"executable,omitempty"`
}

func NewCurrentServiceState(addr, db string) (ServiceState, error) {
	identity, err := CurrentProcessIdentity()
	if err != nil {
		return ServiceState{}, err
	}
	return ServiceState{
		PID:              os.Getpid(),
		Addr:             addr,
		DB:               db,
		StartedAt:        time.Now().UTC(),
		ProcessStartedAt: identity.StartedAt,
		Executable:       identity.Executable,
	}, nil
}

func (state ServiceState) HasSameProcessIdentity(other ServiceState) bool {
	return state.PID == other.PID &&
		!state.ProcessStartedAt.IsZero() &&
		!other.ProcessStartedAt.IsZero() &&
		strings.TrimSpace(state.Executable) != "" &&
		strings.TrimSpace(other.Executable) != "" &&
		state.ProcessStartedAt.Equal(other.ProcessStartedAt) &&
		state.Executable == other.Executable
}

func ReadState() (State, error) {
	home, err := Home()
	if err != nil {
		return State{}, err
	}
	return readStateFile(StatePath(home))
}

func UpdateState(fn func(*State) error) error {
	home, err := EnsureHome()
	if err != nil {
		return err
	}
	lock, err := os.OpenFile(StateLockPath(home), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return fmt.Errorf("open state lock: %w", err)
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return fmt.Errorf("lock state: %w", err)
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)

	path := StatePath(home)
	state, err := readStateFile(path)
	if err != nil {
		return err
	}
	if err := fn(&state); err != nil {
		return err
	}
	return writeStateFile(path, state)
}

func IssueTrackerURLFromState() (string, bool, error) {
	state, err := ReadState()
	if err != nil {
		return "", false, err
	}
	if state.IssueTracker == nil || strings.TrimSpace(state.IssueTracker.Addr) == "" {
		return "", false, nil
	}
	return serviceURL(state.IssueTracker.Addr), true, nil
}

func WebURLFromState() (string, bool, error) {
	state, err := ReadState()
	if err != nil {
		return "", false, err
	}
	if state.Web == nil || strings.TrimSpace(state.Web.Addr) == "" {
		return "", false, nil
	}
	return serviceURL(state.Web.Addr), true, nil
}

func OrchestratorURLFromState() (string, bool, error) {
	state, err := ReadState()
	if err != nil {
		return "", false, err
	}
	if state.Orchestrator == nil || strings.TrimSpace(state.Orchestrator.Addr) == "" {
		return "", false, nil
	}
	return serviceURL(state.Orchestrator.Addr), true, nil
}

func readStateFile(path string) (State, error) {
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return State{}, nil
	}
	if err != nil {
		return State{}, fmt.Errorf("read state.json: %w", err)
	}
	if len(strings.TrimSpace(string(raw))) == 0 {
		return State{}, nil
	}
	var state State
	if err := json.Unmarshal(raw, &state); err != nil {
		return State{}, fmt.Errorf("parse state.json: %w", err)
	}
	return state, nil
}

func writeStateFile(path string, state State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}
	raw, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state.json: %w", err)
	}
	raw = append(raw, '\n')
	tmp, err := os.CreateTemp(filepath.Dir(path), ".state-*.json")
	if err != nil {
		return fmt.Errorf("create state temp: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(raw); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write state temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close state temp: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace state.json: %w", err)
	}
	return nil
}

func serviceURL(addr string) string {
	addr = strings.TrimSpace(addr)
	if parsed, err := url.Parse(addr); err == nil && parsed.Scheme != "" {
		return strings.TrimRight(addr, "/")
	}
	return "http://" + strings.TrimRight(addr, "/")
}
