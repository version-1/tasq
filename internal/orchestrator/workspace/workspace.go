package workspace

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Manager struct {
	root     string
	repoRoot string
	hooks    HookConfig
}

type Workspace struct {
	Path         string
	WorkspaceKey string
	CreatedNow   bool
}

func NewManager(root string) (*Manager, error) {
	return NewManagerWithHooks(root, HookConfig{})
}

func NewManagerWithHooks(root string, hooks HookConfig) (*Manager, error) {
	if root == "" {
		return nil, fmt.Errorf("workspace root is required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace root: %w", err)
	}
	cleanRoot := filepath.Clean(abs)
	if err := os.MkdirAll(cleanRoot, 0o755); err != nil {
		return nil, fmt.Errorf("create workspace root: %w", err)
	}
	manager := &Manager{root: cleanRoot, hooks: hooks}
	repoRoot, err := manager.gitExec("rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("resolve workspace git repository: %w", err)
	}
	manager.repoRoot = filepath.Clean(strings.TrimSpace(repoRoot))
	return manager, nil
}

func (m *Manager) Root() string {
	return m.root
}

func (m *Manager) RepoRoot() string {
	return m.repoRoot
}

func (m *Manager) CreateForIssue(identifier string) (Workspace, error) {
	key := sanitizeKey(identifier)
	if key == "" {
		return Workspace{}, fmt.Errorf("workspace key is empty")
	}
	path := filepath.Join(m.root, key)
	if err := m.validatePath(path); err != nil {
		return Workspace{}, err
	}
	createdNow := false
	info, err := os.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return Workspace{}, fmt.Errorf("workspace path is not a directory: %s", path)
		}
		return Workspace{Path: path, WorkspaceKey: key, CreatedNow: createdNow}, nil
	}
	if !os.IsNotExist(err) {
		return Workspace{}, fmt.Errorf("stat workspace path: %w", err)
	}
	branch := workspaceBranch(key)
	createdBranch := true
	if _, err := m.gitExec("worktree", "add", "-b", branch, path); err != nil {
		createdBranch = false
		if _, retryErr := m.gitExec("worktree", "add", path, branch); retryErr != nil {
			return Workspace{}, fmt.Errorf("create workspace worktree: %w; retry existing branch: %v", err, retryErr)
		}
	}
	createdNow = true
	if err := RunHook(m.hooks.AfterCreate, path, m.hooks.timeout()); err != nil {
		if removeErr := m.removeWorktree(path); removeErr != nil {
			return Workspace{}, fmt.Errorf("after_create hook failed: %w; remove partial workspace: %v", err, removeErr)
		}
		if createdBranch {
			m.deleteBranchBestEffort(branch)
		}
		return Workspace{}, fmt.Errorf("after_create hook failed: %w", err)
	}
	return Workspace{Path: path, WorkspaceKey: key, CreatedNow: createdNow}, nil
}

func (m *Manager) RemoveForIssue(identifier string) error {
	key := sanitizeKey(identifier)
	if key == "" {
		return fmt.Errorf("workspace key is empty")
	}
	path := filepath.Join(m.root, key)
	if err := m.validatePath(path); err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		if hookErr := RunHook(m.hooks.BeforeRemove, path, m.hooks.timeout()); hookErr != nil {
			log.Printf("before_remove hook failed workspace=%s: %v", path, hookErr)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat workspace path: %w", err)
	}
	if err := m.removeWorktree(path); err != nil {
		return fmt.Errorf("remove workspace: %w", err)
	}
	m.deleteBranchBestEffort(workspaceBranch(key))
	return nil
}

func (m *Manager) Prune() error {
	if _, err := m.gitExec("worktree", "prune"); err != nil {
		return fmt.Errorf("prune workspace worktrees: %w", err)
	}
	return nil
}

func (m *Manager) validatePath(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve workspace path: %w", err)
	}
	rel, err := filepath.Rel(m.root, abs)
	if err != nil {
		return fmt.Errorf("compare workspace path: %w", err)
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("workspace path escapes root: %s", abs)
	}
	return nil
}

func (m *Manager) removeWorktree(path string) error {
	if _, err := m.gitExec("worktree", "remove", "--force", path); err == nil {
		return nil
	}
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return nil
}

func (m *Manager) deleteBranchBestEffort(branch string) {
	if _, err := m.gitExec("branch", "-D", branch); err != nil {
		log.Printf("delete workspace branch failed branch=%s: %v", branch, err)
	}
}

func (m *Manager) gitExec(args ...string) (string, error) {
	commandArgs := append([]string{"-C", m.gitRootForCommand()}, args...)
	cmd := exec.Command("git", commandArgs...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		return stdout.String(), fmt.Errorf("git %s: %s", strings.Join(args, " "), message)
	}
	return stdout.String(), nil
}

func (m *Manager) gitRootForCommand() string {
	if m.repoRoot != "" {
		return m.repoRoot
	}
	return m.root
}

func workspaceBranch(key string) string {
	return "agent/" + key
}

func sanitizeKey(identifier string) string {
	var builder strings.Builder
	for _, r := range identifier {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '-' {
			builder.WriteRune(r)
			continue
		}
		builder.WriteByte('_')
	}
	return builder.String()
}
