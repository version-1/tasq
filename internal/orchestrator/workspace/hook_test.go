package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunHookRunsScriptInWorkspace(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	if err := RunHook(`pwd > hook.out`, dir, time.Second); err != nil {
		t.Fatalf("run hook: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "hook.out"))
	if err != nil {
		t.Fatalf("read hook output: %v", err)
	}
	if strings.TrimSpace(string(got)) != dir {
		t.Fatalf("hook cwd = %q, want %q", strings.TrimSpace(string(got)), dir)
	}
}

func TestRunHookReturnsFailureWithStderr(t *testing.T) {
	t.Parallel()

	err := RunHook(`echo "bad hook" >&2; exit 7`, t.TempDir(), time.Second)
	if err == nil {
		t.Fatal("expected hook failure")
	}
	if !strings.Contains(err.Error(), "bad hook") {
		t.Fatalf("error = %v", err)
	}
}

func TestRunHookTimesOut(t *testing.T) {
	t.Parallel()

	err := RunHook(`sleep 2`, t.TempDir(), 20*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout")
	}
	if !errors.Is(err, ErrHookTimeout) {
		t.Fatalf("error = %v, want ErrHookTimeout", err)
	}
}

func TestRunHookEmptyScriptIsNoop(t *testing.T) {
	t.Parallel()

	if err := RunHook("", t.TempDir(), time.Second); err != nil {
		t.Fatalf("empty hook error = %v", err)
	}
}
