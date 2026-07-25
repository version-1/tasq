package tq

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	tqconfig "github.com/version-1/tasq/internal/config"
)

type localProjectFiles struct {
	workflowPath    string
	workflowCreated bool
	gitignore       fileSnapshot
}

type fileSnapshot struct {
	path    string
	exists  bool
	content []byte
	mode    os.FileMode
}

func prepareProjectFiles(root string) (localProjectFiles, error) {
	local := localProjectFiles{}
	workflowPath := filepath.Join(root, workflowFileName)
	local.workflowPath = workflowPath
	if _, err := os.Stat(workflowPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return local, err
		}
		if err := os.WriteFile(workflowPath, []byte(tqconfig.DefaultWorkflowTemplate()), 0o644); err != nil {
			return local, err
		}
		local.workflowCreated = true
	}

	gitignorePath := filepath.Join(root, ".gitignore")
	snapshot, err := snapshotFile(gitignorePath)
	if err != nil {
		return local, err
	}
	local.gitignore = snapshot
	if !gitignoreContains(snapshot.content, ".worktrees") {
		next := append([]byte{}, snapshot.content...)
		if len(next) > 0 && next[len(next)-1] != '\n' {
			next = append(next, '\n')
		}
		next = append(next, []byte(".worktrees\n")...)
		if err := os.WriteFile(gitignorePath, next, 0o644); err != nil {
			return local, err
		}
	}
	return local, nil
}

func (l localProjectFiles) rollback() {
	if l.workflowCreated {
		_ = os.Remove(l.workflowPath)
	}
	if l.gitignore.path == "" {
		return
	}
	if l.gitignore.exists {
		_ = os.WriteFile(l.gitignore.path, l.gitignore.content, l.gitignore.mode)
		return
	}
	_ = os.Remove(l.gitignore.path)
}

func snapshotFile(path string) (fileSnapshot, error) {
	snapshot := fileSnapshot{path: path, mode: 0o644}
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return snapshot, nil
		}
		return snapshot, err
	}
	if info.IsDir() {
		return snapshot, errors.New(".gitignore must be a file")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return snapshot, err
	}
	snapshot.exists = true
	snapshot.content = content
	snapshot.mode = info.Mode().Perm()
	return snapshot, nil
}

func gitignoreContains(content []byte, entry string) bool {
	for _, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) == entry {
			return true
		}
	}
	return false
}

func kebabKey(value string) string {
	key := strings.ToLower(strings.TrimSpace(value))
	key = invalidKeyChars.ReplaceAllString(key, "-")
	key = strings.Trim(key, "-")
	for strings.Contains(key, "--") {
		key = strings.ReplaceAll(key, "--", "-")
	}
	if len(key) > 64 {
		key = strings.TrimRight(key[:64], "-")
	}
	return key
}

func samePath(left, right string) bool {
	return filepath.Clean(left) == filepath.Clean(right)
}
