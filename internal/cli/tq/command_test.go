package tq

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	tqconfig "github.com/version-1/tasq/internal/config"
	"github.com/version-1/tasq/internal/issue/api"
	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/issue/store"
	"github.com/version-1/tasq/internal/orchestrator/runstore"
)

func TestVersion(t *testing.T) {
	version, commit := versionInfo()
	want := fmt.Sprintf("tq %s (commit: %s)\n", version, commit)

	stdout, stderr, code := runCLI(t, []string{"version"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if stderr != "" {
		t.Fatalf("expected empty stderr: %s", stderr)
	}
	if stdout != want {
		t.Fatalf("stdout=%q, want %q", stdout, want)
	}
}

func TestConfigShowsResolvedSettings(t *testing.T) {
	home := t.TempDir()
	t.Setenv(tqconfig.EnvHome, home)
	if err := os.MkdirAll(tqconfig.ConfigDir(home), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tqconfig.ConfigPath(home), []byte("author: config-author\nmax_concurrent_agents: 4\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := runCLI(t, []string{"config"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	for _, want := range []string{
		"Profile: default",
		"TQ_HOME: " + home,
		"Home: " + home,
		"Config path: " + tqconfig.ConfigPath(home),
		"Author: config-author",
		"Max concurrent agents: 4",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q: %s", want, stdout)
		}
	}
}

func TestConfigWritesJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv(tqconfig.EnvHome, home)

	stdout, stderr, code := runCLI(t, []string{"--output", "json", "config"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	var info configInfo
	if err := json.Unmarshal([]byte(stdout), &info); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if info.TQHome == nil || *info.TQHome != home {
		t.Fatalf("TQHome=%v, want %q", info.TQHome, home)
	}
	if info.Home != home {
		t.Fatalf("Home=%q, want %q", info.Home, home)
	}
	if info.ConfigPath != tqconfig.ConfigPath(home) {
		t.Fatalf("ConfigPath=%q", info.ConfigPath)
	}
}

func TestIsPseudoVersion(t *testing.T) {
	tests := []struct {
		version string
		want    bool
	}{
		{version: "v0.1.0", want: false},
		{version: "v0.1.0-rc.1", want: false},
		{version: "v0.0.0-20260603232640-51f272dbd384", want: true},
		{version: "v0.0.0-20260603232640-51f272dbd384+dirty", want: true},
	}
	for _, test := range tests {
		if got := isPseudoVersion(test.version); got != test.want {
			t.Fatalf("isPseudoVersion(%q)=%t, want %t", test.version, got, test.want)
		}
	}
}

func TestVersionInfoUsesInjectedBuildMetadata(t *testing.T) {
	originalVersion := buildVersion
	originalCommit := buildCommit
	t.Cleanup(func() {
		buildVersion = originalVersion
		buildCommit = originalCommit
	})

	buildVersion = "v0.1.0"
	buildCommit = "abc1234"

	version, commit := versionInfo()
	if version != "v0.1.0" {
		t.Fatalf("version=%q, want %q", version, "v0.1.0")
	}
	if commit != "abc1234" {
		t.Fatalf("commit=%q, want %q", commit, "abc1234")
	}
}

func TestNewGHCommandDisablesInteractivePrompts(t *testing.T) {
	t.Setenv("GH_PROMPT_DISABLED", "")

	cmd := newGHCommand(context.Background(), "release", "download")
	for _, env := range cmd.Env {
		if env == "GH_PROMPT_DISABLED=1" {
			return
		}
	}
	t.Fatalf("GH_PROMPT_DISABLED=1 not found in command environment: %v", cmd.Env)
}

func TestUpdateRunsConfirmedFlowWithTag(t *testing.T) {
	runner := &fakeUpdateRunner{
		current:   "tq v0.1.0 (commit: old)",
		target:    "v0.2.0-rc.1",
		installed: "tq v0.2.0-rc.1 (commit: new)",
		confirmOK: true,
	}
	stdout, stderr, code := runCLIWithUpdateRunner(t, []string{"update", "--tag", "v0.2.0-rc.1"}, "yes\n", runner)
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	wantCalls := []string{
		"current",
		"target:v0.2.0-rc.1",
		"confirm",
		"stop",
		"install:v0.2.0-rc.1",
		"installed",
		"migrate",
		"start",
	}
	assertStringSlice(t, runner.calls, wantCalls)
	for _, want := range []string{
		"Current version: tq v0.1.0 (commit: old)",
		"Target version:  v0.2.0-rc.1",
		"Installed version: tq v0.2.0-rc.1 (commit: new)",
		"tq update complete",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q: %s", want, stdout)
		}
	}
}

func TestUpdateProfileAllowed(t *testing.T) {
	if err := updateProfileAllowed(""); err != nil {
		t.Fatalf("default profile: %v", err)
	}
	err := updateProfileAllowed("dev")
	if err == nil || !strings.Contains(err.Error(), "dev") {
		t.Fatalf("profile error=%v", err)
	}
}

func TestUpdateYesSkipsConfirmation(t *testing.T) {
	runner := &fakeUpdateRunner{
		current:   "tq v0.1.0 (commit: old)",
		target:    "v0.2.0",
		installed: "tq v0.2.0 (commit: new)",
	}
	_, stderr, code := runCLIWithUpdateRunner(t, []string{"update", "-y"}, "", runner)
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	wantCalls := []string{
		"current",
		"target:",
		"stop",
		"install:v0.2.0",
		"installed",
		"migrate",
		"start",
	}
	assertStringSlice(t, runner.calls, wantCalls)
}

func TestUpdateCancelledBeforeStoppingServices(t *testing.T) {
	runner := &fakeUpdateRunner{
		current:   "tq v0.1.0 (commit: old)",
		target:    "v0.2.0",
		confirmOK: false,
	}
	stdout, stderr, code := runCLIWithUpdateRunner(t, []string{"update"}, "no\n", runner)
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	assertStringSlice(t, runner.calls, []string{"current", "target:", "confirm"})
	if !strings.Contains(stdout, "Update cancelled") {
		t.Fatalf("stdout missing cancellation: %s", stdout)
	}
}

func TestUpdateStopsAfterInstallFailure(t *testing.T) {
	runner := &fakeUpdateRunner{
		current:    "tq v0.1.0 (commit: old)",
		target:     "v0.2.0",
		installErr: errors.New("download failed"),
	}
	stdout, stderr, code := runCLIWithUpdateRunner(t, []string{"update", "-y"}, "", runner)
	if code != 1 {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	assertStringSlice(t, runner.calls, []string{"current", "target:", "stop", "install:v0.2.0"})
	message := decodeCLIError(t, stderr)
	if !strings.Contains(message, "installing tq failed") || !strings.Contains(message, "download failed") {
		t.Fatalf("unexpected error: %s", message)
	}
}

func TestUpdateStopsBeforeMigrationWhenInstalledVersionCheckFails(t *testing.T) {
	runner := &fakeUpdateRunner{
		current:      "tq v0.1.0 (commit: old)",
		target:       "v0.2.0",
		installedErr: errors.New("not executable"),
	}
	stdout, stderr, code := runCLIWithUpdateRunner(t, []string{"update", "-y"}, "", runner)
	if code != 1 {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	assertStringSlice(t, runner.calls, []string{"current", "target:", "stop", "install:v0.2.0", "installed"})
	message := decodeCLIError(t, stderr)
	if !strings.Contains(message, "check installed version") || !strings.Contains(message, "not executable") {
		t.Fatalf("unexpected error: %s", message)
	}
}

func TestUpdateHelp(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{"update", "--help"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	for _, want := range []string{"Usage: tq update [-y] [--tag TAG]", "--tag TAG", "-y"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout missing %q: %s", want, stdout)
		}
	}
}

func TestIssueListJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/issues" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		writeTestJSON(t, w, apiResponse[[]entity.Issue]{
			Data: []entity.Issue{{ID: 7, Title: "Wire CLI", Status: entity.StatusReady}},
		})
	}))
	defer server.Close()

	stdout, stderr, code := runCLI(t, []string{"--api-url", server.URL, "--output", "json", "issue", "list"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, `"id": 7`) || !strings.Contains(stdout, `"title": "Wire CLI"`) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueCreate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects" {
			writeTestJSON(t, w, apiResponse[[]entity.Project]{
				Data: []entity.Project{{ID: 2, Key: "CLI", Name: "CLI"}},
			})
			return
		}
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/issues" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var input entity.CreateIssueInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if input.ProjectID != 2 || input.Title != "Add CLI" || input.Status != entity.StatusReady || input.Priority != entity.PriorityHigh {
			t.Fatalf("unexpected input: %+v", input)
		}
		w.WriteHeader(http.StatusCreated)
		writeTestJSON(t, w, apiResponse[entity.Issue]{Data: entity.Issue{
			ID:         10,
			ProjectID:  2,
			ProjectKey: "CLI",
			Title:      input.Title,
			Status:     input.Status,
			Priority:   input.Priority,
			CreatedAt:  time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC),
			UpdatedAt:  time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC),
		}})
	}))
	defer server.Close()

	stdout, stderr, code := runCLI(t, []string{
		"issue", "create",
		"--api-url", server.URL,
		"--title", "Add CLI",
		"--project", "CLI",
		"--status", "ready",
		"--priority", "high",
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "ID: 10") || !strings.Contains(stdout, "Project: CLI") || !strings.Contains(stdout, "ready") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueCreateSendsDependencyIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects" {
			writeTestJSON(t, w, apiResponse[[]entity.Project]{
				Data: []entity.Project{{ID: 2, Key: "CLI", Name: "CLI"}},
			})
			return
		}
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/issues" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var input entity.CreateIssueInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if input.ProjectID != 2 || input.Title != "Dependent issue" {
			t.Fatalf("unexpected input: %+v", input)
		}
		assertInt64s(t, input.DependencyIDs, []int64{4, 7})
		w.WriteHeader(http.StatusCreated)
		writeTestJSON(t, w, apiResponse[entity.Issue]{Data: entity.Issue{
			ID:            12,
			ProjectID:     2,
			ProjectKey:    "CLI",
			Title:         input.Title,
			DependencyIDs: input.DependencyIDs,
		}})
	}))
	defer server.Close()

	stdout, stderr, code := runCLI(t, []string{
		"--api-url", server.URL,
		"issue", "create",
		"--title", "Dependent issue",
		"--project", "CLI",
		"--dependency", "4,7",
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "ID: 12") || !strings.Contains(stdout, "Dependencies: 4,7") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueCreateRequiresProject(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{
		"--api-url", defaultAPIURL,
		"issue", "create",
		"--title", "Missing project",
	})
	if code != 2 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout: %s", stdout)
	}
	if got := decodeCLIError(t, stderr); got != "project is required" {
		t.Fatalf("error=%q", got)
	}
}

func TestIssueCreateDependencyUsageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "empty dependency",
			args: []string{"issue", "create", "--title", "Invalid dependency", "--project", "CLI", "--dependency", ""},
			want: "dependency must not be empty",
		},
		{
			name: "invalid dependency",
			args: []string{"issue", "create", "--title", "Invalid dependency", "--project", "CLI", "--dependency", "4,0"},
			want: "dependency must be a comma-separated list of positive integers",
		},
		{
			name: "negative dependency",
			args: []string{"issue", "create", "--title", "Invalid dependency", "--project", "CLI", "--dependency", "-1"},
			want: "dependency must be a comma-separated list of positive integers",
		},
		{
			name: "non integer dependency",
			args: []string{"issue", "create", "--title", "Invalid dependency", "--project", "CLI", "--dependency", "abc"},
			want: "dependency must be a comma-separated list of positive integers",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stdout, stderr, code := runCLI(t, append([]string{"--api-url", defaultAPIURL}, test.args...))
			if code != 2 {
				t.Fatalf("code=%d stderr=%s", code, stderr)
			}
			if stdout != "" {
				t.Fatalf("expected empty stdout: %s", stdout)
			}
			if got := decodeCLIError(t, stderr); got != test.want {
				t.Fatalf("error=%q, want %q", got, test.want)
			}
		})
	}
}

func TestIssueUpdateSendsSpecifiedFieldsOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/v1/issues/12" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var input map[string]string
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if len(input) != 2 || input["status"] != "blocked" || input["assignee"] != "" {
			t.Fatalf("unexpected input: %+v", input)
		}
		writeTestJSON(t, w, apiResponse[entity.Issue]{
			Data: entity.Issue{ID: 12, Title: "Blocked issue", Status: entity.StatusBlocked},
		})
	}))
	defer server.Close()

	stdout, stderr, code := runCLI(t, []string{
		"--api-url=" + server.URL,
		"issue", "update", "12",
		"--status", "blocked",
		"--assignee", "",
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "blocked") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueUpdateReplacesDependencies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/v1/issues/12" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var input updateIssueInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if input.DependencyIDs == nil {
			t.Fatal("dependency_ids was not sent")
		}
		assertInt64s(t, *input.DependencyIDs, []int64{4, 7})
		writeTestJSON(t, w, apiResponse[entity.Issue]{
			Data: entity.Issue{ID: 12, Title: "Dependent issue", DependencyIDs: *input.DependencyIDs},
		})
	}))
	defer server.Close()

	stdout, stderr, code := runCLI(t, []string{
		"--api-url=" + server.URL,
		"issue", "update", "12",
		"--dependency", "4,7",
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "ID: 12") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueUpdateAllowsExplicitFalseClearDependencies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/v1/issues/12" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var input updateIssueInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if input.DependencyIDs == nil {
			t.Fatal("dependency_ids was not sent")
		}
		assertInt64s(t, *input.DependencyIDs, []int64{4})
		writeTestJSON(t, w, apiResponse[entity.Issue]{
			Data: entity.Issue{ID: 12, Title: "Dependent issue", DependencyIDs: *input.DependencyIDs},
		})
	}))
	defer server.Close()

	stdout, stderr, code := runCLI(t, []string{
		"--api-url=" + server.URL,
		"issue", "update", "12",
		"--dependency", "4",
		"--clear-dependencies=false",
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "ID: 12") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueUpdateClearsDependencies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/v1/issues/12" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var input updateIssueInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if input.DependencyIDs == nil {
			t.Fatal("dependency_ids was not sent")
		}
		assertInt64s(t, *input.DependencyIDs, []int64{})
		writeTestJSON(t, w, apiResponse[entity.Issue]{
			Data: entity.Issue{ID: 12, Title: "Independent issue", DependencyIDs: *input.DependencyIDs},
		})
	}))
	defer server.Close()

	stdout, stderr, code := runCLI(t, []string{
		"--api-url=" + server.URL,
		"issue", "update", "12",
		"--clear-dependencies",
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Dependencies: none") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueUpdateDependencyUsageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "dependency and clear",
			args: []string{"issue", "update", "12", "--dependency", "4", "--clear-dependencies"},
			want: "dependency and clear-dependencies cannot be used together",
		},
		{
			name: "empty dependency",
			args: []string{"issue", "update", "12", "--dependency", ""},
			want: "dependency must not be empty; use --clear-dependencies to clear dependencies",
		},
		{
			name: "non integer dependency",
			args: []string{"issue", "update", "12", "--dependency", "abc"},
			want: "dependency must be a comma-separated list of positive integers",
		},
		{
			name: "zero dependency",
			args: []string{"issue", "update", "12", "--dependency", "4,0"},
			want: "dependency must be a comma-separated list of positive integers",
		},
		{
			name: "duplicate dependency",
			args: []string{"issue", "update", "12", "--dependency", "4,4"},
			want: "dependency contains duplicate issue id",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stdout, stderr, code := runCLI(t, append([]string{"--api-url", defaultAPIURL}, test.args...))
			if code != 2 {
				t.Fatalf("code=%d stderr=%s", code, stderr)
			}
			if stdout != "" {
				t.Fatalf("expected empty stdout: %s", stdout)
			}
			if got := decodeCLIError(t, stderr); got != test.want {
				t.Fatalf("error=%q, want %q", got, test.want)
			}
		})
	}
}

func TestIssueUpdateDependencyValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/v1/issues/12" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusBadRequest)
		writeTestJSON(t, w, apiErrorResponse{
			Error: struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}{Code: "issues.update.invalid_dependency_cycle", Message: "dependency cycle detected: 12 -> 4 -> 12"},
		})
	}))
	defer server.Close()

	stdout, stderr, code := runCLI(t, []string{
		"--api-url=" + server.URL,
		"issue", "update", "12",
		"--dependency", "4",
	})
	if code != 1 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout: %s", stdout)
	}
	if got := decodeCLIError(t, stderr); got != "dependency cycle detected: 12 -> 4 -> 12" {
		t.Fatalf("error=%q", got)
	}
}

func TestIssueClose(t *testing.T) {
	stdout := assertIssueShortcut(t, issueShortcutTest{
		args:        []string{"issue", "close", "5"},
		id:          5,
		wantPatch:   map[string]string{"status": "done"},
		response:    entity.Issue{ID: 5, Title: "Close issue", Status: entity.StatusDone},
		wantMessage: ansiGreen + "✓" + ansiReset + " Issue #5 closed",
	})
	if !strings.Contains(stdout, string(entity.StatusDone)) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueCancel(t *testing.T) {
	stdout := assertIssueShortcut(t, issueShortcutTest{
		args:        []string{"issue", "cancel", "5"},
		id:          5,
		wantPatch:   map[string]string{"status": "cancelled"},
		response:    entity.Issue{ID: 5, Title: "Cancel issue", Status: entity.StatusCancelled},
		wantMessage: ansiGreen + "✓" + ansiReset + " Issue #5 cancelled",
	})
	if !strings.Contains(stdout, string(entity.StatusCancelled)) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueReady(t *testing.T) {
	stdout := assertIssueShortcut(t, issueShortcutTest{
		args:        []string{"issue", "ready", "3"},
		id:          3,
		wantPatch:   map[string]string{"status": "ready"},
		response:    entity.Issue{ID: 3, Title: "Ready issue", Status: entity.StatusReady},
		wantMessage: ansiGreen + "✓" + ansiReset + " Issue #3 marked as ready",
	})
	if !strings.Contains(stdout, string(entity.StatusReady)) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueDraft(t *testing.T) {
	stdout := assertIssueShortcut(t, issueShortcutTest{
		args:        []string{"issue", "draft", "8"},
		id:          8,
		wantPatch:   map[string]string{"status": "backlog"},
		response:    entity.Issue{ID: 8, Title: "Draft issue", Status: entity.StatusBacklog},
		wantMessage: ansiGreen + "✓" + ansiReset + " Issue #8 moved to backlog",
	})
	if !strings.Contains(stdout, string(entity.StatusBacklog)) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueRename(t *testing.T) {
	stdout := assertIssueShortcut(t, issueShortcutTest{
		args:        []string{"issue", "rename", "6", "Better title"},
		id:          6,
		wantPatch:   map[string]string{"title": "Better title"},
		response:    entity.Issue{ID: 6, Title: "Better title", Status: entity.StatusReady},
		wantMessage: ansiGreen + "✓" + ansiReset + " Issue #6 renamed",
	})
	if !strings.Contains(stdout, "Title: Better title") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueEdit(t *testing.T) {
	stdout := assertIssueShortcut(t, issueShortcutTest{
		args:        []string{"issue", "edit", "6", "New description"},
		id:          6,
		wantPatch:   map[string]string{"description": "New description"},
		response:    entity.Issue{ID: 6, Title: "Edit issue", Description: "New description", Status: entity.StatusReady},
		wantMessage: ansiGreen + "✓" + ansiReset + " Issue #6 description updated",
	})
	if !strings.Contains(stdout, "Description: New description") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueShortcutJSONOmitsActionMessage(t *testing.T) {
	stdout := assertIssueShortcut(t, issueShortcutTest{
		args:      []string{"--output", "json", "issue", "close", "5"},
		id:        5,
		wantPatch: map[string]string{"status": "done"},
		response:  entity.Issue{ID: 5, Title: "Close issue", Status: entity.StatusDone},
	})
	if strings.Contains(stdout, "Issue #5 closed") || strings.Contains(stdout, ansiGreen) {
		t.Fatalf("unexpected action message or ANSI in JSON stdout: %s", stdout)
	}
	if !strings.Contains(stdout, `"status": "done"`) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueShortcutUsageErrors(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "close", args: []string{"issue", "close"}, want: "usage: tq issue close <id>"},
		{name: "cancel", args: []string{"issue", "cancel"}, want: "usage: tq issue cancel <id>"},
		{name: "ready", args: []string{"issue", "ready"}, want: "usage: tq issue ready <id>"},
		{name: "draft", args: []string{"issue", "draft"}, want: "usage: tq issue draft <id>"},
		{name: "rename", args: []string{"issue", "rename", "6"}, want: "usage: tq issue rename <id> <title>"},
		{name: "edit", args: []string{"issue", "edit", "6"}, want: "usage: tq issue edit <id> <description>"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stdout, stderr, code := runCLI(t, append([]string{"--api-url", defaultAPIURL}, test.args...))
			if code != 2 {
				t.Fatalf("code=%d stderr=%s", code, stderr)
			}
			if stdout != "" {
				t.Fatalf("expected empty stdout: %s", stdout)
			}
			if got := decodeCLIError(t, stderr); got != test.want {
				t.Fatalf("error=%q, want %q", got, test.want)
			}
		})
	}
}

func TestCommentAdd(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/issues/42/comments" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var input entity.CreateCommentInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if input.Author != "codex" || input.Type != entity.CommentBlocker || input.Body != "Blocked on credentials" {
			t.Fatalf("unexpected input: %+v", input)
		}
		w.WriteHeader(http.StatusCreated)
		writeTestJSON(t, w, apiResponse[entity.Comment]{Data: entity.Comment{
			ID:        3,
			IssueID:   42,
			Author:    input.Author,
			Type:      input.Type,
			Body:      input.Body,
			CreatedAt: time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC),
		}})
	}))
	defer server.Close()

	stdout, stderr, code := runCLI(t, []string{
		"--api-url", server.URL,
		"comment", "add", "42",
		"--author", "codex",
		"--type", "blocker",
		"--body", "Blocked on credentials",
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "ID: 3") || !strings.Contains(stdout, "Type: blocker") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestServiceStatusStopped(t *testing.T) {
	t.Setenv(tqconfig.EnvHome, t.TempDir())

	stdout, stderr, code := runCLI(t, []string{"service", "status"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	want := "SERVICE        STATE    PID  PORT  UPTIME\n" +
		"issue-tracker  stopped    -     -       -\n" +
		"orchestrator   stopped    -     -       -\n" +
		"web            stopped    -     -       -\n"
	if got := stripANSI(stdout); got != want {
		t.Fatalf("service status table:\n%s\nwant:\n%s", got, want)
	}
}

func TestServiceStartFailsBeforeStartingServicesWhenMigrationsPending(t *testing.T) {
	home := t.TempDir()
	t.Setenv(tqconfig.EnvHome, home)
	for _, name := range []serviceName{serviceIssueTracker, serviceOrchestrator, serviceWeb} {
		path := serviceExecutablePath(home, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create service bin directory: %v", err)
		}
		if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatalf("write service executable: %v", err)
		}
	}

	stdout, stderr, code := runCLI(t, []string{"service", "start"})
	if code != 1 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout: %s", stdout)
	}
	message := decodeCLIError(t, stderr)
	for _, want := range []string{
		"migration pre-flight check failed",
		"issue-tracker:20260615000000_init",
		"orchestrator:20260615000000_init",
		"run `tq migrate`",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("error missing %q: %s", want, message)
		}
	}
	for _, logName := range []string{"issue-tracker.log", "orchestrator.log", "web.log"} {
		path := filepath.Join(tqconfig.LogDir(home), logName)
		if _, err := os.Stat(path); err == nil {
			t.Fatalf("service log was created before migration pre-flight failed: %s", path)
		} else if !os.IsNotExist(err) {
			t.Fatalf("stat %s: %v", path, err)
		}
	}
}

func TestServiceMigrationPreflightPassesAfterMigrate(t *testing.T) {
	home := t.TempDir()
	t.Setenv(tqconfig.EnvHome, home)

	stdout, stderr, code := runCLI(t, []string{"migrate"})
	if code != 0 {
		t.Fatalf("migrate code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	if err := checkMigrationTargetsNoPending(context.Background()); err != nil {
		t.Fatalf("pre-flight after migrate: %v", err)
	}
}

func TestMigrateAppliesLocalDatabases(t *testing.T) {
	home := t.TempDir()
	t.Setenv(tqconfig.EnvHome, home)

	stdout, stderr, code := runCLI(t, []string{"migrate", "status"})
	if code != 0 {
		t.Fatalf("status code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "issue-tracker") || !strings.Contains(stdout, "pending") {
		t.Fatalf("status stdout missing pending migrations: %s", stdout)
	}

	stdout, stderr, code = runCLI(t, []string{"migrate"})
	if code != 0 {
		t.Fatalf("migrate code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stripANSI(stdout), "applied 20260615000000_init") {
		t.Fatalf("migrate stdout missing init apply: %s", stdout)
	}

	issueStore, err := store.Open(context.Background(), tqconfig.IssueTrackerDBPath(home))
	if err != nil {
		t.Fatalf("open migrated issue store: %v", err)
	}
	_ = issueStore.Close()
	orchestratorStore, err := runstore.Open(context.Background(), tqconfig.OrchestratorDBPath(home))
	if err != nil {
		t.Fatalf("open migrated orchestrator store: %v", err)
	}
	_ = orchestratorStore.Close()

	stdout, stderr, code = runCLI(t, []string{"migrate", "down"})
	if code != 0 {
		t.Fatalf("down code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "rolled back") {
		t.Fatalf("down stdout missing rollback: %s", stdout)
	}

	stdout, stderr, code = runCLI(t, []string{"migrate", "status"})
	if code != 0 {
		t.Fatalf("status after down code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "pending") {
		t.Fatalf("status after down stdout missing pending migration: %s", stdout)
	}
}

func TestServiceStatusJSONRunning(t *testing.T) {
	t.Setenv(tqconfig.EnvHome, t.TempDir())
	startedAt := time.Now().Add(-time.Minute).UTC()
	if err := tqconfig.UpdateState(func(state *tqconfig.State) error {
		state.IssueTracker = &tqconfig.ServiceState{
			PID:       os.Getpid(),
			Addr:      "127.0.0.1:" + strconv.Itoa(tqconfig.DefaultIssueTrackerPort),
			DB:        "/tmp/issues.sqlite",
			StartedAt: startedAt,
		}
		state.Web = &tqconfig.ServiceState{
			PID:       os.Getpid(),
			Addr:      "127.0.0.1:" + strconv.Itoa(tqconfig.DefaultWebPort),
			StartedAt: startedAt,
		}
		return nil
	}); err != nil {
		t.Fatalf("write state: %v", err)
	}

	stdout, stderr, code := runCLI(t, []string{"--output", "json", "service", "status"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	var statuses []serviceStatus
	if err := json.Unmarshal([]byte(stdout), &statuses); err != nil {
		t.Fatalf("decode stdout: %v: %s", err, stdout)
	}
	if len(statuses) != 3 {
		t.Fatalf("statuses=%+v", statuses)
	}
	if statuses[0].Name != "issue-tracker" || statuses[0].State != "running" || statuses[0].PID != os.Getpid() || statuses[0].Port != tqconfig.DefaultIssueTrackerPort {
		t.Fatalf("unexpected issue-tracker status: %+v", statuses[0])
	}
	if statuses[1].Name != "orchestrator" || statuses[1].State != "stopped" {
		t.Fatalf("unexpected orchestrator status: %+v", statuses[1])
	}
	if statuses[2].Name != "web" || statuses[2].State != "running" || statuses[2].PID != os.Getpid() || statuses[2].Port != tqconfig.DefaultWebPort {
		t.Fatalf("unexpected web status: %+v", statuses[2])
	}
}

func TestServiceStatusCleansStaleState(t *testing.T) {
	t.Setenv(tqconfig.EnvHome, t.TempDir())
	if err := tqconfig.UpdateState(func(state *tqconfig.State) error {
		state.IssueTracker = &tqconfig.ServiceState{
			PID:       -1,
			Addr:      "127.0.0.1:" + strconv.Itoa(tqconfig.DefaultIssueTrackerPort),
			StartedAt: time.Now().Add(-time.Hour),
		}
		return nil
	}); err != nil {
		t.Fatalf("write state: %v", err)
	}

	stdout, stderr, code := runCLI(t, []string{"service", "status"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stripANSI(stdout), "issue-tracker  stopped") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
	state, err := tqconfig.ReadState()
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	if state.IssueTracker != nil {
		t.Fatalf("stale issue-tracker state was not removed: %+v", state.IssueTracker)
	}
}

func TestServiceUnknownAction(t *testing.T) {
	t.Setenv(tqconfig.EnvHome, t.TempDir())

	stdout, stderr, code := runCLI(t, []string{"service", "restart"})
	if code != 2 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout: %s", stdout)
	}
	if got := decodeCLIError(t, stderr); got != `unknown service action "restart"` {
		t.Fatalf("error=%q", got)
	}
}

func TestCommentAddDefaultsAuthorFromEnvironment(t *testing.T) {
	t.Setenv("TQ_AUTHOR", "agent")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input entity.CreateCommentInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if input.Author != "agent" || input.Type != entity.CommentGeneral || input.Body != "Progress" {
			t.Fatalf("unexpected input: %+v", input)
		}
		w.WriteHeader(http.StatusCreated)
		writeTestJSON(t, w, apiResponse[entity.Comment]{Data: entity.Comment{
			ID:      4,
			IssueID: 42,
			Author:  input.Author,
			Type:    input.Type,
			Body:    input.Body,
		}})
	}))
	defer server.Close()

	stdout, stderr, code := runCLI(t, []string{
		"--api-url", server.URL,
		"comment", "add", "42",
		"--body", "Progress",
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Author: agent") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestCommentListJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/issues/42/comments" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		writeTestJSON(t, w, apiResponse[[]entity.Comment]{
			Data: []entity.Comment{{ID: 7, IssueID: 42, Author: "codex", Type: entity.CommentProgress, Body: "Working"}},
		})
	}))
	defer server.Close()

	stdout, stderr, code := runCLI(t, []string{"--api-url", server.URL, "--output", "json", "comment", "list", "42"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, `"id": 7`) || !strings.Contains(stdout, `"body": "Working"`) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestCommentCommandsAgainstIssueTrackerAPI(t *testing.T) {
	ctx := context.Background()
	issueStore, err := store.OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer issueStore.Close()

	project, err := issueStore.CreateProject(ctx, entity.CreateProjectInput{Key: "COMMENTS", Name: "Comments", Location: t.TempDir()})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	issue, err := issueStore.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Comment through CLI"})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	server := httptest.NewServer(api.NewServer(issueStore).Handler())
	defer server.Close()

	stdout, stderr, code := runCLI(t, []string{
		"--api-url", server.URL,
		"comment", "add", stringID(issue.ID),
		"--author", "codex",
		"--type", "handoff",
		"--body", "Ready for review",
	})
	if code != 0 {
		t.Fatalf("add code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Type: handoff") {
		t.Fatalf("unexpected add stdout: %s", stdout)
	}

	stdout, stderr, code = runCLI(t, []string{"--api-url", server.URL, "comment", "list", stringID(issue.ID)})
	if code != 0 {
		t.Fatalf("list code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Ready for review") {
		t.Fatalf("unexpected list stdout: %s", stdout)
	}
}

func TestArtifactCommandsAgainstIssueTrackerAPI(t *testing.T) {
	ctx := context.Background()
	issueStore, err := store.OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer issueStore.Close()
	project, err := issueStore.CreateProject(ctx, entity.CreateProjectInput{Key: "ARTCLI", Name: "Artifacts", Location: t.TempDir()})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	issue, err := issueStore.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Artifact through CLI"})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	server := httptest.NewServer(api.NewServer(issueStore).Handler())
	defer server.Close()
	args := []string{"--api-url", server.URL, "artifact", "set", stringID(issue.ID), "--type", "pull_request", "https://example.com/one"}
	stdout, stderr, code := runCLI(t, args)
	if code != 0 || stderr != "" || !strings.Contains(stdout, "https://example.com/one") {
		t.Fatalf("set: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	_, stderr, code = runCLI(t, []string{"--api-url", server.URL, "artifact", "set", stringID(issue.ID), "--type", "pull_request", "https://example.com/two"})
	if code != 0 || stderr != "" {
		t.Fatalf("overwrite: code=%d stderr=%q", code, stderr)
	}
	_, stderr, code = runCLI(t, []string{"--api-url", server.URL, "artifact", "set", "0", "--type", "pull_request", "https://example.com"})
	if code == 0 || !strings.Contains(stderr, "id must be a positive integer") {
		t.Fatalf("invalid ID: code=%d stderr=%q", code, stderr)
	}
	_, stderr, code = runCLI(t, []string{"--api-url", server.URL, "artifact", "set", stringID(issue.ID), "https://example.com"})
	if code == 0 || !strings.Contains(stderr, "type is required") {
		t.Fatalf("missing type: code=%d stderr=%q", code, stderr)
	}
	read, err := issueStore.Issue(ctx, issue.ID)
	if err != nil || len(read.Artifacts) != 1 || read.Artifacts[0].DataValue != "https://example.com/two" {
		t.Fatalf("issue artifacts = %#v, err=%v", read.Artifacts, err)
	}
	stdout, stderr, code = runCLI(t, []string{"--api-url", server.URL, "--output", "json", "artifact", "delete", stringID(issue.ID), "--type", "pull_request"})
	if code != 0 || stderr != "" {
		t.Fatalf("delete: code=%d stderr=%q", code, stderr)
	}
	var deleted artifactDeleteResult
	if err := json.Unmarshal([]byte(stdout), &deleted); err != nil {
		t.Fatalf("decode delete output %q: %v", stdout, err)
	}
	if deleted.IssueID != issue.ID || deleted.Type != "pull_request" || !deleted.Deleted {
		t.Fatalf("delete output = %+v", deleted)
	}
	_, stderr, code = runCLI(t, []string{"--api-url", server.URL, "artifact", "delete", stringID(issue.ID), "--type", "pull_request"})
	if code == 0 || !strings.Contains(stderr, "artifact not found") {
		t.Fatalf("redelete: code=%d stderr=%q", code, stderr)
	}
}

func TestIssueCreateWithAttachmentAgainstIssueTrackerAPI(t *testing.T) {
	ctx := context.Background()
	issueStore, err := store.OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer issueStore.Close()
	_, err = issueStore.CreateProject(ctx, entity.CreateProjectInput{Key: "ATTACH", Name: "Attach", Location: t.TempDir()})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	server := httptest.NewServer(api.NewServerWithAttachmentStorage(issueStore, store.NewAttachmentStorage(t.TempDir())).Handler())
	defer server.Close()

	attachmentPath := filepath.Join(t.TempDir(), "screenshot.png")
	if err := os.WriteFile(attachmentPath, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, 0o644); err != nil {
		t.Fatalf("write attachment: %v", err)
	}

	stdout, stderr, code := runCLI(t, []string{
		"--api-url", server.URL,
		"issue", "create",
		"--project", "ATTACH",
		"--title", "Issue with attachment",
		"--description", "Before image",
		"--attach", attachmentPath,
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Before image") || !strings.Contains(stdout, "attachment://att_") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}

	issues, err := issueStore.Issues(ctx)
	if err != nil {
		t.Fatalf("list issues: %v", err)
	}
	if len(issues) != 1 || !strings.Contains(issues[0].Description, "![screenshot.png](attachment://att_") {
		t.Fatalf("issues = %+v", issues)
	}
}

func TestIssueCreateWithDependenciesAgainstIssueTrackerAPI(t *testing.T) {
	ctx := context.Background()
	issueStore, err := store.OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer issueStore.Close()
	project, err := issueStore.CreateProject(ctx, entity.CreateProjectInput{Key: "DEPS", Name: "Deps", Location: t.TempDir()})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	first, err := issueStore.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "First dependency", Status: entity.StatusDone})
	if err != nil {
		t.Fatalf("create first dependency: %v", err)
	}
	second, err := issueStore.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Second dependency", Status: entity.StatusDone})
	if err != nil {
		t.Fatalf("create second dependency: %v", err)
	}
	server := httptest.NewServer(api.NewServer(issueStore).Handler())
	defer server.Close()

	stdout, stderr, code := runCLI(t, []string{
		"--api-url", server.URL,
		"issue", "create",
		"--project", "DEPS",
		"--title", "Issue with dependencies",
		"--dependency", fmt.Sprintf("%d,%d", first.ID, second.ID),
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	wantDependencyText := fmt.Sprintf("Dependencies: %d,%d", first.ID, second.ID)
	if !strings.Contains(stdout, "Issue with dependencies") || !strings.Contains(stdout, wantDependencyText) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}

	issues, err := issueStore.Issues(ctx)
	if err != nil {
		t.Fatalf("list issues: %v", err)
	}
	var created entity.Issue
	for _, issue := range issues {
		if issue.Title == "Issue with dependencies" {
			created = issue
			break
		}
	}
	if created.ID == 0 {
		t.Fatalf("created issue not found: %+v", issues)
	}
	assertInt64s(t, created.DependencyIDs, []int64{first.ID, second.ID})
}

func TestCommentAddWithAttachmentAgainstIssueTrackerAPI(t *testing.T) {
	ctx := context.Background()
	issueStore, err := store.OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer issueStore.Close()

	project, err := issueStore.CreateProject(ctx, entity.CreateProjectInput{Key: "COMATTACH", Name: "Comment Attach", Location: t.TempDir()})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	issue, err := issueStore.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Comment attachment"})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	server := httptest.NewServer(api.NewServerWithAttachmentStorage(issueStore, store.NewAttachmentStorage(t.TempDir())).Handler())
	defer server.Close()

	attachmentPath := filepath.Join(t.TempDir(), "comment.png")
	if err := os.WriteFile(attachmentPath, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}, 0o644); err != nil {
		t.Fatalf("write attachment: %v", err)
	}

	stdout, stderr, code := runCLI(t, []string{
		"--api-url", server.URL,
		"comment", "add", stringID(issue.ID),
		"--author", "codex",
		"--body", "See image",
		"--attach", attachmentPath,
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "See image") || !strings.Contains(stdout, "attachment://att_") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
	comments, err := issueStore.CommentsByIssueID(ctx, issue.ID, 0, 50)
	if err != nil {
		t.Fatalf("list comments: %v", err)
	}
	if len(comments) != 1 || !strings.Contains(comments[0].Body, "![comment.png](attachment://att_") {
		t.Fatalf("comments = %+v", comments)
	}
}

func TestProjectAddCheckRemoveAgainstIssueTrackerAPI(t *testing.T) {
	ctx := context.Background()
	issueStore, err := store.OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer issueStore.Close()
	server := httptest.NewServer(api.NewServer(issueStore).Handler())
	defer server.Close()

	projectRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(projectRoot, "AGENTS.md"), []byte("See docs/development.md for repository development instructions.\n"), 0o644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	stdout, stderr, code := runCLI(t, []string{
		"--api-url", server.URL,
		"project", "add",
		"--key", "demo-project",
		projectRoot,
	})
	if code != 0 {
		t.Fatalf("add code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Project: demo-project") {
		t.Fatalf("unexpected add stdout: %s", stdout)
	}
	assertFileContains(t, filepath.Join(projectRoot, "WORKFLOW.md"), "codex --sandbox workspace-write app-server")
	assertFileContains(t, filepath.Join(projectRoot, ".gitignore"), ".worktrees")

	stdout, stderr, code = runCLI(t, []string{"--api-url", server.URL, "project", "check", "demo-project"})
	if code != 0 {
		t.Fatalf("check code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	if !strings.Contains(stripANSI(stdout), "PASS\tapi.tq_usage") {
		t.Fatalf("unexpected check stdout: %s", stdout)
	}

	stdout, stderr, code = runCLI(t, []string{"--api-url", server.URL, "project", "list"})
	if code != 0 {
		t.Fatalf("list code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, ansiBold+"ID") ||
		!strings.Contains(stdout, ansiCyan+"demo-project"+ansiReset) ||
		!strings.Contains(stdout, projectRoot) {
		t.Fatalf("unexpected text list stdout: %s", stdout)
	}

	stdout, stderr, code = runCLI(t, []string{"--api-url", server.URL, "--output", "json", "project", "list"})
	if code != 0 {
		t.Fatalf("list code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, `"key": "demo-project"`) {
		t.Fatalf("unexpected list stdout: %s", stdout)
	}

	stdout, stderr, code = runCLI(t, []string{"--api-url", server.URL, "project", "remove", "-y", "demo-project"})
	if code != 0 {
		t.Fatalf("remove code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stripANSI(stdout), "Removed project demo-project") {
		t.Fatalf("unexpected remove stdout: %s", stdout)
	}
	projects, err := issueStore.Projects(ctx)
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if len(projects) != 0 {
		t.Fatalf("projects after remove = %+v", projects)
	}
}

func TestProjectRemoveRequiresMatchingKeyConfirmation(t *testing.T) {
	requests := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			writeTestJSON(t, w, apiResponse[[]entity.Project]{
				Data: []entity.Project{{ID: 7, Key: "demo-project", Name: "Demo Project"}},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/projects/7":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	stdout, stderr, code := runCLIWithStdin(t, []string{
		"--api-url", server.URL,
		"project", "remove", "demo-project",
	}, "demo-project\n")
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	plain := stripANSI(stdout)
	if !strings.Contains(plain, "WARNING: This operation cannot be undone.") ||
		!strings.Contains(plain, "issues, comments, attachments, workflow overrides, and run data") ||
		!strings.Contains(plain, "Removed project demo-project") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
	assertStringSlice(t, requests, []string{"GET /api/v1/projects", "DELETE /api/v1/projects/7"})
}

func TestProjectRemoveCancelsWhenConfirmationKeyDoesNotMatch(t *testing.T) {
	deleteCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			writeTestJSON(t, w, apiResponse[[]entity.Project]{
				Data: []entity.Project{{ID: 7, Key: "demo-project", Name: "Demo Project"}},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/projects/7":
			deleteCalled = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	stdout, stderr, code := runCLIWithStdin(t, []string{
		"--api-url", server.URL,
		"project", "remove", "demo-project",
	}, "wrong-project\n")
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr=%s", stderr)
	}
	if deleteCalled {
		t.Fatal("delete was called for mismatched confirmation")
	}
	if !strings.Contains(stdout, "Project removal cancelled") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestProjectRemoveYesSkipsConfirmation(t *testing.T) {
	requests := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			writeTestJSON(t, w, apiResponse[[]entity.Project]{
				Data: []entity.Project{{ID: 7, Key: "demo-project", Name: "Demo Project"}},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/projects/7":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	stdout, stderr, code := runCLIWithStdin(t, []string{
		"--api-url", server.URL,
		"project", "remove", "-y", "demo-project",
	}, "")
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if strings.Contains(stdout, "Type the project key to confirm") {
		t.Fatalf("confirmation prompt was shown: %s", stdout)
	}
	if !strings.Contains(stripANSI(stdout), "Removed project demo-project") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
	assertStringSlice(t, requests, []string{"GET /api/v1/projects", "DELETE /api/v1/projects/7"})
}

func TestProjectRemoveNonInteractiveCancelsWithoutDelete(t *testing.T) {
	deleteCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			writeTestJSON(t, w, apiResponse[[]entity.Project]{
				Data: []entity.Project{{ID: 7, Key: "demo-project", Name: "Demo Project"}},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/projects/7":
			deleteCalled = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	stdout, stderr, code := runCLIWithStdin(t, []string{
		"--api-url", server.URL,
		"project", "remove", "demo-project",
	}, "")
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if deleteCalled {
		t.Fatal("delete was called without confirmation input")
	}
	if !strings.Contains(stdout, "Project removal cancelled") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestProjectRemoveRunningRunErrorIncludesReason(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			writeTestJSON(t, w, apiResponse[[]entity.Project]{
				Data: []entity.Project{{ID: 7, Key: "demo-project", Name: "Demo Project"}},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/projects/7":
			w.WriteHeader(http.StatusConflict)
			writeTestJSON(t, w, apiErrorResponse{
				Error: struct {
					Code    string `json:"code"`
					Message string `json:"message"`
				}{Code: "projects.delete.running_runs", Message: "project has running runs: 1 running run(s)"},
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	stdout, stderr, code := runCLIWithStdin(t, []string{
		"--api-url", server.URL,
		"project", "remove", "-y", "demo-project",
	}, "")
	if code != 1 {
		t.Fatalf("code=%d stdout=%s stderr=%s", code, stdout, stderr)
	}
	message := decodeCLIError(t, stderr)
	if !strings.Contains(message, "project has running runs") || !strings.Contains(message, "1 running run(s)") {
		t.Fatalf("unexpected error: %s", message)
	}
}

func TestWorkflowRemoveDeletesProjectWorkflow(t *testing.T) {
	requests := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			writeTestJSON(t, w, apiResponse[[]entity.Project]{
				Data: []entity.Project{{ID: 7, Key: "demo-project", Name: "Demo Project"}},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/api/v1/projects/7/workflow":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	stdout, stderr, code := runCLI(t, []string{
		"--api-url", server.URL,
		"workflow", "remove",
		"--project", "demo-project",
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stripANSI(stdout), "Removed workflow override for project demo-project") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
	want := []string{"GET /api/v1/projects", "DELETE /api/v1/projects/7/workflow"}
	if strings.Join(requests, ",") != strings.Join(want, ",") {
		t.Fatalf("requests = %+v, want %+v", requests, want)
	}
}

func TestWorkflowAddFromBodyUpsertsProjectWorkflow(t *testing.T) {
	content := "---\ntracker:\n  kind: tasq\nagent:\n  max_turns: 3\n---\n# Prompt\nUse tq.\n"
	checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
	requests := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/projects":
			writeTestJSON(t, w, apiResponse[[]entity.Project]{
				Data: []entity.Project{{ID: 7, Key: "demo-project", Name: "Demo Project"}},
			})
		case r.Method == http.MethodPut && r.URL.Path == "/api/v1/projects/7/workflow":
			var input upsertProjectWorkflowInput
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				t.Fatal(err)
			}
			tracker, ok := input.Frontmatter["tracker"].(map[string]any)
			if !ok || tracker["kind"] != "tasq" {
				t.Fatalf("frontmatter = %#v", input.Frontmatter)
			}
			agent, ok := input.Frontmatter["agent"].(map[string]any)
			if !ok || agent["max_turns"] != float64(3) {
				t.Fatalf("frontmatter = %#v", input.Frontmatter)
			}
			if input.Body != "# Prompt\nUse tq.\n" || input.Checksum != checksum {
				t.Fatalf("input = %+v, checksum want %s", input, checksum)
			}
			writeTestJSON(t, w, apiResponse[entity.ProjectWorkflow]{
				Data: entity.ProjectWorkflow{
					ProjectID:   7,
					Frontmatter: input.Frontmatter,
					Body:        input.Body,
					Checksum:    input.Checksum,
				},
			})
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	stdout, stderr, code := runCLI(t, []string{
		"--api-url", server.URL,
		"workflow", "add",
		"--project", "demo-project",
		"--body", content,
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Workflow override updated for project demo-project") || !strings.Contains(stdout, checksum) {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
	want := []string{"GET /api/v1/projects", "PUT /api/v1/projects/7/workflow"}
	if strings.Join(requests, ",") != strings.Join(want, ",") {
		t.Fatalf("requests = %+v, want %+v", requests, want)
	}
}

func TestWorkflowAddFileAndBodyOverwriteAgainstIssueTrackerAPI(t *testing.T) {
	ctx := context.Background()
	issueStore, err := store.OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer issueStore.Close()
	project, err := issueStore.CreateProject(ctx, entity.CreateProjectInput{Key: "demo-project", Name: "Demo Project", Location: t.TempDir()})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	server := httptest.NewServer(api.NewServer(issueStore).Handler())
	defer server.Close()

	fileContent := "---\ntracker:\n  kind: file\n---\nFile prompt.\n"
	filePath := filepath.Join(t.TempDir(), "WORKFLOW.md")
	if err := os.WriteFile(filePath, []byte(fileContent), 0o644); err != nil {
		t.Fatalf("write workflow file: %v", err)
	}
	stdout, stderr, code := runCLI(t, []string{
		"--api-url", server.URL,
		"workflow", "add",
		"--project", "demo-project",
		"--file", filePath,
	})
	if code != 0 {
		t.Fatalf("file add code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Workflow override updated for project demo-project") {
		t.Fatalf("unexpected file add stdout: %s", stdout)
	}

	bodyContent := "---\ntracker:\n  kind: body\n---\nBody prompt.\n"
	stdout, stderr, code = runCLI(t, []string{
		"--api-url", server.URL,
		"workflow", "add",
		"--project", "demo-project",
		"--body", bodyContent,
	})
	if code != 0 {
		t.Fatalf("body add code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, fmt.Sprintf("%x", sha256.Sum256([]byte(bodyContent)))) {
		t.Fatalf("unexpected body add stdout: %s", stdout)
	}

	stored, err := issueStore.ProjectWorkflow(ctx, project.ID)
	if err != nil {
		t.Fatalf("read stored workflow: %v", err)
	}
	tracker, ok := stored.Frontmatter["tracker"].(map[string]any)
	if !ok || tracker["kind"] != "body" || stored.Body != "Body prompt.\n" {
		t.Fatalf("stored workflow = %+v", stored)
	}
	if stored.Checksum != fmt.Sprintf("%x", sha256.Sum256([]byte(bodyContent))) {
		t.Fatalf("checksum = %s", stored.Checksum)
	}
}

func TestWorkflowAddRequiresExactlyOneSource(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "missing source", args: []string{"--api-url", defaultAPIURL, "workflow", "add", "--project", "demo-project"}},
		{name: "multiple sources", args: []string{"--api-url", defaultAPIURL, "workflow", "add", "--project", "demo-project", "--file", "WORKFLOW.md", "--body", "---\n---\n"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stdout, stderr, code := runCLI(t, test.args)
			if code != 2 {
				t.Fatalf("code=%d stderr=%s", code, stderr)
			}
			if stdout != "" {
				t.Fatalf("expected empty stdout: %s", stdout)
			}
			if got := decodeCLIError(t, stderr); got != "exactly one of file or body is required" {
				t.Fatalf("error=%q", got)
			}
		})
	}
}

func TestWorkflowRemoveRequiresProject(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{
		"--api-url", defaultAPIURL,
		"workflow", "remove",
	})
	if code != 2 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout: %s", stdout)
	}
	if got := decodeCLIError(t, stderr); got != "project is required" {
		t.Fatalf("error=%q", got)
	}
}

func TestWorkflowShowResolvesPhysicalFile(t *testing.T) {
	ctx := context.Background()
	issueStore, err := store.OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer issueStore.Close()
	server := httptest.NewServer(api.NewServer(issueStore).Handler())
	defer server.Close()

	projectRoot := t.TempDir()
	workflowPath := filepath.Join(projectRoot, "WORKFLOW.md")
	if err := os.WriteFile(workflowPath, []byte("# File Workflow\n\nRun from file.\n"), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}
	project, err := issueStore.CreateProject(ctx, entity.CreateProjectInput{Key: "file-project", Name: "File Project", Location: projectRoot})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := issueStore.UpsertProjectWorkflow(ctx, entity.UpsertProjectWorkflowInput{
		ProjectID:       project.ID,
		FrontmatterJSON: `{}`,
		Body:            "DB workflow should not win.",
		Checksum:        strings.Repeat("a", 64),
	}); err != nil {
		t.Fatalf("upsert workflow: %v", err)
	}

	stdout, stderr, code := runCLI(t, []string{"--api-url", server.URL, "workflow", "show", "--project", "file-project"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "# Source: file ("+workflowPath+")") || !strings.Contains(stdout, "Run from file.") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
	if strings.Contains(stdout, "DB workflow should not win.") {
		t.Fatalf("file source did not take priority: %s", stdout)
	}
}

func TestWorkflowShowResolvesDBWorkflow(t *testing.T) {
	ctx := context.Background()
	issueStore, err := store.OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer issueStore.Close()
	server := httptest.NewServer(api.NewServer(issueStore).Handler())
	defer server.Close()

	project, err := issueStore.CreateProject(ctx, entity.CreateProjectInput{Key: "db-project", Name: "DB Project", Location: t.TempDir()})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := issueStore.UpsertProjectWorkflow(ctx, entity.UpsertProjectWorkflowInput{
		ProjectID:       project.ID,
		FrontmatterJSON: `{"agent":{"max_turns":3}}`,
		Body:            "# DB Workflow\n\nRun from DB.",
		Checksum:        strings.Repeat("b", 64),
	}); err != nil {
		t.Fatalf("upsert workflow: %v", err)
	}

	stdout, stderr, code := runCLI(t, []string{"--api-url", server.URL, "workflow", "show", "--project", "db-project"})
	if code != 0 {
		t.Fatalf("text code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, fmt.Sprintf("# Source: db (project db-project, id %d)", project.ID)) ||
		!strings.Contains(stdout, "---\nagent:\n    max_turns: 3\n---\n# DB Workflow") {
		t.Fatalf("unexpected text stdout: %s", stdout)
	}

	stdout, stderr, code = runCLI(t, []string{"--api-url", server.URL, "workflow", "show", "--project", "db-project", "--json"})
	if code != 0 {
		t.Fatalf("json code=%d stderr=%s", code, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr=%s", stderr)
	}
	var result workflowShowResult
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("decode stdout: %v\n%s", err, stdout)
	}
	if result.Source.Type != "db" || result.Source.ProjectID != project.ID {
		t.Fatalf("source = %+v", result.Source)
	}
	if !strings.Contains(result.Content, "agent:\n    max_turns: 3") || !strings.Contains(result.Content, "Run from DB.") {
		t.Fatalf("content = %q", result.Content)
	}
}

func TestWorkflowShowResolvesGlobalWorkflow(t *testing.T) {
	ctx := context.Background()
	issueStore, err := store.OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer issueStore.Close()
	server := httptest.NewServer(api.NewServer(issueStore).Handler())
	defer server.Close()

	if _, err := issueStore.CreateProject(ctx, entity.CreateProjectInput{Key: "global-project", Name: "Global Project", Location: t.TempDir()}); err != nil {
		t.Fatalf("create project: %v", err)
	}
	home := t.TempDir()
	t.Setenv(tqconfig.EnvHome, home)
	globalPath := tqconfig.WorkflowPath(home)
	if err := os.WriteFile(globalPath, []byte("# Global Workflow\n\nRun globally.\n"), 0o644); err != nil {
		t.Fatalf("write global workflow: %v", err)
	}

	stdout, stderr, code := runCLI(t, []string{"--api-url", server.URL, "workflow", "show", "--project", "global-project"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "# Source: global ("+globalPath+")") || !strings.Contains(stdout, "Run globally.") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestIssueGetAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		writeTestJSON(t, w, apiErrorResponse{
			Error: struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			}{Code: "issues.get.not_found", Message: "issue not found"},
		})
	}))
	defer server.Close()

	stdout, stderr, code := runCLI(t, []string{"issue", "get", "99", "--api-url", server.URL})
	if code == 0 {
		t.Fatalf("expected non-zero code")
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout: %s", stdout)
	}
	if strings.TrimSpace(stderr) != "Error: issue not found" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestCreateRequiresTitle(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{"--api-url", defaultAPIURL, "issue", "create"})
	if code != 2 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout: %s", stdout)
	}
	if strings.TrimSpace(stderr) != "Error: title is required" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestFlagParseErrorWritesOnlyJSON(t *testing.T) {
	stdout, stderr, code := runCLI(t, []string{"--api-url", defaultAPIURL, "issue", "create", "--unknown"})
	if code != 2 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if stdout != "" {
		t.Fatalf("expected empty stdout: %s", stdout)
	}
	if strings.TrimSpace(stderr) != "Error: flag provided but not defined: -unknown" {
		t.Fatalf("unexpected stderr: %s", stderr)
	}
}

func TestAPIURLDefaultsToEnvironment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeTestJSON(t, w, apiResponse[[]entity.Issue]{Data: []entity.Issue{}})
	}))
	defer server.Close()

	t.Setenv("TQ_API_URL", server.URL)
	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), []string{"issue", "list"}, &out, &errOut)
	stdout := out.String()
	stderr := errOut.String()
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if stdout != "" {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestAPIURLDefaultsToState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/issues" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		writeTestJSON(t, w, apiResponse[[]entity.Issue]{Data: []entity.Issue{}})
	}))
	defer server.Close()
	t.Setenv("TQ_API_URL", "")
	t.Setenv(tqconfig.EnvHome, t.TempDir())
	if err := tqconfig.UpdateState(func(state *tqconfig.State) error {
		state.IssueTracker = &tqconfig.ServiceState{Addr: server.URL}
		return nil
	}); err != nil {
		t.Fatalf("write state: %v", err)
	}

	stdout, stderr, code := runCLI(t, []string{"issue", "list"})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if stdout != "" {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func TestCommentAddDefaultsAuthorFromConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv(tqconfig.EnvHome, home)
	t.Setenv("TQ_AUTHOR", "")
	if err := os.MkdirAll(tqconfig.ConfigDir(home), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tqconfig.ConfigPath(home), []byte("author: config-author\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var input entity.CreateCommentInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if input.Author != "config-author" {
			t.Fatalf("author=%q", input.Author)
		}
		w.WriteHeader(http.StatusCreated)
		writeTestJSON(t, w, apiResponse[entity.Comment]{Data: entity.Comment{
			ID:      8,
			IssueID: 42,
			Author:  input.Author,
			Type:    entity.CommentGeneral,
			Body:    input.Body,
		}})
	}))
	defer server.Close()

	stdout, stderr, code := runCLI(t, []string{
		"--api-url", server.URL,
		"comment", "add", "42",
		"--body", "Progress",
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Author: config-author") {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func stringID(id int64) string {
	return strconv.FormatInt(id, 10)
}

type issueShortcutTest struct {
	args        []string
	id          int64
	wantPatch   map[string]string
	response    entity.Issue
	wantMessage string
}

func assertIssueShortcut(t *testing.T, test issueShortcutTest) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/v1/issues/"+stringID(test.id) {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var input map[string]string
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if !stringMapEqual(input, test.wantPatch) {
			t.Fatalf("unexpected input: %+v", input)
		}
		writeTestJSON(t, w, apiResponse[entity.Issue]{Data: test.response})
	}))
	defer server.Close()

	args := append([]string{"--api-url", server.URL}, test.args...)
	stdout, stderr, code := runCLI(t, args)
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if test.wantMessage != "" && !strings.Contains(stdout, test.wantMessage) {
		t.Fatalf("stdout does not contain %q: %s", test.wantMessage, stdout)
	}
	return stdout
}

func stringMapEqual(left map[string]string, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, leftValue := range left {
		if right[key] != leftValue {
			return false
		}
	}
	return true
}

func assertInt64s(t *testing.T, got []int64, want []int64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("ids=%v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("ids=%v, want %v", got, want)
		}
	}
}

func decodeCLIError(t *testing.T, stderr string) string {
	t.Helper()
	stderr = strings.TrimSpace(stripANSI(stderr))
	if strings.HasPrefix(stderr, "Error: ") {
		return strings.TrimPrefix(stderr, "Error: ")
	}
	var payload struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal([]byte(stderr), &payload); err != nil {
		t.Fatalf("decode stderr: %v: %s", err, stderr)
	}
	return payload.Error
}

func runCLI(t *testing.T, args []string) (string, string, int) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run(context.Background(), args, &stdout, &stderr)
	return stdout.String(), stripANSI(stderr.String()), code
}

func runCLIWithStdin(t *testing.T, args []string, stdin string) (string, string, int) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run(context.Background(), args, strings.NewReader(stdin), &stdout, &stderr)
	return stdout.String(), stripANSI(stderr.String()), code
}

func stripANSI(value string) string {
	for _, code := range []string{ansiBold, ansiGreen, ansiYellow, ansiRed, ansiMagenta, ansiCyan, ansiFaint, ansiReset} {
		value = strings.ReplaceAll(value, code, "")
	}
	return value
}

func runCLIWithUpdateRunner(t *testing.T, args []string, stdin string, runner *fakeUpdateRunner) (string, string, int) {
	t.Helper()
	original := newUpdateRunner
	newUpdateRunner = func(app, config) updateRunner {
		return runner
	}
	t.Cleanup(func() {
		newUpdateRunner = original
	})

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run(context.Background(), args, strings.NewReader(stdin), &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

type fakeUpdateRunner struct {
	calls        []string
	current      string
	target       string
	installed    string
	confirmOK    bool
	installErr   error
	migrateErr   error
	startErr     error
	installedErr error
}

func (r *fakeUpdateRunner) currentVersion(context.Context) (string, error) {
	r.calls = append(r.calls, "current")
	return r.current, nil
}

func (r *fakeUpdateRunner) targetVersion(_ context.Context, tag string) (string, error) {
	r.calls = append(r.calls, "target:"+tag)
	return r.target, nil
}

func (r *fakeUpdateRunner) confirm(context.Context) (bool, error) {
	r.calls = append(r.calls, "confirm")
	return r.confirmOK, nil
}

func (r *fakeUpdateRunner) stopServices(context.Context) error {
	r.calls = append(r.calls, "stop")
	return nil
}

func (r *fakeUpdateRunner) install(_ context.Context, tag string) error {
	r.calls = append(r.calls, "install:"+tag)
	return r.installErr
}

func (r *fakeUpdateRunner) installedVersion(context.Context) (string, error) {
	r.calls = append(r.calls, "installed")
	return r.installed, r.installedErr
}

func (r *fakeUpdateRunner) migrate(context.Context) error {
	r.calls = append(r.calls, "migrate")
	return r.migrateErr
}

func (r *fakeUpdateRunner) startServices(context.Context) error {
	r.calls = append(r.calls, "start")
	return r.startErr
}

func assertStringSlice(t *testing.T, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got calls=%v, want %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("got calls=%v, want %v", got, want)
		}
	}
}

func writeTestJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatal(err)
	}
}

func assertFileContains(t *testing.T, path string, want string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(content), want) {
		t.Fatalf("%s does not contain %q: %s", path, want, string(content))
	}
}
