package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Manager struct {
	root string
}

type Workspace struct {
	Path         string
	WorkspaceKey string
	CreatedNow   bool
}

func NewManager(root string) (*Manager, error) {
	if root == "" {
		return nil, fmt.Errorf("workspace root is required")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace root: %w", err)
	}
	return &Manager{root: filepath.Clean(abs)}, nil
}

func (m *Manager) Root() string {
	return m.root
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
	return Workspace{Path: path, WorkspaceKey: key, CreatedNow: createdNow}, nil
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
