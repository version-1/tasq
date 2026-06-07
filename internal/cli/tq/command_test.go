package tq

import (
	"bytes"
	"context"
	"encoding/json"
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
	if !strings.Contains(stdout, "issue-tracker\tstopped") {
		t.Fatalf("stdout missing issue-tracker stopped status: %s", stdout)
	}
	if !strings.Contains(stdout, "orchestrator\tstopped") {
		t.Fatalf("stdout missing orchestrator stopped status: %s", stdout)
	}
	if !strings.Contains(stdout, "web\tstopped") {
		t.Fatalf("stdout missing web stopped status: %s", stdout)
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
	if !strings.Contains(stdout, "issue-tracker\tstopped") {
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
	issueStore, err := store.Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
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

func TestIssueCreateWithAttachmentAgainstIssueTrackerAPI(t *testing.T) {
	ctx := context.Background()
	issueStore, err := store.Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
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

func TestCommentAddWithAttachmentAgainstIssueTrackerAPI(t *testing.T) {
	ctx := context.Background()
	issueStore, err := store.Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
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
	issueStore, err := store.Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
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
	assertFileContains(t, filepath.Join(projectRoot, "WORKFLOW.md"), "codex app-server")
	assertFileContains(t, filepath.Join(projectRoot, ".gitignore"), ".worktrees")

	stdout, stderr, code = runCLI(t, []string{"--api-url", server.URL, "project", "check", "demo-project"})
	if code != 0 {
		t.Fatalf("check code=%d stderr=%s stdout=%s", code, stderr, stdout)
	}
	if !strings.Contains(stdout, "PASS\tapi.tq_usage") {
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

	stdout, stderr, code = runCLI(t, []string{"--api-url", server.URL, "project", "remove", "demo-project"})
	if code != 0 {
		t.Fatalf("remove code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Removed project demo-project") {
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
	if strings.TrimSpace(stderr) != `{"error":"issue not found"}` {
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
	if strings.TrimSpace(stderr) != `{"error":"title is required"}` {
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
	if strings.TrimSpace(stderr) != `{"error":"flag provided but not defined: -unknown"}` {
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

func decodeCLIError(t *testing.T, stderr string) string {
	t.Helper()
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
	return stdout.String(), stderr.String(), code
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
