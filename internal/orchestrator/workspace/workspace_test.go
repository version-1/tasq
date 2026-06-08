package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestManagerCreatesSanitizedWorkspaceUnderRoot(t *testing.T) {
	t.Parallel()

	_, root := initTestRepo(t)
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
	if filepath.Dir(workspace.Path) != manager.Root() {
		t.Fatalf("workspace path = %q, root = %q", workspace.Path, manager.Root())
	}
}

func TestManagerCreatesWorkspaceUnderProjectLocation(t *testing.T) {
	t.Parallel()

	_, root := initTestRepo(t)
	manager, err := NewManager(root)
	if err != nil {
		t.Fatalf("create workspace manager: %v", err)
	}
	projectRoot, _ := initTestRepo(t)

	workspace, err := manager.CreateForIssueInProjectLocation("TASK-11", projectRoot)
	if err != nil {
		t.Fatalf("create project workspace: %v", err)
	}

	wantPath := filepath.Join(canonicalTestPath(t, projectRoot), ".worktrees", "TASK-11")
	if workspace.Path != wantPath {
		t.Fatalf("workspace path = %q, want %q", workspace.Path, wantPath)
	}
	if _, err := os.Stat(filepath.Join(workspace.Path, ".git")); err != nil {
		t.Fatalf("worktree git file missing: %v", err)
	}
}

func TestManagerRunsAfterCreateHookForNewWorkspace(t *testing.T) {
	t.Parallel()

	_, root := initTestRepo(t)
	manager, err := NewManagerWithHooks(root, HookConfig{
		AfterCreate: `echo created > after-create.out`,
		Timeout:     hookTestTimeout,
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

	_, root := initTestRepo(t)
	manager, err := NewManagerWithHooks(root, HookConfig{
		AfterCreate: `echo created >> after-create.out`,
		Timeout:     hookTestTimeout,
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

	_, root := initTestRepo(t)
	manager, err := NewManagerWithHooks(root, HookConfig{
		AfterCreate: `exit 9`,
		Timeout:     hookTestTimeout,
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

	_, root := initTestRepo(t)
	manager, err := NewManagerWithHooks(root, HookConfig{
		BeforeRemove: `echo removing > ../before-remove.out; exit 8`,
		Timeout:      hookTestTimeout,
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

	_, root := initTestRepo(t)
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

func TestManagerCreatesWorktreeWithCorrectBranch(t *testing.T) {
	t.Parallel()

	repoRoot, root := initTestRepo(t)
	if err := os.WriteFile(filepath.Join(repoRoot, "docs.md"), []byte("hello"), 0o644); err != nil {
		t.Fatalf("write tracked file: %v", err)
	}
	gitCommand(t, repoRoot, "add", "docs.md")
	gitCommand(t, repoRoot, "commit", "-m", "add docs")
	manager, err := NewManager(root)
	if err != nil {
		t.Fatalf("create workspace manager: %v", err)
	}

	workspace, err := manager.CreateForIssue("TASK-2")
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	if _, err := os.Stat(filepath.Join(workspace.Path, ".git")); err != nil {
		t.Fatalf("worktree git file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(workspace.Path, "docs.md")); err != nil {
		t.Fatalf("tracked file missing from worktree: %v", err)
	}
	branch := strings.TrimSpace(gitCommand(t, workspace.Path, "branch", "--show-current"))
	if branch != "agent/TASK-2" {
		t.Fatalf("branch = %q", branch)
	}
	if err := manager.RemoveForIssue("TASK-2"); err != nil {
		t.Fatalf("remove workspace: %v", err)
	}
	if _, err := os.Stat(workspace.Path); !os.IsNotExist(err) {
		t.Fatalf("workspace should be removed, err=%v", err)
	}
}

func TestManagerReusesExistingBranchOnRetry(t *testing.T) {
	t.Parallel()

	repoRoot, root := initTestRepo(t)
	gitCommand(t, repoRoot, "branch", "agent/TASK-3")
	manager, err := NewManager(root)
	if err != nil {
		t.Fatalf("create workspace manager: %v", err)
	}

	workspace, err := manager.CreateForIssue("TASK-3")
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}

	branch := strings.TrimSpace(gitCommand(t, workspace.Path, "branch", "--show-current"))
	if branch != "agent/TASK-3" {
		t.Fatalf("branch = %q", branch)
	}
}

func TestManagerPrunesCleansStaleMetadata(t *testing.T) {
	t.Parallel()

	repoRoot, root := initTestRepo(t)
	manager, err := NewManager(root)
	if err != nil {
		t.Fatalf("create workspace manager: %v", err)
	}
	workspace, err := manager.CreateForIssue("TASK-4")
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if err := os.RemoveAll(workspace.Path); err != nil {
		t.Fatalf("remove workspace directory manually: %v", err)
	}

	if err := manager.Prune(); err != nil {
		t.Fatalf("prune worktrees: %v", err)
	}

	list := gitCommand(t, repoRoot, "worktree", "list", "--porcelain")
	if strings.Contains(list, workspace.Path) {
		t.Fatalf("stale worktree remains in list:\n%s", list)
	}
}

func initTestRepo(t *testing.T) (string, string) {
	t.Helper()
	repoRoot := t.TempDir()
	gitCommand(t, repoRoot, "init")
	gitCommand(t, repoRoot, "config", "user.email", "test@example.com")
	gitCommand(t, repoRoot, "config", "user.name", "Test User")
	if err := os.WriteFile(filepath.Join(repoRoot, "README.md"), []byte("test repo\n"), 0o644); err != nil {
		t.Fatalf("write readme: %v", err)
	}
	gitCommand(t, repoRoot, "add", "README.md")
	gitCommand(t, repoRoot, "commit", "-m", "initial commit")
	workspaceRoot := filepath.Join(repoRoot, ".worktrees")
	if err := os.MkdirAll(workspaceRoot, 0o755); err != nil {
		t.Fatalf("create workspace root: %v", err)
	}
	return repoRoot, workspaceRoot
}

func canonicalTestPath(t *testing.T, path string) string {
	t.Helper()
	cleanPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("canonicalize test path %q: %v", path, err)
	}
	return filepath.Clean(cleanPath)
}

func gitCommand(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	return string(output)
}
