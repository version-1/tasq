package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/issue/store"
)

func TestSuccessResponseEnvelope(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var payload struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
		Meta map[string]any `json:"meta"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Status != "ok" {
		t.Fatalf("data.status = %q", payload.Data.Status)
	}
	if payload.Meta == nil || len(payload.Meta) != 0 {
		t.Fatalf("meta = %#v", payload.Meta)
	}
}

func TestErrorResponseEnvelope(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/invalid", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		Meta map[string]any `json:"meta"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Error.Code != "projects.get.invalid_id" {
		t.Fatalf("error.code = %q", payload.Error.Code)
	}
	if payload.Error.Message != "project id is invalid" {
		t.Fatalf("error.message = %q", payload.Error.Message)
	}
	if payload.Meta == nil || len(payload.Meta) != 0 {
		t.Fatalf("meta = %#v", payload.Meta)
	}
}

func TestWorkspaceRoutesAreNotRegistered(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/workspaces", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestNoContentResponseStaysEmpty(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	project, err := server.store.CreateProject(context.Background(), entity.CreateProjectInput{
		Key:      "DOCS",
		Name:     "Docs",
		Location: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/"+stringID(project.ID), nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestDeleteProjectWorkflowAllowsMissingOverride(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	project, err := server.store.CreateProject(context.Background(), entity.CreateProjectInput{
		Key:      "DOCS",
		Name:     "Docs",
		Location: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/"+stringID(project.ID)+"/workflow", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestProjectCheckPassesWithDefaultTaskWorkPrompt(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	project, err := server.store.CreateProject(context.Background(), entity.CreateProjectInput{
		Key:      "project-api",
		Name:     "Project API",
		Location: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	body := bytes.NewBufferString("---\nagent:\n  max_turns: 3\n---\nWork on {{ issue.id }}.")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+stringID(project.ID)+"/check", body)
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Data struct {
			Passed bool   `json:"passed"`
			Reason string `json:"reason"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.Data.Passed {
		t.Fatalf("expected check to pass: %+v", payload.Data)
	}
}

func TestDeleteProjectWorkflow(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	project, err := server.store.CreateProject(context.Background(), entity.CreateProjectInput{
		Key:      "WORKFLOW",
		Name:     "Workflow",
		Location: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := server.store.UpsertProjectWorkflow(context.Background(), entity.UpsertProjectWorkflowInput{
		ProjectID:       project.ID,
		FrontmatterJSON: `{"tracker":{"kind":"tasq"}}`,
		Body:            "Use tq to keep the issue tracker synchronized.",
		Checksum:        strings.Repeat("a", 64),
	}); err != nil {
		t.Fatalf("upsert project workflow: %v", err)
	}
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/"+stringID(project.ID)+"/workflow", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body = %q", rec.Body.String())
	}
	if _, err := server.store.ProjectWorkflow(context.Background(), project.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("project workflow err = %v, want sql.ErrNoRows", err)
	}
}

func TestUpsertProjectWorkflow(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	project, err := server.store.CreateProject(context.Background(), entity.CreateProjectInput{
		Key:      "WFUPSERT",
		Name:     "Workflow Upsert",
		Location: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	body := bytes.NewBufferString(`{
		"frontmatter": {"tracker": {"kind": "tasq"}},
		"body": "Use tq to keep the issue tracker synchronized.",
		"checksum": "` + strings.Repeat("a", 64) + `"
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/"+stringID(project.ID)+"/workflow", body)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Data struct {
			ProjectID   int64 `json:"projectId"`
			Frontmatter struct {
				Tracker struct {
					Kind string `json:"kind"`
				} `json:"tracker"`
			} `json:"frontmatter"`
			Body     string `json:"body"`
			Checksum string `json:"checksum"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.ProjectID != project.ID || payload.Data.Frontmatter.Tracker.Kind != "tasq" || payload.Data.Body == "" || payload.Data.Checksum != strings.Repeat("a", 64) {
		t.Fatalf("data = %+v", payload.Data)
	}
	stored, err := server.store.ProjectWorkflow(context.Background(), project.ID)
	if err != nil {
		t.Fatalf("read project workflow: %v", err)
	}
	tracker, ok := stored.Frontmatter["tracker"].(map[string]any)
	if !ok || tracker["kind"] != "tasq" {
		t.Fatalf("frontmatter = %#v", stored.Frontmatter)
	}
}

func TestUpsertProjectWorkflowSkipsMatchingChecksum(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	project, err := server.store.CreateProject(context.Background(), entity.CreateProjectInput{
		Key:      "WFSKIP",
		Name:     "Workflow Skip",
		Location: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	checksum := strings.Repeat("b", 64)
	if _, err := server.store.UpsertProjectWorkflow(context.Background(), entity.UpsertProjectWorkflowInput{
		ProjectID:       project.ID,
		FrontmatterJSON: `{"tracker":{"kind":"tasq"}}`,
		Body:            "Original prompt template.",
		Checksum:        checksum,
	}); err != nil {
		t.Fatalf("upsert project workflow: %v", err)
	}
	body := bytes.NewBufferString(`{
		"frontmatter": {"tracker": {"kind": "changed"}},
		"body": "Changed prompt template.",
		"checksum": "` + checksum + `"
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/"+stringID(project.ID)+"/workflow", body)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Data struct {
			Frontmatter struct {
				Tracker struct {
					Kind string `json:"kind"`
				} `json:"tracker"`
			} `json:"frontmatter"`
			Body string `json:"body"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Frontmatter.Tracker.Kind != "tasq" || payload.Data.Body != "Original prompt template." {
		t.Fatalf("data = %+v", payload.Data)
	}
	stored, err := server.store.ProjectWorkflow(context.Background(), project.ID)
	if err != nil {
		t.Fatalf("read project workflow: %v", err)
	}
	tracker, ok := stored.Frontmatter["tracker"].(map[string]any)
	if !ok || tracker["kind"] != "tasq" || stored.Body != "Original prompt template." {
		t.Fatalf("stored workflow = %+v", stored)
	}
}

func TestDeleteProjectWorkflowReturnsNotFoundForMissingProject(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/999999/workflow", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
	assertErrorCode(t, rec, "projects.workflow.delete.not_found")
}

func TestProjectCheckFailsWhenTaskWorkPromptDisabledWithoutTQUsage(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	project, err := server.store.CreateProject(context.Background(), entity.CreateProjectInput{
		Key:      "project-api",
		Name:     "Project API",
		Location: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	body := bytes.NewBufferString("---\ntasq:\n  task_work_prompt: false\n---\nWork on {{ issue.id }}.")
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/"+stringID(project.ID)+"/check", body)
	req.Header.Set("Content-Type", "text/plain")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Data struct {
			Passed bool   `json:"passed"`
			Reason string `json:"reason"`
		} `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Passed {
		t.Fatalf("expected check to fail: %+v", payload.Data)
	}
}

func TestProjectWorkflowReturnsStoredWorkflow(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	project := createProject(t, server, "WORKFLOW")
	_, err := server.store.UpsertProjectWorkflow(context.Background(), entity.UpsertProjectWorkflowInput{
		ProjectID:       project.ID,
		FrontmatterJSON: `{"agent":{"max_turns":3},"tasq":{"task_work_prompt":false}}`,
		Body:            "Work on {{ issue.id }}.",
		Checksum:        strings.Repeat("a", 64),
	})
	if err != nil {
		t.Fatalf("upsert workflow: %v", err)
	}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+stringID(project.ID)+"/workflow", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	workflow := decodeData[entity.ProjectWorkflow](t, rec)
	if workflow.ProjectID != project.ID || workflow.Body != "Work on {{ issue.id }}." || workflow.Checksum != strings.Repeat("a", 64) {
		t.Fatalf("workflow = %+v", workflow)
	}
	agent, ok := workflow.Frontmatter["agent"].(map[string]any)
	if !ok || agent["max_turns"] != float64(3) {
		t.Fatalf("frontmatter = %#v", workflow.Frontmatter)
	}
}

func TestProjectWorkflowReturnsNotFoundWhenMissing(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	project := createProject(t, server, "NOWORKFLOW")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+stringID(project.ID)+"/workflow", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec, "projects.workflow.not_found")
}

func TestOptionsResponseStaysEmpty(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/projects", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestSummaryIncludesIssueStats(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	withComments := createIssue(t, server, "Issue with comments", entity.StatusReady)
	withoutComments := createIssue(t, server, "Issue without comments", entity.StatusReady)
	createComment(t, server, withComments.ID, "first comment")
	createComment(t, server, withComments.ID, "second comment")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/summary", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	summary := decodeData[entity.Summary](t, rec)
	statsByIssueID := map[int64]entity.IssueStats{}
	for _, column := range summary.Columns {
		for _, issue := range column.Issues {
			statsByIssueID[issue.ID] = issue.Stats
		}
	}
	if statsByIssueID[withComments.ID].CommentCount != 2 {
		t.Fatalf("comment count for issue with comments = %d, want 2", statsByIssueID[withComments.ID].CommentCount)
	}
	if statsByIssueID[withoutComments.ID].CommentCount != 0 {
		t.Fatalf("comment count for issue without comments = %d, want 0", statsByIssueID[withoutComments.ID].CommentCount)
	}
	queueStatusByIssueID := map[int64]entity.QueueStatus{}
	for _, column := range summary.Columns {
		for _, issue := range column.Issues {
			queueStatusByIssueID[issue.ID] = issue.QueueStatus
		}
	}
	if queueStatusByIssueID[withComments.ID] != entity.QueueStatusQueued {
		t.Fatalf("queue status for issue with comments = %q, want %q", queueStatusByIssueID[withComments.ID], entity.QueueStatusQueued)
	}
}

func TestIssuesFiltersByStates(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	ready := createIssue(t, server, "Ready issue", entity.StatusReady)
	_ = createIssue(t, server, "Backlog issue", entity.StatusBacklog)
	review := createIssue(t, server, "Review issue", entity.StatusReview)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?states=ready,review", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	issues := decodeData[[]entity.Issue](t, rec)
	if len(issues) != 2 {
		t.Fatalf("issues length = %d, issues = %+v", len(issues), issues)
	}
	assertIssueIDs(t, issues, []int64{review.ID, ready.ID})
}

func TestIssuesEmptyStatesReturnsAll(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	first := createIssue(t, server, "First issue", entity.StatusReady)
	second := createIssue(t, server, "Second issue", entity.StatusBacklog)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?states=", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	issues := decodeData[[]entity.Issue](t, rec)
	if len(issues) != 2 {
		t.Fatalf("issues length = %d, issues = %+v", len(issues), issues)
	}
	assertIssueIDs(t, issues, []int64{second.ID, first.ID})
}

func TestIssuesRejectsInvalidStates(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?states=ready,invalid", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Error.Code != "issues.list.invalid_states" {
		t.Fatalf("error.code = %q", payload.Error.Code)
	}
}

func TestProjectWorkflowRoundTrip(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	project, err := server.store.CreateProject(context.Background(), entity.CreateProjectInput{
		Key:      "WORKFLOW",
		Name:     "Workflow Project",
		Location: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	checksum := strings.Repeat("a", 64)
	body := bytes.NewBufferString(`{"frontmatter":{"agent":{"max_turns":3}},"body":"Project prompt","checksum":"` + checksum + `"}`)
	putReq := httptest.NewRequest(http.MethodPut, "/api/v1/projects/"+stringID(project.ID)+"/workflow", body)
	putRec := httptest.NewRecorder()

	server.Handler().ServeHTTP(putRec, putReq)

	if putRec.Code != http.StatusOK {
		t.Fatalf("put status = %d body=%s", putRec.Code, putRec.Body.String())
	}
	updated := decodeData[entity.ProjectWorkflow](t, putRec)
	if updated.Body != "Project prompt" || updated.Checksum != checksum {
		t.Fatalf("updated workflow = %+v", updated)
	}
	agent, ok := updated.Frontmatter["agent"].(map[string]any)
	if !ok || agent["max_turns"] != float64(3) {
		t.Fatalf("updated workflow frontmatter = %+v", updated.Frontmatter)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/projects/"+stringID(project.ID)+"/workflow", nil)
	getRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d body=%s", getRec.Code, getRec.Body.String())
	}
	read := decodeData[entity.ProjectWorkflow](t, getRec)
	if read.Body != updated.Body || read.Checksum != updated.Checksum {
		t.Fatalf("read workflow = %+v, want %+v", read, updated)
	}
}

func TestIssuesFiltersByProjectID(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	primary := createProject(t, server, "PRIMARY")
	secondary := createProject(t, server, "SECONDARY")
	first := createIssueInProject(t, server, primary.ID, "Primary issue", entity.StatusReady)
	_ = createIssueInProject(t, server, secondary.ID, "Secondary issue", entity.StatusReady)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?project_id="+stringID(primary.ID), nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	issues := decodeData[[]entity.Issue](t, rec)
	assertIssueIDs(t, issues, []int64{first.ID})
	if issues[0].ProjectID != primary.ID || issues[0].ProjectKey != primary.Key {
		t.Fatalf("issue project = %+v", issues[0])
	}
}

func TestIssuesRejectsInvalidProjectID(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?project_id=bad", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	assertErrorCode(t, rec, "issues.list.invalid_project_id")
}

func TestIssuesFiltersByProjectIDs(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	firstProject := createProject(t, server, "FIRST")
	secondProject := createProject(t, server, "SECOND")
	thirdProject := createProject(t, server, "THIRD")
	first := createIssueInProject(t, server, firstProject.ID, "First project issue", entity.StatusReady)
	second := createIssueInProject(t, server, secondProject.ID, "Second project issue", entity.StatusReady)
	_ = createIssueInProject(t, server, thirdProject.ID, "Third project issue", entity.StatusReady)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?project_ids="+stringID(firstProject.ID)+","+stringID(secondProject.ID)+"&sort_by=id&sort_direction=asc", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	issues := decodeData[[]entity.Issue](t, rec)
	assertIssueIDs(t, issues, []int64{first.ID, second.ID})
}

func TestIssuesRejectsInvalidProjectIDs(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?project_ids=1,bad", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	assertErrorCode(t, rec, "issues.list.invalid_project_ids")
}

func TestIssuesFiltersByPriorities(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	urgent := createIssueWithPriority(t, server, "Urgent issue", entity.StatusReady, entity.PriorityUrgent)
	high := createIssueWithPriority(t, server, "High issue", entity.StatusReady, entity.PriorityHigh)
	_ = createIssueWithPriority(t, server, "Normal issue", entity.StatusReady, entity.PriorityNormal)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?priorities=urgent,high&sort_by=id&sort_direction=asc", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	issues := decodeData[[]entity.Issue](t, rec)
	assertIssueIDs(t, issues, []int64{urgent.ID, high.ID})
}

func TestIssuesRejectsInvalidPriorities(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?priorities=high,unknown", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	assertErrorCode(t, rec, "issues.list.invalid_priorities")
}

func TestIssuesFiltersByAssignee(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	assigned := createIssue(t, server, "Assigned issue", entity.StatusReady)
	assignee := "frontend"
	if _, err := server.store.UpdateIssue(context.Background(), assigned.ID, entity.UpdateIssueInput{
		Assignee: &assignee,
	}); err != nil {
		t.Fatalf("update issue assignee: %v", err)
	}
	_ = createIssue(t, server, "Unassigned issue", entity.StatusReady)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?assignee=frontend", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	issues := decodeData[[]entity.Issue](t, rec)
	assertIssueIDs(t, issues, []int64{assigned.ID})
}

func TestIssuesSearchesByIDAndTitle(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	idMatch := createIssue(t, server, "Unrelated issue", entity.StatusReady)
	_ = createIssue(t, server, "No match", entity.StatusReady)
	titleMatch := createIssue(t, server, "Investigate issue "+stringID(idMatch.ID), entity.StatusReady)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?search="+stringID(idMatch.ID)+"&sort_by=id&sort_direction=asc", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	issues := decodeData[[]entity.Issue](t, rec)
	assertIssueIDs(t, issues, []int64{idMatch.ID, titleMatch.ID})
}

func TestIssuesSearchesTitleCaseInsensitively(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	match := createIssue(t, server, "Fix LOGIN flow", entity.StatusReady)
	_ = createIssue(t, server, "Fix signup flow", entity.StatusReady)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?search=login", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	issues := decodeData[[]entity.Issue](t, rec)
	assertIssueIDs(t, issues, []int64{match.ID})
}

func TestIssuesSearchCombinesWithFiltersAndPagination(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	project := createProject(t, server, "SEARCH")
	otherProject := createProject(t, server, "OTHERSEARCH")
	first := createIssueInProject(t, server, project.ID, "Login first", entity.StatusReady)
	second := createIssueInProject(t, server, project.ID, "Login second", entity.StatusReady)
	_ = createIssueInProject(t, server, project.ID, "Login backlog", entity.StatusBacklog)
	_ = createIssueInProject(t, server, otherProject.ID, "Login external", entity.StatusReady)
	_ = createIssueInProject(t, server, project.ID, "Signup ready", entity.StatusReady)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/issues?search=login&states=ready&project_id="+stringID(project.ID)+"&sort_by=id&sort_direction=asc&limit=1&offset=1",
		nil,
	)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Data []entity.Issue `json:"data"`
		Meta struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	assertIssueIDs(t, payload.Data, []int64{second.ID})
	if payload.Meta.Total != 2 {
		t.Fatalf("total = %d, want 2", payload.Meta.Total)
	}
	if first.ID >= second.ID {
		t.Fatalf("test setup issue IDs are not increasing: first=%d second=%d", first.ID, second.ID)
	}
}

func TestIssuesEmptySearchReturnsAll(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	first := createIssue(t, server, "First issue", entity.StatusReady)
	second := createIssue(t, server, "Second issue", entity.StatusBacklog)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?search=%20%20", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	issues := decodeData[[]entity.Issue](t, rec)
	assertIssueIDs(t, issues, []int64{second.ID, first.ID})
}

func TestIssuesSortsByIDAscending(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	first := createIssue(t, server, "First issue", entity.StatusReady)
	second := createIssue(t, server, "Second issue", entity.StatusReady)
	third := createIssue(t, server, "Third issue", entity.StatusReady)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?sort_by=id&sort_direction=asc", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	issues := decodeData[[]entity.Issue](t, rec)
	assertIssueIDs(t, issues, []int64{first.ID, second.ID, third.ID})
}

func TestIssuesSortsByPriorityDescending(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	low := createIssueWithPriority(t, server, "Low issue", entity.StatusReady, entity.PriorityLow)
	urgent := createIssueWithPriority(t, server, "Urgent issue", entity.StatusReady, entity.PriorityUrgent)
	normal := createIssueWithPriority(t, server, "Normal issue", entity.StatusReady, entity.PriorityNormal)
	high := createIssueWithPriority(t, server, "High issue", entity.StatusReady, entity.PriorityHigh)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?sort_by=priority&sort_direction=desc", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	issues := decodeData[[]entity.Issue](t, rec)
	assertIssueIDs(t, issues, []int64{urgent.ID, high.ID, normal.ID, low.ID})
}

func TestIssuesRejectsInvalidSort(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?sort_by=title", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	assertErrorCode(t, rec, "issues.list.invalid_sort_by")
}

func TestIssuesPaginatesWithMeta(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	first := createIssue(t, server, "First issue", entity.StatusReady)
	second := createIssue(t, server, "Second issue", entity.StatusReady)
	third := createIssue(t, server, "Third issue", entity.StatusReady)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues?limit=2&offset=0", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Data []entity.Issue `json:"data"`
		Meta struct {
			Limit      int  `json:"limit"`
			Offset     int  `json:"offset"`
			Total      int  `json:"total"`
			NextOffset *int `json:"nextOffset"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	assertIssueIDs(t, payload.Data, []int64{third.ID, second.ID})
	if payload.Meta.Limit != 2 || payload.Meta.Offset != 0 || payload.Meta.Total != 3 {
		t.Fatalf("meta = %+v", payload.Meta)
	}
	if payload.Meta.NextOffset == nil || *payload.Meta.NextOffset != 2 {
		t.Fatalf("nextOffset = %v", payload.Meta.NextOffset)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v1/issues?limit=2&offset=2", nil)
	rec = httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode second response: %v", err)
	}
	assertIssueIDs(t, payload.Data, []int64{first.ID})
	if payload.Meta.NextOffset != nil {
		t.Fatalf("nextOffset = %v, want nil", payload.Meta.NextOffset)
	}
}

func TestIssuesRejectsInvalidPagination(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		path string
		code string
	}{
		{name: "limit", path: "/api/v1/issues?limit=51", code: "issues.list.invalid_limit"},
		{name: "offset", path: "/api/v1/issues?offset=-1", code: "issues.list.invalid_offset"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := newTestServer(t)
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			server.Handler().ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d", rec.Code)
			}
			assertErrorCode(t, rec, tc.code)
		})
	}
}

func TestIssueStatesReturnsMatchingStates(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	ready := createIssue(t, server, "Ready issue", entity.StatusReady)
	_ = createIssue(t, server, "Backlog issue", entity.StatusBacklog)
	done := createIssue(t, server, "Done issue", entity.StatusDone)
	body := bytes.NewBufferString(`{"ids":[` + stringID(ready.ID) + `,999999,` + stringID(done.ID) + `]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/issues/states", body)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	states := decodeData[[]entity.IssueState](t, rec)
	if len(states) != 2 {
		t.Fatalf("states length = %d, states = %+v", len(states), states)
	}
	assertIssueStates(t, states, []entity.IssueState{
		{ID: ready.ID, Status: entity.StatusReady},
		{ID: done.ID, Status: entity.StatusDone},
	})
}

func TestIssueStatesEmptyIDsReturnsEmpty(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/issues/states", bytes.NewBufferString(`{"ids":[]}`))
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	states := decodeData[[]entity.IssueState](t, rec)
	if len(states) != 0 {
		t.Fatalf("states = %+v", states)
	}
}

func TestQueueReturnsQueuedAndPendingIssues(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	doneDep := createIssue(t, server, "Done dependency", entity.StatusDone)
	activeDep := createIssue(t, server, "Active dependency", entity.StatusReady)
	queued := createIssueWithPriority(t, server, "Queued issue", entity.StatusReady, entity.PriorityHigh)
	pending := createIssueWithPriority(t, server, "Pending issue", entity.StatusReady, entity.PriorityUrgent)
	if err := server.store.SetDependencyIDs(context.Background(), queued.ID, []int64{doneDep.ID}); err != nil {
		t.Fatalf("set queued dependencies: %v", err)
	}
	if err := server.store.SetDependencyIDs(context.Background(), pending.ID, []int64{activeDep.ID}); err != nil {
		t.Fatalf("set pending dependencies: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/queue", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	queue := decodeData[entity.Queue](t, rec)
	if len(queue.Queued) != 2 || queue.Queued[0].ID != queued.ID || queue.Queued[1].ID != activeDep.ID {
		t.Fatalf("queued = %+v", queue.Queued)
	}
	if len(queue.Pending) != 1 || queue.Pending[0].ID != pending.ID {
		t.Fatalf("pending = %+v", queue.Pending)
	}
	if len(queue.Pending[0].BlockedDependencyIDs) != 1 || queue.Pending[0].BlockedDependencyIDs[0] != activeDep.ID {
		t.Fatalf("blocked dependency ids = %+v", queue.Pending[0].BlockedDependencyIDs)
	}
}

func TestCreateIssueReturnsDependencyIDs(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	project := createProject(t, server, "CREATEAPI")
	dependency := createIssueInProject(t, server, project.ID, "Dependency issue", entity.StatusDone)
	body := bytes.NewBufferString(`{"projectId":` + stringID(project.ID) + `,"title":"Parent issue","dependency_ids":[` + stringID(dependency.ID) + `]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/issues", body)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	assertResponseDataHasKey(t, rec.Body.Bytes(), "dependency_ids")
	created := decodeData[entity.Issue](t, rec)
	assertInt64s(t, created.DependencyIDs, []int64{dependency.ID})
}

func TestCreateIssueWithoutDependencyIDsReturnsEmptyDependencyIDs(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	project := createProject(t, server, "CREATENODEPS")
	body := bytes.NewBufferString(`{"projectId":` + stringID(project.ID) + `,"title":"Independent issue"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/issues", body)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	assertResponseDataHasKey(t, rec.Body.Bytes(), "dependency_ids")
	created := decodeData[entity.Issue](t, rec)
	if created.DependencyIDs == nil || len(created.DependencyIDs) != 0 {
		t.Fatalf("dependency ids = %+v, want empty slice", created.DependencyIDs)
	}
}

func TestCreateIssueRejectsSelfDependency(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	project := createProject(t, server, "CREATESELF")
	existing := createIssueInProject(t, server, project.ID, "Existing issue", entity.StatusBacklog)
	body := bytes.NewBufferString(`{"projectId":` + stringID(project.ID) + `,"title":"Self dependency","dependency_ids":[` + stringID(existing.ID+1) + `]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/issues", body)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec, "issues.create.invalid_input")
}

func TestCreateIssueRejectsMissingDependency(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	project := createProject(t, server, "CREATEMISSING")
	body := bytes.NewBufferString(`{"projectId":` + stringID(project.ID) + `,"title":"Missing dependency","dependency_ids":[999999]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/issues", body)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec, "issues.create.invalid_input")
}

func TestCreateIssueRejectsDuplicateDependencyIDs(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	project := createProject(t, server, "CREATEDUP")
	dependency := createIssueInProject(t, server, project.ID, "Dependency issue", entity.StatusBacklog)
	body := bytes.NewBufferString(`{"projectId":` + stringID(project.ID) + `,"title":"Duplicate dependency","dependency_ids":[` + stringID(dependency.ID) + `,` + stringID(dependency.ID) + `]}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/issues", body)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec, "issues.create.invalid_input")
}

func TestIssueRoundTripDependencyIDs(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	parent := createIssue(t, server, "Parent issue", entity.StatusReady)
	dependency := createIssue(t, server, "Dependency issue", entity.StatusDone)
	body := bytes.NewBufferString(`{"dependency_ids":[` + stringID(dependency.ID) + `]}`)
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/issues/"+stringID(parent.ID), body)
	patchRec := httptest.NewRecorder()

	server.Handler().ServeHTTP(patchRec, patchReq)

	if patchRec.Code != http.StatusOK {
		t.Fatalf("patch status = %d body=%s", patchRec.Code, patchRec.Body.String())
	}
	updated := decodeData[entity.Issue](t, patchRec)
	if len(updated.DependencyIDs) != 1 || updated.DependencyIDs[0] != dependency.ID {
		t.Fatalf("updated dependency ids = %+v", updated.DependencyIDs)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/v1/issues/"+stringID(parent.ID), nil)
	getRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("get status = %d body=%s", getRec.Code, getRec.Body.String())
	}
	read := decodeData[entity.Issue](t, getRec)
	if len(read.DependencyIDs) != 1 || read.DependencyIDs[0] != dependency.ID {
		t.Fatalf("read dependency ids = %+v", read.DependencyIDs)
	}

	clearReq := httptest.NewRequest(http.MethodPatch, "/api/v1/issues/"+stringID(parent.ID), bytes.NewBufferString(`{"dependency_ids":[]}`))
	clearRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(clearRec, clearReq)

	if clearRec.Code != http.StatusOK {
		t.Fatalf("clear status = %d body=%s", clearRec.Code, clearRec.Body.String())
	}
	cleared := decodeData[entity.Issue](t, clearRec)
	if cleared.DependencyIDs == nil || len(cleared.DependencyIDs) != 0 {
		t.Fatalf("cleared dependency ids = %+v", cleared.DependencyIDs)
	}
}

func TestUpdateIssueRejectsDependencyCycle(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	parent := createIssue(t, server, "Parent issue", entity.StatusReady)
	dependency := createIssue(t, server, "Dependency issue", entity.StatusReady)
	if err := server.store.SetDependencyIDs(context.Background(), parent.ID, []int64{dependency.ID}); err != nil {
		t.Fatalf("set parent dependency: %v", err)
	}
	body := bytes.NewBufferString(`{"dependency_ids":[` + stringID(parent.ID) + `]}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/issues/"+stringID(dependency.ID), body)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), stringID(dependency.ID)) || !strings.Contains(rec.Body.String(), stringID(parent.ID)) {
		t.Fatalf("cycle error does not include issue ids: %s", rec.Body.String())
	}
	assertErrorCode(t, rec, "issues.update.invalid_input")
}

func TestUpdateIssueRejectsMissingDependency(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	parent := createIssue(t, server, "Parent issue", entity.StatusReady)
	body := bytes.NewBufferString(`{"dependency_ids":[999999]}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/issues/"+stringID(parent.ID), body)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec, "issues.update.invalid_input")
}

func TestIssueStatesRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/issues/states", bytes.NewBufferString(`{"ids":[`))
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCreateComment(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	issue := createIssue(t, server, "Commented issue", entity.StatusBacklog)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/issues/"+stringID(issue.ID)+"/comments", bytes.NewBufferString(`{
		"author": "codex",
		"type": "progress",
		"body": "Added comment endpoint."
	}`))
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	comment := decodeData[entity.Comment](t, rec)
	if comment.IssueID != issue.ID || comment.Author != "codex" || comment.Type != entity.CommentProgress || comment.Body != "Added comment endpoint." {
		t.Fatalf("comment = %+v", comment)
	}
}

func TestCreateCommentMissingIssue(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/issues/999999/comments", bytes.NewBufferString(`{
		"author": "codex",
		"body": "Missing issue."
	}`))
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
	assertErrorCode(t, rec, "comments.create.issue_not_found")
}

func TestCreateCommentRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	issue := createIssue(t, server, "Invalid comment", entity.StatusBacklog)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/issues/"+stringID(issue.ID)+"/comments", bytes.NewBufferString(`{
		"author": "",
		"body": "Missing author."
	}`))
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	assertErrorCode(t, rec, "comments.create.invalid_input")
}

func TestCommentsPagination(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	issue := createIssue(t, server, "List comments", entity.StatusBacklog)
	first := createComment(t, server, issue.ID, "first")
	second := createComment(t, server, issue.ID, "second")
	third := createComment(t, server, issue.ID, "third")
	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues/"+stringID(issue.ID)+"/comments?cursor="+stringID(first.ID)+"&limit=2", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var payload struct {
		Data []entity.Comment `json:"data"`
		Meta struct {
			Cursor     int64  `json:"cursor"`
			Limit      int    `json:"limit"`
			NextCursor *int64 `json:"nextCursor"`
		} `json:"meta"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	assertCommentIDs(t, payload.Data, []int64{second.ID, third.ID})
	if payload.Meta.Cursor != first.ID || payload.Meta.Limit != 2 {
		t.Fatalf("meta = %+v", payload.Meta)
	}
	if payload.Meta.NextCursor == nil || *payload.Meta.NextCursor != third.ID {
		t.Fatalf("nextCursor = %v, want %d", payload.Meta.NextCursor, third.ID)
	}
}

func TestCommentsRejectsInvalidQuery(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	issue := createIssue(t, server, "Invalid comment query", entity.StatusBacklog)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/issues/"+stringID(issue.ID)+"/comments?cursor=bad", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	assertErrorCode(t, rec, "comments.list.invalid_input")
}

func TestAttachmentUploadListContentAndDelete(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	issue := createIssue(t, server, "Issue with image", entity.StatusBacklog)
	req := newAttachmentUploadRequest(t, entity.AttachmentEntityIssue, stringID(issue.ID), "screenshot.png", "image/png", []byte{
		0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n',
	})
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	attachment := decodeData[entity.Attachment](t, rec)
	if attachment.ID == "" || attachment.EntityType != entity.AttachmentEntityIssue || attachment.EntityID != stringID(issue.ID) || attachment.Filename != "screenshot.png" {
		t.Fatalf("attachment = %+v", attachment)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/attachments?entity_type=issue&entity_id="+stringID(issue.ID), nil)
	listRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d, body = %s", listRec.Code, listRec.Body.String())
	}
	attachments := decodeData[[]entity.Attachment](t, listRec)
	if len(attachments) != 1 || attachments[0].ID != attachment.ID {
		t.Fatalf("attachments = %+v", attachments)
	}

	contentReq := httptest.NewRequest(http.MethodGet, "/api/v1/attachments/"+attachment.ID+"/content", nil)
	contentRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(contentRec, contentReq)
	if contentRec.Code != http.StatusOK {
		t.Fatalf("content status = %d, body = %s", contentRec.Code, contentRec.Body.String())
	}
	if contentRec.Header().Get("Content-Type") != "image/png" || !bytes.HasPrefix(contentRec.Body.Bytes(), []byte{0x89, 'P', 'N', 'G'}) {
		t.Fatalf("content headers=%v body=%x", contentRec.Header(), contentRec.Body.Bytes())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/v1/attachments/"+attachment.ID, nil)
	deleteRec := httptest.NewRecorder()
	server.Handler().ServeHTTP(deleteRec, deleteReq)
	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, body = %s", deleteRec.Code, deleteRec.Body.String())
	}
}

func TestAttachmentUploadRejectsInvalidType(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	issue := createIssue(t, server, "Issue with invalid image", entity.StatusBacklog)
	req := newAttachmentUploadRequest(t, entity.AttachmentEntityIssue, stringID(issue.ID), "note.txt", "text/plain", []byte("hello"))
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	assertErrorCode(t, rec, "attachments.create.invalid_file_type")
}

func newTestServer(t *testing.T) *Server {
	t.Helper()

	issueStore, err := store.OpenMigrated(context.Background(), filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() {
		if err := issueStore.Close(); err != nil {
			t.Fatalf("close store: %v", err)
		}
	})
	return NewServerWithAttachmentStorage(issueStore, store.NewAttachmentStorage(t.TempDir()))
}

func stringID(id int64) string {
	return strconv.FormatInt(id, 10)
}

func createIssue(t *testing.T, server *Server, title string, status entity.Status) entity.Issue {
	t.Helper()

	project, err := server.store.ProjectByKey(context.Background(), "TEST")
	if err != nil {
		project = createProject(t, server, "TEST")
	}
	return createIssueInProject(t, server, project.ID, title, status)
}

func createIssueInProject(t *testing.T, server *Server, projectID int64, title string, status entity.Status) entity.Issue {
	t.Helper()

	return createIssueInProjectWithPriority(t, server, projectID, title, status, "")
}

func createIssueWithPriority(t *testing.T, server *Server, title string, status entity.Status, priority entity.Priority) entity.Issue {
	t.Helper()

	project, err := server.store.ProjectByKey(context.Background(), "TEST")
	if err != nil {
		project = createProject(t, server, "TEST")
	}
	return createIssueInProjectWithPriority(t, server, project.ID, title, status, priority)
}

func createIssueInProjectWithPriority(t *testing.T, server *Server, projectID int64, title string, status entity.Status, priority entity.Priority) entity.Issue {
	t.Helper()

	issue, err := server.store.CreateIssue(context.Background(), entity.CreateIssueInput{
		ProjectID: projectID,
		Title:     title,
		Status:    status,
		Priority:  priority,
	})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	return issue
}

func createProject(t *testing.T, server *Server, key string) entity.Project {
	t.Helper()

	project, err := server.store.CreateProject(context.Background(), entity.CreateProjectInput{
		Key:      key,
		Name:     key,
		Location: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	return project
}

func createComment(t *testing.T, server *Server, issueID int64, body string) entity.Comment {
	t.Helper()

	comment, err := server.store.CreateComment(context.Background(), entity.CreateCommentInput{
		IssueID: issueID,
		Author:  "codex",
		Type:    entity.CommentGeneral,
		Body:    body,
	})
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	return comment
}

func decodeData[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()

	var payload struct {
		Data T `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return payload.Data
}

func assertResponseDataHasKey(t *testing.T, body []byte, key string) {
	t.Helper()

	var payload struct {
		Data map[string]json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := payload.Data[key]; !ok {
		t.Fatalf("response data missing key %q: %s", key, string(body))
	}
}

func assertInt64s(t *testing.T, got []int64, want []int64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("ids = %+v, want %+v", got, want)
	}
	for i, id := range want {
		if got[i] != id {
			t.Fatalf("ids = %+v, want %+v", got, want)
		}
	}
}

func newAttachmentUploadRequest(t *testing.T, entityType string, entityID string, filename string, contentType string, content []byte) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for name, value := range map[string]string{
		"entity_type": entityType,
		"entity_id":   entityID,
	} {
		field, err := writer.CreateFormField(name)
		if err != nil {
			t.Fatalf("create field %s: %v", name, err)
		}
		if _, err := io.WriteString(field, value); err != nil {
			t.Fatalf("write field %s: %v", name, err)
		}
	}
	partHeader := make(textproto.MIMEHeader)
	partHeader.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{
		"name":     "file",
		"filename": filename,
	}))
	partHeader.Set("Content-Type", contentType)
	part, err := writer.CreatePart(partHeader)
	if err != nil {
		t.Fatalf("create file part: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("write file part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/attachments", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func assertErrorCode(t *testing.T, rec *httptest.ResponseRecorder, want string) {
	t.Helper()

	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Error.Code != want {
		t.Fatalf("error.code = %q, want %q", payload.Error.Code, want)
	}
}

func assertIssueIDs(t *testing.T, issues []entity.Issue, want []int64) {
	t.Helper()

	if len(issues) != len(want) {
		t.Fatalf("issue length = %d, want %d", len(issues), len(want))
	}
	for i, issue := range issues {
		if issue.ID != want[i] {
			t.Fatalf("issues[%d].id = %d, want %d; issues = %+v", i, issue.ID, want[i], issues)
		}
	}
}

func assertCommentIDs(t *testing.T, comments []entity.Comment, want []int64) {
	t.Helper()

	if len(comments) != len(want) {
		t.Fatalf("comment length = %d, want %d; comments = %+v", len(comments), len(want), comments)
	}
	for i, comment := range comments {
		if comment.ID != want[i] {
			t.Fatalf("comments[%d].id = %d, want %d; comments = %+v", i, comment.ID, want[i], comments)
		}
	}
}

func assertIssueStates(t *testing.T, got []entity.IssueState, want []entity.IssueState) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("states length = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("states[%d] = %+v, want %+v; states = %+v", i, got[i], want[i], got)
		}
	}
}
