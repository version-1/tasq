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

func TestNewCurrentServiceStateRecordsCurrentProcessIdentity(t *testing.T) {
	originalReadProcessIdentity := readProcessIdentity
	defer func() { readProcessIdentity = originalReadProcessIdentity }()

	startedAt := time.Date(2026, time.September, 6, 3, 4, 5, 0, time.UTC)
	readProcessIdentity = func(pid int) (ProcessIdentity, error) {
		if pid != os.Getpid() {
			t.Fatalf("identity pid = %d, want %d", pid, os.Getpid())
		}
		return ProcessIdentity{StartedAt: startedAt, Executable: "/usr/local/bin/tq"}, nil
	}

	state, err := NewCurrentServiceState("127.0.0.1:37651", "/tmp/issues.sqlite")
	if err != nil {
		t.Fatalf("new current service state: %v", err)
	}
	if state.PID != os.Getpid() || state.Addr != "127.0.0.1:37651" || state.DB != "/tmp/issues.sqlite" {
		t.Fatalf("service state = %+v", state)
	}
	if !state.ProcessStartedAt.Equal(startedAt) || state.Executable != "/usr/local/bin/tq" {
		t.Fatalf("process identity = %+v", state)
	}
}

func TestServiceStateHasSameProcessIdentity(t *testing.T) {
	startedAt := time.Date(2026, time.September, 6, 3, 4, 5, 0, time.UTC)
	state := ServiceState{PID: 42, ProcessStartedAt: startedAt, Executable: "/usr/local/bin/tq"}

	if !state.HasSameProcessIdentity(state) {
		t.Fatal("same state does not have the same process identity")
	}
	if state.HasSameProcessIdentity(ServiceState{PID: 42, ProcessStartedAt: startedAt.Add(time.Second), Executable: "/usr/local/bin/tq"}) {
		t.Fatal("different start time has the same process identity")
	}
	if state.HasSameProcessIdentity(ServiceState{PID: 43, ProcessStartedAt: startedAt, Executable: "/usr/local/bin/tq"}) {
		t.Fatal("different pid has the same process identity")
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

func TestServiceStateMatchesProcessIdentityRegardlessOfObservedExecutable(t *testing.T) {
	identity := ProcessIdentity{
		StartedAt:  time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC),
		Executable: "/tmp/tasq",
	}
	original := readProcessIdentity
	readProcessIdentity = func(int) (ProcessIdentity, error) { return identity, nil }
	t.Cleanup(func() { readProcessIdentity = original })

	state := ServiceState{
		PID:              42,
		ProcessStartedAt: identity.StartedAt,
		Executable:       "/tmp/services/web",
	}
	matches, err := state.MatchesProcessIdentity()
	if err != nil {
		t.Fatalf("match process identity: %v", err)
	}
	if !matches {
		t.Fatal("process start time did not match")
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
