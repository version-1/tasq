package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManagerCreatesSanitizedWorkspaceUnderRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	manager, err := NewManager(root)
	if err != nil {
		t.Fatalf("create workspace manager: %v", err)
	}

	workspace, err := manager.CreateForIssue("TEAM/123: Add task")
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	if workspace.WorkspaceKey != "TEAM_123__Add_task" {
		t.Fatalf("workspace key = %q", workspace.WorkspaceKey)
	}
	if !workspace.CreatedNow {
		t.Fatal("workspace was not marked created")
	}
	if _, err := os.Stat(workspace.Path); err != nil {
		t.Fatalf("stat workspace: %v", err)
	}
	if filepath.Dir(workspace.Path) != root {
		t.Fatalf("workspace path = %q, root = %q", workspace.Path, root)
	}
}

func TestManagerRejectsExistingFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	manager, err := NewManager(root)
	if err != nil {
		t.Fatalf("create workspace manager: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "TASK-1"), []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	if _, err := manager.CreateForIssue("TASK-1"); err == nil {
		t.Fatal("expected existing file error")
	}
}

func TestManagerPopulatesWorkspaceFromSource(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "workspaces")
	source := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(filepath.Join(source, "docs"), 0o755); err != nil {
		t.Fatalf("create source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(source, "docs", "README.md"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write source file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(source, ".git"), 0o755); err != nil {
		t.Fatalf("create git dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(source, ".git", "config"), []byte("secret"), 0o644); err != nil {
		t.Fatalf("write git file: %v", err)
	}
	manager, err := NewManagerWithSource(root, source)
	if err != nil {
		t.Fatalf("create workspace manager: %v", err)
	}

	workspace, err := manager.CreateForIssue("TASK-2")
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	if _, err := os.Stat(filepath.Join(workspace.Path, "docs", "README.md")); err != nil {
		t.Fatalf("workspace was not populated: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workspace.Path, ".git", "config")); !os.IsNotExist(err) {
		t.Fatalf(".git should not be copied, err=%v", err)
	}
	if err := manager.RemoveForIssue("TASK-2"); err != nil {
		t.Fatalf("remove workspace: %v", err)
	}
	if _, err := os.Stat(workspace.Path); !os.IsNotExist(err) {
		t.Fatalf("workspace should be removed, err=%v", err)
	}
}
