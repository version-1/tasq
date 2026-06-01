package tq

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"gopkg.in/yaml.v3"
)

const workflowFileName = "WORKFLOW.md"

type projectAddResult struct {
	Project   entity.Project   `json:"project"`
	Workspace entity.Workspace `json:"workspace"`
}

type projectCheckResult struct {
	Passed bool   `json:"passed"`
	Reason string `json:"reason"`
}

type projectCheckItem struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
	Reason string `json:"reason"`
}

func (a app) projectList(ctx context.Context, args []string, cfg config) error {
	if len(args) > 0 {
		return usageError("project list does not accept positional arguments")
	}
	projects, err := a.client.listProjects(ctx)
	if err != nil {
		return err
	}
	return writeProjects(a.stdout, cfg.output, projects)
}

func (a app) projectRemove(ctx context.Context, args []string, cfg config) error {
	if len(args) != 1 {
		return usageError("usage: tq project remove <project-key>")
	}
	project, err := a.projectByKey(ctx, args[0])
	if err != nil {
		return err
	}
	if err := a.client.deleteProject(ctx, project.ID); err != nil {
		return err
	}
	if cfg.output == "json" {
		return writeJSON(a.stdout, map[string]any{"removed": true, "project": project})
	}
	fmt.Fprintf(a.stdout, "Removed project %s\n", project.Key)
	return nil
}

func (a app) projectAdd(ctx context.Context, args []string, cfg config) error {
	fs := newFlagSet("project add")
	key := fs.String("key", "", "project key")
	workspaceName := fs.String("workspace-name", "", "workspace name")
	if err := fs.Parse(args); err != nil {
		return usageError(err.Error())
	}
	if fs.NArg() > 1 {
		return usageError("usage: tq project add [flags] [path]")
	}

	root := "."
	if fs.NArg() == 1 {
		root = fs.Arg(0)
	}
	projectRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	projectRoot = filepath.Clean(projectRoot)
	info, err := os.Stat(projectRoot)
	if err != nil {
		return fmt.Errorf("project path is invalid: %w", err)
	}
	if !info.IsDir() {
		return errors.New("project path must be a directory")
	}

	name := filepath.Base(projectRoot)
	if *key == "" {
		*key = kebabKey(name)
	}
	if *workspaceName == "" {
		*workspaceName = name
	}
	if *key == "" {
		return usageError("key is required")
	}
	if err := a.ensureProjectAddDoesNotDuplicate(ctx, *key, projectRoot); err != nil {
		return err
	}

	project, err := a.client.createProject(ctx, entity.CreateProjectInput{
		Key:      *key,
		Name:     name,
		Location: projectRoot,
	})
	if err != nil {
		return err
	}
	rollbackAPI := true
	defer func() {
		if rollbackAPI {
			_ = a.client.deleteProject(context.Background(), project.ID)
		}
	}()

	workspace, err := a.client.createWorkspace(ctx, entity.CreateWorkspaceInput{
		ProjectID: project.ID,
		Name:      *workspaceName,
		Path:      projectRoot,
	})
	if err != nil {
		return err
	}

	local, err := prepareProjectFiles(projectRoot)
	if err != nil {
		local.rollback()
		return err
	}
	rollbackAPI = false
	return writeProjectAddResult(a.stdout, cfg.output, projectAddResult{Project: project, Workspace: workspace})
}

func (a app) projectCheck(ctx context.Context, args []string, cfg config) error {
	if len(args) > 1 {
		return usageError("usage: tq project check [project-key]")
	}
	project, err := a.projectForCheck(ctx, args)
	if err != nil {
		return err
	}

	workflowPath := filepath.Join(project.Location, workflowFileName)
	workflowContent, readErr := os.ReadFile(workflowPath)
	items := []projectCheckItem{
		checkWorkflowExists(readErr),
	}
	if readErr == nil {
		items = append(items, checkWorkflowFrontMatter(workflowContent))
		apiResult, err := a.client.checkProject(ctx, project.ID, string(workflowContent))
		if err != nil {
			return err
		}
		items = append(items, projectCheckItem{Name: "api.tq_usage", Passed: apiResult.Passed, Reason: apiResult.Reason})
	} else {
		items = append(items, projectCheckItem{Name: "workflow.front_matter", Passed: false, Reason: "WORKFLOW.md is missing"})
		items = append(items, projectCheckItem{Name: "api.tq_usage", Passed: false, Reason: "WORKFLOW.md is missing"})
	}
	items = append(items, checkAgentsFile(project.Location))

	if err := writeProjectCheckItems(a.stdout, cfg.output, items); err != nil {
		return err
	}
	for _, item := range items {
		if !item.Passed {
			return cliError{message: "project check failed", code: 1}
		}
	}
	return nil
}

func (a app) projectForCheck(ctx context.Context, args []string) (entity.Project, error) {
	if len(args) == 1 {
		return a.projectByKey(ctx, args[0])
	}
	cwd, err := os.Getwd()
	if err != nil {
		return entity.Project{}, err
	}
	cwd, err = filepath.Abs(cwd)
	if err != nil {
		return entity.Project{}, err
	}
	projects, err := a.client.listProjects(ctx)
	if err != nil {
		return entity.Project{}, err
	}
	for _, project := range projects {
		if samePath(project.Location, cwd) {
			return project, nil
		}
	}
	return entity.Project{}, cliError{message: "project not found for current directory", code: 1}
}

func (a app) projectByKey(ctx context.Context, key string) (entity.Project, error) {
	projects, err := a.client.listProjects(ctx)
	if err != nil {
		return entity.Project{}, err
	}
	for _, project := range projects {
		if project.Key == key {
			return project, nil
		}
	}
	return entity.Project{}, cliError{message: "project not found", code: 1}
}

func (a app) ensureProjectAddDoesNotDuplicate(ctx context.Context, key string, location string) error {
	projects, err := a.client.listProjects(ctx)
	if err != nil {
		return err
	}
	for _, project := range projects {
		if project.Key == key {
			return cliError{message: "project key already exists", code: 1}
		}
		if samePath(project.Location, location) {
			return cliError{message: "project path already exists", code: 1}
		}
	}
	workspaces, err := a.client.listWorkspaces(ctx)
	if err != nil {
		return err
	}
	for _, workspace := range workspaces {
		if samePath(workspace.Path, location) {
			return cliError{message: "workspace path already exists", code: 1}
		}
	}
	return nil
}

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
		if err := os.WriteFile(workflowPath, []byte(defaultWorkflowTemplate()), 0o644); err != nil {
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

var invalidKeyChars = regexp.MustCompile(`[^a-z0-9]+`)

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

func checkWorkflowExists(err error) projectCheckItem {
	if err == nil {
		return projectCheckItem{Name: "workflow.exists", Passed: true, Reason: "WORKFLOW.md exists"}
	}
	return projectCheckItem{Name: "workflow.exists", Passed: false, Reason: "WORKFLOW.md is missing"}
}

func checkWorkflowFrontMatter(content []byte) projectCheckItem {
	frontMatter, ok := workflowFrontMatter(string(content))
	if !ok {
		return projectCheckItem{Name: "workflow.front_matter", Passed: false, Reason: "WORKFLOW.md front matter is missing"}
	}
	var raw map[string]any
	if err := yaml.Unmarshal([]byte(frontMatter), &raw); err != nil {
		return projectCheckItem{Name: "workflow.front_matter", Passed: false, Reason: err.Error()}
	}
	missing := missingWorkflowFields(raw)
	if len(missing) > 0 {
		return projectCheckItem{Name: "workflow.front_matter", Passed: false, Reason: "missing fields: " + strings.Join(missing, ", ")}
	}
	return projectCheckItem{Name: "workflow.front_matter", Passed: true, Reason: "required front matter fields are present"}
}

func workflowFrontMatter(content string) (string, bool) {
	if !strings.HasPrefix(content, "---\n") {
		return "", false
	}
	rest := strings.TrimPrefix(content, "---\n")
	frontMatter, _, ok := strings.Cut(rest, "\n---\n")
	return frontMatter, ok
}

func missingWorkflowFields(raw map[string]any) []string {
	required := []string{
		"polling.interval_ms",
		"workspace.root",
		"workspace.source",
		"agent.max_concurrent_agents",
		"agent.max_turns",
		"agent.continuation_turns_enabled",
		"agent.max_retry_attempts",
		"agent.max_retry_backoff_ms",
		"codex.command",
		"codex.read_timeout_ms",
		"codex.turn_timeout_ms",
		"codex.stall_timeout_ms",
	}
	missing := []string{}
	for _, field := range required {
		if !hasNestedField(raw, strings.Split(field, ".")) {
			missing = append(missing, field)
		}
	}
	return missing
}

func hasNestedField(raw map[string]any, path []string) bool {
	current := any(raw)
	for _, key := range path {
		asMap, ok := current.(map[string]any)
		if !ok {
			return false
		}
		next, ok := asMap[key]
		if !ok {
			return false
		}
		current = next
	}
	return true
}

func checkAgentsFile(root string) projectCheckItem {
	content, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	if err != nil {
		return projectCheckItem{Name: "agents.references_workflow", Passed: false, Reason: "AGENTS.md is missing"}
	}
	if !strings.Contains(string(content), workflowFileName) {
		return projectCheckItem{Name: "agents.references_workflow", Passed: false, Reason: "AGENTS.md does not reference WORKFLOW.md"}
	}
	return projectCheckItem{Name: "agents.references_workflow", Passed: true, Reason: "AGENTS.md references WORKFLOW.md"}
}

func defaultWorkflowTemplate() string {
	return `---
polling:
  interval_ms: 30000
workspace:
  root: .worktrees
  source: .
agent:
  max_concurrent_agents: 1
  max_turns: 20
  continuation_turns_enabled: false
  max_retry_attempts: 3
  max_retry_backoff_ms: 300000
codex:
  command: codex app-server
  read_timeout_ms: 5000
  turn_timeout_ms: 3600000
  stall_timeout_ms: 300000
---

Use tq commands to keep issue tracker state synchronized.

When starting work:

  tq issue update {{ issue.id }} --status in_progress

When work is ready for review:

  tq issue update {{ issue.id }} --status review

Leave progress or handoff notes with:

  tq comment add {{ issue.id }} --type progress --body "Progress note"
`
}
