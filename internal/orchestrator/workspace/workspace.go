package workspace

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Manager struct {
	root   string
	source string
}

type Workspace struct {
	Path         string
	WorkspaceKey string
	CreatedNow   bool
}

func NewManager(root string) (*Manager, error) {
	return NewManagerWithSource(root, "")
}

func NewManagerWithSource(root string, source string) (*Manager, error) {
	if root == "" {
		return nil, fmt.Errorf("workspace root is required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace root: %w", err)
	}
	manager := &Manager{root: filepath.Clean(abs)}
	if source != "" {
		sourceAbs, err := filepath.Abs(source)
		if err != nil {
			return nil, fmt.Errorf("resolve workspace source: %w", err)
		}
		manager.source = filepath.Clean(sourceAbs)
	}
	return manager, nil
}

func (m *Manager) Root() string {
	return m.root
}

func (m *Manager) Source() string {
	return m.source
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
	if err := os.MkdirAll(path, 0o755); err != nil {
		return Workspace{}, fmt.Errorf("create workspace: %w", err)
	}
	createdNow = true
	if m.source != "" {
		if err := m.populate(path); err != nil {
			return Workspace{}, err
		}
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
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("remove workspace: %w", err)
	}
	return nil
}

func (m *Manager) populate(destination string) error {
	info, err := os.Stat(m.source)
	if err != nil {
		return fmt.Errorf("stat workspace source: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("workspace source is not a directory: %s", m.source)
	}
	return filepath.WalkDir(m.source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(m.source, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if shouldSkip(rel, path, m.root) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		target := filepath.Join(destination, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		return copyFile(path, target)
	})
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

func shouldSkip(rel string, path string, root string) bool {
	name := strings.Split(rel, string(filepath.Separator))[0]
	if name == ".git" || name == ".worktrees" || name == "node_modules" {
		return true
	}
	relativeToRoot, err := filepath.Rel(root, path)
	return err == nil && (relativeToRoot == "." || !strings.HasPrefix(relativeToRoot, ".."+string(filepath.Separator)))
}

func copyFile(source string, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	info, err := input.Stat()
	if err != nil {
		return err
	}
	output, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer output.Close()
	_, err = io.Copy(output, input)
	return err
}
