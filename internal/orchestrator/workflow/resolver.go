package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

const workflowFileName = "WORKFLOW.md"

type ProjectWorkflowClient interface {
	Project(ctx context.Context, id int64) (entity.Project, error)
	Workflow(ctx context.Context, projectID int64) (entity.ProjectWorkflow, bool, error)
}

type Resolver struct {
	client   ProjectWorkflowClient
	fallback Definition
}

func NewResolver(client ProjectWorkflowClient, fallback Definition) *Resolver {
	return &Resolver{client: client, fallback: fallback}
}

func (r *Resolver) Resolve(ctx context.Context, projectID int64) (Definition, error) {
	if r.client == nil {
		return Definition{}, fmt.Errorf("workflow resolver client is required")
	}
	project, err := r.client.Project(ctx, projectID)
	if err != nil {
		return Definition{}, fmt.Errorf("fetch project %d: %w", projectID, err)
	}
	if project.Location != "" {
		path := filepath.Join(project.Location, workflowFileName)
		if _, err := os.Stat(path); err == nil {
			definition, err := Load(path)
			if err != nil {
				return Definition{}, fmt.Errorf("load project workflow %s: %w", path, err)
			}
			return definition, nil
		} else if !os.IsNotExist(err) {
			return Definition{}, fmt.Errorf("stat project workflow %s: %w", path, err)
		}
	}
	stored, found, err := r.client.Workflow(ctx, projectID)
	if err != nil {
		return Definition{}, fmt.Errorf("fetch stored workflow for project %d: %w", projectID, err)
	}
	if found {
		return parseStoredWorkflow(project, stored)
	}
	return r.fallback, nil
}

func parseStoredWorkflow(project entity.Project, stored entity.ProjectWorkflow) (Definition, error) {
	workflowDir := project.Location
	if workflowDir == "" {
		workflowDir = "."
	}
	path := filepath.Join(workflowDir, workflowFileName)
	raw := []byte(stored.Body)
	if len(stored.Frontmatter) > 0 {
		frontmatter, err := json.Marshal(stored.Frontmatter)
		if err != nil {
			return Definition{}, fmt.Errorf("encode stored workflow frontmatter for project %d: %w", project.ID, err)
		}
		raw = []byte("---\n" + string(frontmatter) + "\n---\n" + stored.Body)
	}
	definition, err := ParseDefinition(path, raw, workflowDir)
	if err != nil {
		return Definition{}, fmt.Errorf("parse stored workflow for project %d: %w", project.ID, err)
	}
	return definition, nil
}
