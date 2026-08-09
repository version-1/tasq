package tq

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func TestTUIAliasesUseCanonicalHelp(t *testing.T) {
	for _, command := range []string{"tui", "console", "c"} {
		t.Run(command, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := run(context.Background(), []string{command, "--help"}, strings.NewReader(""), &stdout, &stderr)
			if code != 0 || stderr.Len() != 0 {
				t.Fatalf("code=%d stderr=%s", code, stderr.String())
			}
			if !strings.Contains(stdout.String(), "Usage: tq [--api-url URL] [--output text] tui") || !strings.Contains(stdout.String(), "Aliases: tq console, tq c") {
				t.Fatalf("help = %s", stdout.String())
			}
		})
	}
}

func TestTUIRejectsNonTextAndNonTTYBeforeRequests(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { requests++ }))
	defer server.Close()

	for _, args := range [][]string{{"--api-url", server.URL, "--output", "json", "tui"}, {"--api-url", server.URL, "tui"}} {
		var stdout, stderr bytes.Buffer
		code := run(context.Background(), args, strings.NewReader(""), &stdout, &stderr)
		if code != 2 {
			t.Fatalf("args=%v code=%d stderr=%s", args, code, stderr.String())
		}
	}
	if requests != 0 {
		t.Fatalf("requests=%d, want no requests before validation", requests)
	}
}

func TestTUIReadClientsSendOnlyGET(t *testing.T) {
	var methods []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		methods = append(methods, r.Method+" "+r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/issues":
			if r.URL.Query().Get("limit") != "50" || r.URL.Query().Get("sort_by") != "updated_at" || r.URL.Query().Get("sort_direction") != "desc" {
				t.Fatalf("query=%s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"data":[],"meta":{"nextOffset":null}}`))
		case "/api/v1/projects":
			_, _ = w.Write([]byte(`{"data":[],"meta":{}}`))
		case "/api/v1/issues/42":
			_, _ = w.Write([]byte(`{"data":{"id":42,"projectId":1,"projectKey":"tasq","title":"TUI","description":"","status":"ready","priority":"normal","assignee":"","dependency_ids":[],"artifacts":[],"createdAt":"2026-01-01T00:00:00Z","updatedAt":"2026-01-01T00:00:00Z"},"meta":{}}`))
		case "/api/v1/issues/42/comments":
			if r.URL.Query().Get("direction") != "backward" || r.URL.Query().Get("limit") != "50" {
				t.Fatalf("comments query=%s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"data":[],"meta":{"nextCursor":null}}`))
		case "/api/v1/issue-42":
			_, _ = w.Write([]byte(`{"status":"running","workspace":{"path":"/tmp/work"},"attempts":{"restart_count":0,"current_retry_attempt":1},"runs":[],"recent_events":[],"last_error":null}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	tracker, err := newAPIClient(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	orchestrator, err := newOrchestratorClient(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := tracker.listIssuesPage(ctx, issueListQuery{}); err != nil {
		t.Fatal(err)
	}
	if _, err := tracker.listProjects(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := tracker.getIssue(ctx, 42); err != nil {
		t.Fatal(err)
	}
	if _, err := tracker.commentsPage(ctx, 42, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := orchestrator.issue(ctx, 42); err != nil {
		t.Fatal(err)
	}
	for _, method := range methods {
		if !strings.HasPrefix(method, http.MethodGet+" ") {
			t.Fatalf("mutating request observed: %s", method)
		}
	}
}

func TestTUIModelRejectsStaleResultsAndPreservesSelection(t *testing.T) {
	now := time.Now()
	m := newTUIModel(context.Background(), nil, nil, func(string) error { return nil })
	m.generation = 2
	m.listRequest = 3
	m.issues = []entity.Issue{{ID: 1, UpdatedAt: now}, {ID: 2, UpdatedAt: now.Add(-time.Minute)}}
	m.selection = 1

	updated, _ := m.Update(listResultMsg{generation: 1, request: 3, page: issuePage{Issues: []entity.Issue{{ID: 9}}}})
	stale := updated.(tuiModel)
	if len(stale.issues) != 2 {
		t.Fatalf("stale result changed issues: %+v", stale.issues)
	}

	updated, _ = stale.Update(listResultMsg{generation: 2, request: 3, page: issuePage{Issues: []entity.Issue{{ID: 2, UpdatedAt: now.Add(time.Minute)}, {ID: 3, UpdatedAt: now}}}})
	fresh := updated.(tuiModel)
	if fresh.selectedID() != 2 {
		t.Fatalf("selected=%d, want 2", fresh.selectedID())
	}
}

func TestTUIModelRejectsOlderRefreshAndKeepsLoadedTail(t *testing.T) {
	now := time.Now()
	next := 100
	m := newTUIModel(context.Background(), nil, nil, func(string) error { return nil })
	m.replaceList = false
	m.listRequest = 2
	m.nextOffset = &next
	m.issues = []entity.Issue{{ID: 60, UpdatedAt: now.Add(-time.Minute)}, {ID: 1, UpdatedAt: now}}
	m.selection = 0

	updated, _ := m.Update(listResultMsg{generation: 0, request: 1, page: issuePage{Issues: []entity.Issue{{ID: 9}}}})
	if got := updated.(tuiModel); len(got.issues) != 2 {
		t.Fatalf("stale refresh changed issues: %+v", got.issues)
	}

	updated, _ = updated.(tuiModel).Update(listResultMsg{generation: 0, request: 2, page: issuePage{Issues: []entity.Issue{{ID: 1, UpdatedAt: now.Add(time.Minute)}}}})
	got := updated.(tuiModel)
	if got.selectedID() != 60 || got.nextOffset == nil || *got.nextOffset != 100 {
		t.Fatalf("selected=%d next=%v issues=%+v", got.selectedID(), got.nextOffset, got.issues)
	}
}

func TestDetailRefreshPreservesOlderComments(t *testing.T) {
	m := newTUIModel(context.Background(), nil, nil, func(string) error { return nil })
	m.issues = []entity.Issue{{ID: 15}}
	m.detail = &entity.Issue{ID: 15}
	m.detailRequest = 4
	cursor := int64(10)
	m.commentCursor = &cursor
	m.comments = []entity.Comment{{ID: 10}, {ID: 20, Body: "old"}}
	refreshedCursor := int64(20)
	updated, _ := m.Update(detailResultMsg{generation: 0, request: 4, issueID: 15, issue: entity.Issue{ID: 15}, comments: commentPage{Comments: []entity.Comment{{ID: 20, Body: "fresh"}, {ID: 30}}, NextCursor: &refreshedCursor}, runtimeAbsent: true})
	got := updated.(tuiModel)
	if len(got.comments) != 3 || got.comments[0].ID != 10 || got.comments[1].Body != "fresh" || got.commentCursor == nil || *got.commentCursor != 10 {
		t.Fatalf("comments=%+v cursor=%v", got.comments, got.commentCursor)
	}
}

func TestTUIResponsiveAndDegradedViews(t *testing.T) {
	m := newTUIModel(context.Background(), nil, nil, func(string) error { return nil })
	m.width, m.height = minimumTUIWidth-1, minimumTUIHeight
	if !strings.Contains(m.View(), "Terminal too small") || !strings.Contains(m.View(), "EXPERIMENTAL") {
		t.Fatalf("view=%s", m.View())
	}
	m.width, m.height = 100, 30
	m.detail = &entity.Issue{ID: 15, Title: "Console", Artifacts: []entity.Artifact{{Type: entity.ArtifactTypePullRequest, DataType: entity.ArtifactDataTypeURL, DataValue: "https://example.com/pr/1"}}}
	m.tab = tabRun
	m.runtimeError = "connection refused"
	m.runtimeErrorAt = time.Date(2026, 8, 9, 1, 2, 3, 0, time.UTC)
	m.resizeViewport()
	if !strings.Contains(m.View(), "Run data unavailable") || !strings.Contains(m.View(), "connection refused") {
		t.Fatalf("view=%s", m.View())
	}
}

func TestOpenArtifactAllowsOnlyHTTPURLs(t *testing.T) {
	opened := ""
	m := newTUIModel(context.Background(), nil, nil, func(value string) error { opened = value; return nil })
	m.detail = &entity.Issue{Artifacts: []entity.Artifact{{DataValue: "javascript:alert(1)"}, {DataValue: "https://example.com/pr/1"}}}
	m.artifactSelection = 1
	msg := m.openArtifact()()
	if result := msg.(browserResultMsg); result.err != nil {
		t.Fatal(result.err)
	}
	if opened != "https://example.com/pr/1" {
		t.Fatalf("opened=%q", opened)
	}
}

func TestWideArtifactSelectionUsesHorizontalKeys(t *testing.T) {
	m := newTUIModel(context.Background(), nil, nil, func(string) error { return nil })
	m.width = narrowTUIWidth
	m.tab = tabArtifacts
	m.detail = &entity.Issue{Artifacts: []entity.Artifact{{DataValue: "https://example.com/1"}, {DataValue: "https://example.com/2"}}}
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	if got := updated.(tuiModel).artifactSelection; got != 1 {
		t.Fatalf("selection=%d, want 1", got)
	}
}

func TestOrchestratorClientClassifiesNotFound(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	client, err := newOrchestratorClient(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.issue(context.Background(), 1)
	if !errors.Is(err, errRuntimeNotFound) {
		t.Fatalf("err=%v", err)
	}
	if _, err := newOrchestratorClient("localhost:1234"); err == nil {
		t.Fatal("expected invalid URL")
	}
}

func TestIssueListQueryUsesServerSideFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query, _ := url.ParseQuery(r.URL.RawQuery)
		if query.Get("states") != "ready,blocked" || query.Get("project_ids") != "2,3" || query.Get("search") != "terminal" || query.Get("offset") != "50" {
			t.Fatalf("query=%v", query)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"data": []any{}, "meta": map[string]any{}})
	}))
	defer server.Close()
	client, _ := newAPIClient(server.URL)
	_, err := client.listIssuesPage(context.Background(), issueListQuery{States: []entity.Status{entity.StatusReady, entity.StatusBlocked}, ProjectIDs: []int64{2, 3}, Search: "terminal", Offset: 50})
	if err != nil {
		t.Fatal(err)
	}
}

var _ tea.Model = tuiModel{}
