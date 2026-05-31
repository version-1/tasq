package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
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

func TestNoContentResponseStaysEmpty(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	project, err := server.store.CreateProject(context.Background(), entity.CreateProjectInput{
		Key:  "docs",
		Name: "Docs",
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

func newTestServer(t *testing.T) *Server {
	t.Helper()

	store, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("close store: %v", err)
		}
	})
	return NewServer(store)
}

func stringID(id int64) string {
	return strconv.FormatInt(id, 10)
}

func createIssue(t *testing.T, server *Server, title string, status entity.Status) entity.Issue {
	t.Helper()

	issue, err := server.store.CreateIssue(context.Background(), entity.CreateIssueInput{
		Title:  title,
		Status: status,
	})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	return issue
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
