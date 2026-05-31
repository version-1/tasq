package workspace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestManagerRunsAfterCreateHookForNewWorkspace(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	manager, err := NewManagerWithSourceAndHooks(root, "", HookConfig{
		AfterCreate: `echo created > after-create.out`,
		Timeout:     time.Second,
	})
	if err != nil {
		t.Fatalf("create workspace manager: %v", err)
	}

	workspace, err := manager.CreateForIssue("TASK-1")
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(workspace.Path, "after-create.out"))
	if err != nil {
		t.Fatalf("read hook output: %v", err)
	}
	if strings.TrimSpace(string(content)) != "created" {
		t.Fatalf("hook output = %q", content)
	}
}

func TestManagerSkipsAfterCreateHookForExistingWorkspace(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	manager, err := NewManagerWithSourceAndHooks(root, "", HookConfig{
		AfterCreate: `echo created >> after-create.out`,
		Timeout:     time.Second,
	})
	if err != nil {
		t.Fatalf("create workspace manager: %v", err)
	}
	workspace, err := manager.CreateForIssue("TASK-1")
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if _, err := manager.CreateForIssue("TASK-1"); err != nil {
		t.Fatalf("reuse workspace: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(workspace.Path, "after-create.out"))
	if err != nil {
		t.Fatalf("read hook output: %v", err)
	}
	if string(content) != "created\n" {
		t.Fatalf("hook output = %q", content)
	}
}

func TestManagerAfterCreateHookFailureRemovesPartialWorkspace(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	manager, err := NewManagerWithSourceAndHooks(root, "", HookConfig{
		AfterCreate: `exit 9`,
		Timeout:     time.Second,
	})
	if err != nil {
		t.Fatalf("create workspace manager: %v", err)
	}

	if _, err := manager.CreateForIssue("TASK-1"); err == nil {
		t.Fatal("expected create error")
	}
	if _, err := os.Stat(filepath.Join(root, "TASK-1")); !os.IsNotExist(err) {
		t.Fatalf("partial workspace should be removed, err=%v", err)
	}
}

func TestManagerRunsBeforeRemoveAndContinuesOnFailure(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	manager, err := NewManagerWithSourceAndHooks(root, "", HookConfig{
		BeforeRemove: `echo removing > ../before-remove.out; exit 8`,
		Timeout:      time.Second,
	})
	if err != nil {
		t.Fatalf("create workspace manager: %v", err)
	}
	workspace, err := manager.CreateForIssue("TASK-1")
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	if err := manager.RemoveForIssue("TASK-1"); err != nil {
		t.Fatalf("remove workspace: %v", err)
	}

	if _, err := os.Stat(workspace.Path); !os.IsNotExist(err) {
		t.Fatalf("workspace should be removed, err=%v", err)
	}
	content, err := os.ReadFile(filepath.Join(root, "before-remove.out"))
	if err != nil {
		t.Fatalf("read hook output: %v", err)
	}
	if strings.TrimSpace(string(content)) != "removing" {
		t.Fatalf("hook output = %q", content)
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
