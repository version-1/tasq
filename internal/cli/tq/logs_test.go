package tq

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tqconfig "github.com/version-1/tasq/internal/config"
)

func TestResolveLogService(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantFile string
		wantErr  bool
	}{
		{name: "tracker alias", input: "tracker", wantName: "issue-tracker", wantFile: "issue-tracker.log"},
		{name: "issue tracker", input: "issue-tracker", wantName: "issue-tracker", wantFile: "issue-tracker.log"},
		{name: "orchestrator", input: "orchestrator", wantName: "orchestrator", wantFile: "orchestrator.log"},
		{name: "web", input: "web", wantName: "web", wantFile: "web.log"},
		{name: "unknown", input: "api", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service, err := resolveLogService(test.input)
			if test.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolve service: %v", err)
			}
			if service.name != test.wantName || service.file != test.wantFile {
				t.Fatalf("service=%+v", service)
			}
		})
	}
}

func TestLogsShowsLastLines(t *testing.T) {
	home := t.TempDir()
	t.Setenv(tqconfig.EnvHome, home)
	writeTestLog(t, home, "issue-tracker.log", "one\n", "two\n", "three\n")

	stdout, stderr, code := runCLI(t, []string{"logs", "tracker", "-n", "2"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if stdout != "two\nthree\n" {
		t.Fatalf("stdout=%q", stdout)
	}
}

func TestLogsMissingService(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{"logs"})
	if code != 2 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout: %s", stdout)
	}
	if got := decodeCLIError(t, stderr); got != "usage: tq logs <service> [-n lines] [-f]" {
		t.Fatalf("error=%q", got)
	}
}

func TestLogsInvalidService(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{"logs", "api"})
	if code != 2 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout: %s", stdout)
	}
	if got := decodeCLIError(t, stderr); got != `unknown log service "api"` {
		t.Fatalf("error=%q", got)
	}
}

func TestLogsMissingFile(t *testing.T) {
	t.Setenv(tqconfig.EnvHome, t.TempDir())

	stdout, stderr, code := runCLI(t, []string{"logs", "web"})
	if code != 1 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout: %s", stdout)
	}
	if !strings.Contains(decodeCLIError(t, stderr), "log file not found:") {
		t.Fatalf("stderr=%s", stderr)
	}
}

func TestLogsRejectsJSONOutput(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{"--output", "json", "logs", "tracker"})
	if code != 2 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout: %s", stdout)
	}
	if got := decodeCLIError(t, stderr); got != "logs does not support json output" {
		t.Fatalf("error=%q", got)
	}
}

func TestCopyAppendedLog(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service.log")
	if err := os.WriteFile(path, []byte("before\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	offset, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("before\nafter\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	nextOffset, err := copyAppendedLog(&output, file, offset)
	if err != nil {
		t.Fatalf("copy appended: %v", err)
	}
	if output.String() != "after\n" {
		t.Fatalf("output=%q", output.String())
	}
	if nextOffset != int64(len("before\nafter\n")) {
		t.Fatalf("nextOffset=%d", nextOffset)
	}
}

func TestWriteServiceLogFollowOnlySkipsPastOutput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "service.log")
	if err := os.WriteFile(path, []byte("past\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var output bytes.Buffer
	if err := writeServiceLog(ctx, &output, path, 0, true); err != nil {
		t.Fatalf("write service log: %v", err)
	}
	if output.String() != "" {
		t.Fatalf("output=%q", output.String())
	}
}

func writeTestLog(t *testing.T, home string, name string, lines ...string) {
	t.Helper()
	dir := tqconfig.LogDir(home)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(strings.Join(lines, "")), 0o644); err != nil {
		t.Fatal(err)
	}
}
