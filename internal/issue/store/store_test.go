package store

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func TestOpenAppliesIssueTrackerSchema(t *testing.T) {
	t.Parallel()

	store, err := Open(context.Background(), filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	for _, name := range []string{
		"issues",
		"issues_status_idx",
		"projects",
		"projects_key_idx",
		"workspaces",
		"workspaces_project_id_idx",
		"workspaces_status_idx",
		"work_items",
		"work_items_status_lease_idx",
		"orchestrator_events",
		"run_snapshots",
	} {
		if !schemaObjectExists(t, store, name) {
			t.Fatalf("schema object %q does not exist", name)
		}
	}
}

func TestProjectCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	created, err := store.CreateProject(ctx, entity.CreateProjectInput{
		Key:         "product",
		Name:        "Product Website",
		Description: "Public marketing and product site",
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("created project id is zero")
	}
	if created.Key != "product" || created.Name != "Product Website" || created.Description != "Public marketing and product site" {
		t.Fatalf("created project = %+v", created)
	}

	renamed := "Product Experience"
	updated, err := store.UpdateProject(ctx, created.ID, entity.UpdateProjectInput{Name: &renamed})
	if err != nil {
		t.Fatalf("update project: %v", err)
	}
	if updated.Name != "Product Experience" {
		t.Fatalf("updated project name = %q", updated.Name)
	}

	projects, err := store.Projects(ctx)
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("project count = %d", len(projects))
	}

	if err := store.DeleteProject(ctx, created.ID); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	projects, err = store.Projects(ctx)
	if err != nil {
		t.Fatalf("list projects after delete: %v", err)
	}
	if len(projects) != 0 {
		t.Fatalf("project count after delete = %d", len(projects))
	}
}

func TestProjectsReturnsEmptyArray(t *testing.T) {
	t.Parallel()

	store, err := Open(context.Background(), filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	projects, err := store.Projects(context.Background())
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if projects == nil {
		t.Fatal("projects is nil")
	}

	payload, err := json.Marshal(projects)
	if err != nil {
		t.Fatalf("marshal projects: %v", err)
	}
	if string(payload) != "[]" {
		t.Fatalf("projects json = %s", payload)
	}
}

func TestWorkspaceCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project, err := store.CreateProject(ctx, entity.CreateProjectInput{Key: "api", Name: "API Backend"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	created, err := store.CreateWorkspace(ctx, entity.CreateWorkspaceInput{
		ProjectID: project.ID,
		Name:      "API Main",
		Path:      ".workspaces/api-main",
	})
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if created.Status != entity.WorkspaceActive {
		t.Fatalf("created workspace status = %q", created.Status)
	}

	status := entity.WorkspaceInactive
	updated, err := store.UpdateWorkspace(ctx, created.ID, entity.UpdateWorkspaceInput{Status: &status})
	if err != nil {
		t.Fatalf("update workspace: %v", err)
	}
	if updated.Status != entity.WorkspaceInactive {
		t.Fatalf("updated workspace status = %q", updated.Status)
	}

	workspaces, err := store.Workspaces(ctx)
	if err != nil {
		t.Fatalf("list workspaces: %v", err)
	}
	if len(workspaces) != 1 {
		t.Fatalf("workspace count = %d", len(workspaces))
	}

	if err := store.DeleteWorkspace(ctx, created.ID); err != nil {
		t.Fatalf("delete workspace: %v", err)
	}
	workspaces, err = store.Workspaces(ctx)
	if err != nil {
		t.Fatalf("list workspaces after delete: %v", err)
	}
	if len(workspaces) != 0 {
		t.Fatalf("workspace count after delete = %d", len(workspaces))
	}
}

func TestIssuesByStates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	backlog, err := store.CreateIssue(ctx, entity.CreateIssueInput{Title: "Plan tracker", Status: entity.StatusBacklog})
	if err != nil {
		t.Fatalf("create backlog issue: %v", err)
	}
	ready, err := store.CreateIssue(ctx, entity.CreateIssueInput{Title: "Build worker", Status: entity.StatusReady})
	if err != nil {
		t.Fatalf("create ready issue: %v", err)
	}
	review, err := store.CreateIssue(ctx, entity.CreateIssueInput{Title: "Review API", Status: entity.StatusReview})
	if err != nil {
		t.Fatalf("create review issue: %v", err)
	}

	filtered, err := store.IssuesByStates(ctx, []entity.Status{entity.StatusReady, entity.StatusReview})
	if err != nil {
		t.Fatalf("list issues by states: %v", err)
	}
	assertIssueIDs(t, filtered, []int64{ready.ID, review.ID})
	assertIssueStatuses(t, filtered, []entity.Status{entity.StatusReady, entity.StatusReview})

	allByStates, err := store.IssuesByStates(ctx, nil)
	if err != nil {
		t.Fatalf("list issues with empty states: %v", err)
	}
	all, err := store.Issues(ctx)
	if err != nil {
		t.Fatalf("list all issues: %v", err)
	}
	assertIssueIDs(t, allByStates, []int64{backlog.ID, ready.ID, review.ID})
	if len(allByStates) != len(all) {
		t.Fatalf("empty states issue count = %d, all issue count = %d", len(allByStates), len(all))
	}
}

func TestIssueStatesByIDs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	backlog, err := store.CreateIssue(ctx, entity.CreateIssueInput{Title: "Plan tracker", Status: entity.StatusBacklog})
	if err != nil {
		t.Fatalf("create backlog issue: %v", err)
	}
	ready, err := store.CreateIssue(ctx, entity.CreateIssueInput{Title: "Build worker", Status: entity.StatusReady})
	if err != nil {
		t.Fatalf("create ready issue: %v", err)
	}
	review, err := store.CreateIssue(ctx, entity.CreateIssueInput{Title: "Review API", Status: entity.StatusReview})
	if err != nil {
		t.Fatalf("create review issue: %v", err)
	}

	states, err := store.IssueStatesByIDs(ctx, []int64{review.ID, 999999, backlog.ID, ready.ID})
	if err != nil {
		t.Fatalf("list issue states by ids: %v", err)
	}
	if len(states) != 3 {
		t.Fatalf("issue state count = %d, states = %+v", len(states), states)
	}
	got := map[int64]entity.Status{}
	for _, state := range states {
		got[state.ID] = state.Status
	}
	for id, want := range map[int64]entity.Status{
		backlog.ID: entity.StatusBacklog,
		ready.ID:   entity.StatusReady,
		review.ID:  entity.StatusReview,
	} {
		if got[id] != want {
			t.Fatalf("issue state for id %d = %q, want %q", id, got[id], want)
		}
	}
	if _, ok := got[999999]; ok {
		t.Fatal("missing issue id was returned")
	}
}

func TestIssueStatesByIDsWithEmptyIDsDoesNotQuery(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}

	states, err := store.IssueStatesByIDs(ctx, nil)
	if err != nil {
		t.Fatalf("list issue states by empty ids: %v", err)
	}
	if states == nil {
		t.Fatal("issue states is nil")
	}
	if len(states) != 0 {
		t.Fatalf("issue state count = %d", len(states))
	}
}

func TestSummaryReturnsEmptyRunsArray(t *testing.T) {
	t.Parallel()

	store, err := Open(context.Background(), filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	summary, err := store.Summary(context.Background())
	if err != nil {
		t.Fatalf("read summary: %v", err)
	}
	if summary.Runs == nil {
		t.Fatal("summary runs is nil")
	}

	payload, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	if !strings.Contains(string(payload), `"runs":[]`) {
		t.Fatalf("summary json does not contain empty runs array: %s", payload)
	}
}

func TestRenewWorkItemLeaseRequiresCurrentClaim(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	issue, err := store.CreateIssue(ctx, entity.CreateIssueInput{
		Title:  "Implement runner",
		Status: entity.StatusReady,
	})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	item, err := store.ClaimWorkItem(ctx, entity.ClaimWorkItemInput{
		OrchestratorID: "orch-1",
		LeaseSeconds:   1,
	})
	if err != nil {
		t.Fatalf("claim work item: %v", err)
	}
	if item == nil || item.IssueID != issue.ID {
		t.Fatalf("claimed item = %+v, issue_id = %d", item, issue.ID)
	}

	renewed, err := store.RenewWorkItemLease(ctx, entity.RenewWorkItemLeaseInput{
		WorkItemID:     item.ID,
		ClaimToken:     item.ClaimToken,
		OrchestratorID: "orch-1",
		LeaseSeconds:   60,
	})
	if err != nil {
		t.Fatalf("renew lease: %v", err)
	}
	if renewed.LeaseUntil == nil || !renewed.LeaseUntil.After(*item.LeaseUntil) {
		t.Fatalf("lease was not extended: before=%v after=%v", item.LeaseUntil, renewed.LeaseUntil)
	}

	if _, err := store.RenewWorkItemLease(ctx, entity.RenewWorkItemLeaseInput{
		WorkItemID:     item.ID,
		ClaimToken:     "stale-token",
		OrchestratorID: "orch-1",
		LeaseSeconds:   60,
	}); err == nil {
		t.Fatal("expected stale token renewal to fail")
	}
}

func schemaObjectExists(t *testing.T, store *Store, name string) bool {
	t.Helper()

	var exists bool
	err := store.db.QueryRow(`SELECT EXISTS (
		SELECT 1 FROM sqlite_master WHERE name = ?
	)`, name).Scan(&exists)
	if err != nil {
		t.Fatalf("query schema object %q: %v", name, err)
	}
	return exists
}

func assertIssueIDs(t *testing.T, issues []entity.Issue, want []int64) {
	t.Helper()

	got := map[int64]bool{}
	for _, issue := range issues {
		got[issue.ID] = true
	}
	for _, id := range want {
		if !got[id] {
			t.Fatalf("issue id %d is missing from %+v", id, issues)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("issue ids count = %d, want %d, issues = %+v", len(got), len(want), issues)
	}
}

func assertIssueStatuses(t *testing.T, issues []entity.Issue, want []entity.Status) {
	t.Helper()

	allowed := map[entity.Status]bool{}
	for _, status := range want {
		allowed[status] = true
	}
	for _, issue := range issues {
		if !allowed[issue.Status] {
			t.Fatalf("issue status %q is not expected in %+v", issue.Status, issues)
		}
	}
}
