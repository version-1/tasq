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
