package workflow

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func TestResolverLoadsPhysicalProjectWorkflowFirst(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeWorkflow(t, filepath.Join(dir, workflowFileName), `---
agent:
  max_turns: 4
codex:
  command: codex app-server --physical
  read_timeout_ms: 2000
  turn_timeout_ms: 3000
---
Physical workflow`)
	client := &fakeWorkflowClient{
		project: entity.Project{ID: 7, Location: dir},
		workflow: entity.ProjectWorkflow{
			Frontmatter: map[string]any{
				"agent": map[string]any{"max_turns": 9},
				"codex": map[string]any{
					"command":         "codex app-server --stored",
					"read_timeout_ms": 9000,
					"turn_timeout_ms": 10000,
				},
			},
			Body: "Stored workflow",
		},
		workflowFound: true,
	}
	definition, err := NewResolver(client, fallbackDefinition()).Resolve(context.Background(), 7)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if definition.PromptTemplate != "Physical workflow" {
		t.Fatalf("prompt = %q", definition.PromptTemplate)
	}
	if definition.Config.MaxTurns != 4 || definition.Config.CodexCommand != "codex app-server --physical" {
		t.Fatalf("config = %+v", definition.Config)
	}
}

func TestResolverLoadsStoredProjectWorkflowWhenPhysicalFileMissing(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	falseTaskWorkPrompt := false
	client := &fakeWorkflowClient{
		project: entity.Project{ID: 7, Location: dir},
		workflow: entity.ProjectWorkflow{
			Frontmatter: map[string]any{
				"tasq":      map[string]any{"task_work_prompt": false},
				"workspace": map[string]any{"root": ".worktrees"},
				"agent":     map[string]any{"max_turns": 6},
				"codex": map[string]any{
					"command":         "codex app-server --stored",
					"read_timeout_ms": 2000,
					"turn_timeout_ms": 3000,
				},
			},
			Body: "Stored workflow",
		},
		workflowFound: true,
	}
	definition, err := NewResolver(client, fallbackDefinition()).Resolve(context.Background(), 7)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if definition.PromptTemplate != "Stored workflow" {
		t.Fatalf("prompt = %q", definition.PromptTemplate)
	}
	if definition.Config.MaxTurns != 6 || definition.Config.CodexCommand != "codex app-server --stored" {
		t.Fatalf("config = %+v", definition.Config)
	}
	if definition.Config.Tasq.TaskWorkPrompt == nil || *definition.Config.Tasq.TaskWorkPrompt != falseTaskWorkPrompt {
		t.Fatalf("task work prompt = %v", definition.Config.Tasq.TaskWorkPrompt)
	}
	if definition.Config.WorkspaceRoot != filepath.Join(dir, ".worktrees") {
		t.Fatalf("workspace root = %q", definition.Config.WorkspaceRoot)
	}
}

func TestResolverFallsBackWhenProjectHasNoWorkflow(t *testing.T) {
	t.Parallel()

	fallback := fallbackDefinition()
	client := &fakeWorkflowClient{
		project: entity.Project{ID: 7, Location: t.TempDir()},
	}
	definition, err := NewResolver(client, fallback).Resolve(context.Background(), 7)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}

	if definition.PromptTemplate != fallback.PromptTemplate || definition.Config.CodexCommand != fallback.Config.CodexCommand {
		t.Fatalf("definition = %+v, want fallback %+v", definition, fallback)
	}
}

func fallbackDefinition() Definition {
	return Definition{
		Path: "fallback",
		Config: Config{
			WorkspaceRoot:    os.TempDir(),
			MaxTurns:         2,
			CodexCommand:     "codex app-server --fallback",
			CodexReadTimeout: time.Second,
			CodexTurnTimeout: time.Second,
		},
		PromptTemplate: "Fallback workflow",
	}
}

func writeWorkflow(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}
}

type fakeWorkflowClient struct {
	project       entity.Project
	workflow      entity.ProjectWorkflow
	workflowFound bool
}

func (c *fakeWorkflowClient) Project(ctx context.Context, id int64) (entity.Project, error) {
	if c.project.ID == id {
		return c.project, nil
	}
	return entity.Project{}, sql.ErrNoRows
}

func (c *fakeWorkflowClient) Workflow(ctx context.Context, projectID int64) (entity.ProjectWorkflow, bool, error) {
	return c.workflow, c.workflowFound, nil
}
