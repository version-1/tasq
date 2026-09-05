package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadStateMissingReturnsEmpty(t *testing.T) {
	t.Setenv(EnvHome, t.TempDir())
	state, err := ReadState()
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if state.IssueTracker != nil || state.Orchestrator != nil || state.Web != nil {
		t.Fatalf("state=%+v", state)
	}
}

func TestUpdateStateWritesAtomically(t *testing.T) {
	home := t.TempDir()
	t.Setenv(EnvHome, home)
	startedAt := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	if err := UpdateState(func(state *State) error {
		state.IssueTracker = &ServiceState{
			PID:       123,
			Addr:      "127.0.0.1:51234",
			DB:        "system/data/issues.sqlite",
			StartedAt: startedAt,
		}
		state.Web = &ServiceState{
			PID:       124,
			Addr:      "127.0.0.1:37653",
			StartedAt: startedAt,
		}
		return nil
	}); err != nil {
		t.Fatalf("update state: %v", err)
	}
	state, err := ReadState()
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if state.IssueTracker == nil || state.IssueTracker.PID != 123 || !state.IssueTracker.StartedAt.Equal(startedAt) {
		t.Fatalf("state=%+v", state)
	}
	if state.Web == nil || state.Web.PID != 124 || state.Web.Addr != "127.0.0.1:37653" {
		t.Fatalf("state=%+v", state)
	}
	if _, err := os.Stat(StateLockPath(home)); err != nil {
		t.Fatalf("lock missing: %v", err)
	}
}

func TestIssueTrackerURLFromState(t *testing.T) {
	t.Setenv(EnvHome, t.TempDir())
	if err := UpdateState(func(state *State) error {
		state.IssueTracker = &ServiceState{Addr: "127.0.0.1:51234"}
		return nil
	}); err != nil {
		t.Fatalf("update state: %v", err)
	}
	apiURL, ok, err := IssueTrackerURLFromState()
	if err != nil {
		t.Fatalf("url from state: %v", err)
	}
	if !ok || apiURL != "http://127.0.0.1:51234" {
		t.Fatalf("apiURL=%q ok=%v", apiURL, ok)
	}
}

func TestOrchestratorURLFromState(t *testing.T) {
	t.Setenv(EnvHome, t.TempDir())
	if err := UpdateState(func(state *State) error {
		state.Orchestrator = &ServiceState{Addr: "http://127.0.0.1:51235/"}
		return nil
	}); err != nil {
		t.Fatalf("update state: %v", err)
	}
	apiURL, ok, err := OrchestratorURLFromState()
	if err != nil {
		t.Fatalf("url from state: %v", err)
	}
	if !ok || apiURL != "http://127.0.0.1:51235" {
		t.Fatalf("apiURL=%q ok=%v", apiURL, ok)
	}
}

func TestServiceStateMatchesCurrentProcessIdentity(t *testing.T) {
	identity := ProcessIdentity{
		StartedAt:  time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC),
		Executable: "/tmp/orchestrator",
	}
	original := readProcessIdentity
	readProcessIdentity = func(pid int) (ProcessIdentity, error) {
		if pid != os.Getpid() {
			t.Fatalf("pid=%d, want %d", pid, os.Getpid())
		}
		return identity, nil
	}
	t.Cleanup(func() { readProcessIdentity = original })
	state := ServiceState{
		PID:              os.Getpid(),
		ProcessStartedAt: identity.StartedAt,
		Executable:       identity.Executable,
	}
	matches, err := state.MatchesProcessIdentity()
	if err != nil {
		t.Fatalf("match current process identity: %v", err)
	}
	if !matches {
		t.Fatal("current process identity did not match")
	}

	state.ProcessStartedAt = state.ProcessStartedAt.Add(time.Second)
	matches, err = state.MatchesProcessIdentity()
	if err != nil {
		t.Fatalf("match changed process identity: %v", err)
	}
	if matches {
		t.Fatal("changed process identity matched")
	}
}

func TestReadStateParsesExistingFile(t *testing.T) {
	home := t.TempDir()
	t.Setenv(EnvHome, home)
	if err := os.MkdirAll(SystemDir(home), 0o755); err != nil {
		t.Fatal(err)
	}
	raw := `{"orchestrator":{"pid":7,"addr":"http://127.0.0.1:8081","started_at":"2026-06-01T10:00:00Z"}}`
	if err := os.WriteFile(filepath.Join(SystemDir(home), "state.json"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	state, err := ReadState()
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if state.Orchestrator == nil || state.Orchestrator.PID != 7 {
		t.Fatalf("state=%+v", state)
	}
	if state.Web != nil {
		t.Fatalf("web should be empty for legacy state: %+v", state)
	}
}
