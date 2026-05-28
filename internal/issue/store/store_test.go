package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/version-1/tasq/internal/issue"
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

	created, err := store.CreateProject(ctx, issue.CreateProjectInput{
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
	updated, err := store.UpdateProject(ctx, created.ID, issue.UpdateProjectInput{Name: &renamed})
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

func TestWorkspaceCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project, err := store.CreateProject(ctx, issue.CreateProjectInput{Key: "api", Name: "API Backend"})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	created, err := store.CreateWorkspace(ctx, issue.CreateWorkspaceInput{
		ProjectID: project.ID,
		Name:      "API Main",
		Path:      ".workspaces/api-main",
	})
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if created.Status != issue.WorkspaceActive {
		t.Fatalf("created workspace status = %q", created.Status)
	}

	status := issue.WorkspaceInactive
	updated, err := store.UpdateWorkspace(ctx, created.ID, issue.UpdateWorkspaceInput{Status: &status})
	if err != nil {
		t.Fatalf("update workspace: %v", err)
	}
	if updated.Status != issue.WorkspaceInactive {
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
