package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

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
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/issues" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var input entity.CreateIssueInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatal(err)
		}
		if input.Title != "Add CLI" || input.Status != entity.StatusReady || input.Priority != entity.PriorityHigh {
			t.Fatalf("unexpected input: %+v", input)
		}
		w.WriteHeader(http.StatusCreated)
		writeTestJSON(t, w, apiResponse[entity.Issue]{Data: entity.Issue{
			ID:        10,
			Title:     input.Title,
			Status:    input.Status,
			Priority:  input.Priority,
			CreatedAt: time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC),
			UpdatedAt: time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC),
		}})
	}))
	defer server.Close()

	stdout, stderr, code := runCLI(t, []string{
		"issue", "create",
		"--api-url", server.URL,
		"--title", "Add CLI",
		"--status", "ready",
		"--priority", "high",
	})
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "ID: 10") || !strings.Contains(stdout, "Status: ready") {
		t.Fatalf("unexpected stdout: %s", stdout)
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
	if !strings.Contains(stdout, "Status: blocked") {
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
	code := run(context.Background(), []string{"issue", "list"}, &out, &errOut)
	stdout := out.String()
	stderr := errOut.String()
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, stderr)
	}
	if stdout != "" {
		t.Fatalf("unexpected stdout: %s", stdout)
	}
}

func runCLI(t *testing.T, args []string) (string, string, int) {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run(context.Background(), args, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

func writeTestJSON(t *testing.T, w http.ResponseWriter, value any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatal(err)
	}
}
