package tracker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func TestIssuesByStatesFiltersIssues(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/issues" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if got := r.URL.Query().Get("states"); got != "ready,in_progress" {
			t.Fatalf("states query = %q", got)
		}
		writeTestData(t, w, []entity.Issue{
			{ID: 1, Status: entity.StatusReady},
			{ID: 3, Status: entity.StatusInProgress},
		})
	}))
	defer server.Close()

	issues, err := NewClient(server.URL).IssuesByStates(context.Background(), []string{string(entity.StatusReady), string(entity.StatusInProgress)})
	if err != nil {
		t.Fatalf("issues by states: %v", err)
	}
	if len(issues) != 2 || issues[0].ID != 1 || issues[1].ID != 3 {
		t.Fatalf("issues = %+v", issues)
	}
}

func TestIssuesByStatesSkipsRequestForEmptyStates(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	issues, err := NewClient(server.URL).IssuesByStates(context.Background(), nil)
	if err != nil {
		t.Fatalf("issues by states: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("issues = %+v", issues)
	}
}

func TestQueueGetsIssueQueue(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/queue" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		writeTestData(t, w, entity.Queue{
			Queued: []entity.QueueIssue{
				{Issue: entity.Issue{ID: 1, Status: entity.StatusReady}},
			},
			Pending: []entity.QueueIssue{
				{Issue: entity.Issue{ID: 2, Status: entity.StatusReady}, BlockedDependencyIDs: []int64{1}},
			},
		})
	}))
	defer server.Close()

	queue, err := NewClient(server.URL).Queue(context.Background())
	if err != nil {
		t.Fatalf("queue: %v", err)
	}
	if len(queue.Queued) != 1 || queue.Queued[0].ID != 1 {
		t.Fatalf("queued = %+v", queue.Queued)
	}
	if len(queue.Pending) != 1 || queue.Pending[0].ID != 2 || len(queue.Pending[0].BlockedDependencyIDs) != 1 || queue.Pending[0].BlockedDependencyIDs[0] != 1 {
		t.Fatalf("pending = %+v", queue.Pending)
	}
}

func TestIssueStatesByIDsPostsIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/issues/states" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var payload struct {
			IDs []int64 `json:"ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if len(payload.IDs) != 2 || payload.IDs[0] != 1 || payload.IDs[1] != 2 {
			t.Fatalf("ids = %+v", payload.IDs)
		}
		writeTestData(t, w, []IssueState{
			{ID: 1, Status: entity.StatusReady},
			{ID: 2, Status: entity.StatusDone},
		})
	}))
	defer server.Close()

	states, err := NewClient(server.URL).IssueStatesByIDs(context.Background(), []int64{1, 2})
	if err != nil {
		t.Fatalf("issue states by ids: %v", err)
	}
	if len(states) != 2 {
		t.Fatalf("states = %+v", states)
	}
	if states[0].ID != 1 || states[0].Status != entity.StatusReady || states[1].ID != 2 || states[1].Status != entity.StatusDone {
		t.Fatalf("states = %+v", states)
	}
}

func TestIssueStatesByIDsSkipsRequestForEmptyIDs(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	states, err := NewClient(server.URL).IssueStatesByIDs(context.Background(), nil)
	if err != nil {
		t.Fatalf("issue states by ids: %v", err)
	}
	if len(states) != 0 {
		t.Fatalf("states = %+v", states)
	}
}

func TestUpdateIssuePatchesIssue(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/v1/issues/42" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var payload entity.UpdateIssueInput
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload.Status == nil || *payload.Status != entity.StatusBlocked {
			t.Fatalf("status = %+v", payload.Status)
		}
		writeTestData(t, w, entity.Issue{ID: 42, Status: entity.StatusBlocked})
	}))
	defer server.Close()

	blocked := entity.StatusBlocked
	issue, err := NewClient(server.URL).UpdateIssue(context.Background(), 42, entity.UpdateIssueInput{Status: &blocked})
	if err != nil {
		t.Fatalf("update issue: %v", err)
	}
	if issue.ID != 42 || issue.Status != entity.StatusBlocked {
		t.Fatalf("issue = %+v", issue)
	}
}

func TestWorkflowGetsProjectWorkflow(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/projects/42/workflow" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		writeTestData(t, w, entity.ProjectWorkflow{
			ProjectID:   42,
			Frontmatter: map[string]any{"agent": map[string]any{"max_turns": float64(3)}},
			Body:        "Work on {{ issue.id }}.",
			Checksum:    "sha256:abc123",
		})
	}))
	defer server.Close()

	workflow, found, err := NewClient(server.URL).Workflow(context.Background(), 42)
	if err != nil {
		t.Fatalf("workflow: %v", err)
	}
	if !found {
		t.Fatal("found = false")
	}
	if workflow.Body != "Work on {{ issue.id }}." || workflow.Checksum != "sha256:abc123" {
		t.Fatalf("workflow = %+v", workflow)
	}
	agent, ok := workflow.Frontmatter["agent"].(map[string]any)
	if !ok || agent["max_turns"] != float64(3) {
		t.Fatalf("workflow = %+v", workflow)
	}
}

func TestProjectsListsProjects(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/projects" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		writeTestData(t, w, []entity.Project{
			{ID: 1, Key: "APP", Location: "/repo/app"},
			{ID: 2, Key: "CLI", Location: "/repo/cli"},
		})
	}))
	defer server.Close()

	projects, err := NewClient(server.URL).Projects(context.Background())
	if err != nil {
		t.Fatalf("projects: %v", err)
	}
	if len(projects) != 2 || projects[0].Key != "APP" || projects[1].Location != "/repo/cli" {
		t.Fatalf("projects = %+v", projects)
	}
}

func TestWorkflowReturnsNotFoundIndicator(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/projects/42/workflow" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{
				"code":    "projects.workflow_not_found",
				"message": "workflow not found",
			},
			"meta": map[string]any{},
		}); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer server.Close()

	workflow, found, err := NewClient(server.URL).Workflow(context.Background(), 42)
	if err != nil {
		t.Fatalf("workflow: %v", err)
	}
	if found {
		t.Fatal("found = true")
	}
	if workflow.Frontmatter != nil || workflow.Body != "" || workflow.Checksum != "" {
		t.Fatalf("workflow = %+v", workflow)
	}
}

func TestUpsertWorkflowPutsProjectWorkflow(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/api/v1/projects/7/workflow" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var payload struct {
			Frontmatter json.RawMessage `json:"frontmatter"`
			Body        string          `json:"body"`
			Checksum    string          `json:"checksum"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if string(payload.Frontmatter) != `{"agent":{"max_turns":3}}` || payload.Body != "Project prompt" || payload.Checksum != "abc" {
			t.Fatalf("payload = %+v", payload)
		}
		writeTestData(t, w, entity.ProjectWorkflow{
			ProjectID:   7,
			Frontmatter: map[string]any{"agent": map[string]any{"max_turns": float64(3)}},
			Body:        payload.Body,
			Checksum:    payload.Checksum,
		})
	}))
	defer server.Close()

	workflow, err := NewClient(server.URL).UpsertWorkflow(context.Background(), 7, entity.UpsertProjectWorkflowInput{
		ProjectID:       7,
		FrontmatterJSON: `{"agent":{"max_turns":3}}`,
		Body:            "Project prompt",
		Checksum:        "abc",
	})
	if err != nil {
		t.Fatalf("upsert workflow: %v", err)
	}
	if workflow.Body != "Project prompt" || workflow.Checksum != "abc" {
		t.Fatalf("workflow = %+v", workflow)
	}
}

func TestCreateCommentPostsComment(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/issues/42/comments" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var payload entity.CreateCommentInput
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload.Author != "tasq-orchestrator" || payload.Type != entity.CommentBlocker || payload.Body != "runner failed" {
			t.Fatalf("payload = %+v", payload)
		}
		writeTestData(t, w, entity.Comment{ID: 7, IssueID: 42, Author: payload.Author, Type: payload.Type, Body: payload.Body})
	}))
	defer server.Close()

	comment, err := NewClient(server.URL).CreateComment(context.Background(), 42, entity.CreateCommentInput{
		Author: "tasq-orchestrator",
		Type:   entity.CommentBlocker,
		Body:   "runner failed",
	})
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	if comment.ID != 7 || comment.IssueID != 42 || comment.Type != entity.CommentBlocker {
		t.Fatalf("comment = %+v", comment)
	}
}

func writeTestData(t *testing.T, w http.ResponseWriter, data any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{
		"data": data,
		"meta": map[string]any{},
	}); err != nil {
		t.Fatalf("write response: %v", err)
	}
}
