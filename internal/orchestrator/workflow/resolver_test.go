package workflow

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func TestResolverUpdatesProjectWorkflowWhenFileChecksumDiffers(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	content := "---\nagent:\n  max_turns: 3\n---\nProject prompt"
	writeWorkflow(t, projectRoot, content)
	client := newFakeWorkflowClient()
	resolver := NewResolver(client)

	definition, err := resolver.Resolve(context.Background(), entity.Project{
		ID:               7,
		Location:         projectRoot,
		WorkflowChecksum: "old",
	})
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if definition.PromptTemplate != "Project prompt" || definition.Config.MaxTurns != 3 {
		t.Fatalf("definition = %+v", definition)
	}
	if client.updateCount != 1 {
		t.Fatalf("update count = %d", client.updateCount)
	}
	if client.workflowCount != 0 {
		t.Fatalf("workflow count = %d", client.workflowCount)
	}

	if _, err := resolver.Resolve(context.Background(), entity.Project{
		ID:               7,
		Location:         projectRoot,
		WorkflowChecksum: checksum([]byte(content)),
	}); err != nil {
		t.Fatalf("resolve cached: %v", err)
	}
	if client.workflowCount != 0 {
		t.Fatalf("workflow count after cache hit = %d", client.workflowCount)
	}
}

func TestResolverInvalidatesProjectCacheOnChecksumChange(t *testing.T) {
	t.Parallel()

	projectRoot := t.TempDir()
	client := newFakeWorkflowClient()
	first := testProjectWorkflow("First prompt")
	second := testProjectWorkflow("Second prompt")
	client.workflows[9] = first
	resolver := NewResolver(client)

	definition, err := resolver.Resolve(context.Background(), entity.Project{ID: 9, Location: projectRoot, WorkflowChecksum: first.Checksum})
	if err != nil {
		t.Fatalf("resolve first: %v", err)
	}
	if definition.PromptTemplate != "First prompt" {
		t.Fatalf("first prompt = %q", definition.PromptTemplate)
	}
	client.workflows[9] = second
	definition, err = resolver.Resolve(context.Background(), entity.Project{ID: 9, Location: projectRoot, WorkflowChecksum: second.Checksum})
	if err != nil {
		t.Fatalf("resolve second: %v", err)
	}
	if definition.PromptTemplate != "Second prompt" {
		t.Fatalf("second prompt = %q", definition.PromptTemplate)
	}
	if client.workflowCount != 2 {
		t.Fatalf("workflow count = %d", client.workflowCount)
	}
}

func TestResolverFallsBackToGlobalWorkflowAndCachesByChecksum(t *testing.T) {
	home := t.TempDir()
	t.Setenv("TQ_HOME", home)
	content := "Global prompt"
	writeWorkflow(t, home, content)
	client := newFakeWorkflowClient()
	resolver := NewResolver(client)

	definition, err := resolver.Resolve(context.Background(), entity.Project{ID: 11, Location: t.TempDir()})
	if err != nil {
		t.Fatalf("resolve global: %v", err)
	}
	if definition.PromptTemplate != "Global prompt" {
		t.Fatalf("global prompt = %q", definition.PromptTemplate)
	}

	content = "Updated global prompt"
	writeWorkflow(t, home, content)
	definition, err = resolver.Resolve(context.Background(), entity.Project{ID: 11, Location: t.TempDir()})
	if err != nil {
		t.Fatalf("resolve updated global: %v", err)
	}
	if definition.PromptTemplate != "Updated global prompt" {
		t.Fatalf("updated global prompt = %q", definition.PromptTemplate)
	}
}

type fakeWorkflowClient struct {
	workflows     map[int64]entity.ProjectWorkflow
	updateCount   int
	workflowCount int
}

func newFakeWorkflowClient() *fakeWorkflowClient {
	return &fakeWorkflowClient{workflows: map[int64]entity.ProjectWorkflow{}}
}

func (c *fakeWorkflowClient) Workflow(ctx context.Context, projectID int64) (entity.ProjectWorkflow, bool, error) {
	c.workflowCount++
	workflow, ok := c.workflows[projectID]
	if !ok {
		return entity.ProjectWorkflow{}, false, nil
	}
	return workflow, true, nil
}

func (c *fakeWorkflowClient) UpsertWorkflow(ctx context.Context, projectID int64, input entity.UpsertProjectWorkflowInput) (entity.ProjectWorkflow, error) {
	c.updateCount++
	frontmatter := map[string]any{}
	if err := json.Unmarshal([]byte(input.FrontmatterJSON), &frontmatter); err != nil {
		return entity.ProjectWorkflow{}, err
	}
	workflow := entity.ProjectWorkflow{
		ProjectID:   projectID,
		Frontmatter: frontmatter,
		Body:        input.Body,
		Checksum:    input.Checksum,
	}
	c.workflows[projectID] = workflow
	return workflow, nil
}

func testProjectWorkflow(body string) entity.ProjectWorkflow {
	return entity.ProjectWorkflow{
		Frontmatter: map[string]any{},
		Body:        body,
		Checksum:    checksum([]byte(body)),
	}
}

func writeWorkflow(t *testing.T, dir string, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, workflowFileName), []byte(content), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}
}
