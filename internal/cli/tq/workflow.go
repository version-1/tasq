package tq

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	tqconfig "github.com/version-1/tasq/internal/config"
	"github.com/version-1/tasq/internal/issue/domain/entity"
	"gopkg.in/yaml.v3"
)

type workflowShowResult struct {
	Project entity.Project     `json:"project"`
	Source  workflowShowSource `json:"source"`
	Content string             `json:"content"`
}

type workflowShowSource struct {
	Type      string `json:"type"`
	Path      string `json:"path,omitempty"`
	ProjectID int64  `json:"projectId,omitempty"`
}

func (a app) workflowShow(ctx context.Context, args []string, cfg config) error {
	fs := newFlagSet("workflow show")
	projectKey := fs.String("project", "", "project key")
	jsonOutput := fs.Bool("json", false, "output structured JSON")
	if err := fs.Parse(args); err != nil {
		return usageError(err.Error())
	}
	if fs.NArg() != 0 {
		return usageError("workflow show does not accept positional arguments")
	}
	if *projectKey == "" {
		return usageError("project is required")
	}

	project, err := a.projectByKey(ctx, *projectKey)
	if err != nil {
		return err
	}
	result, err := a.resolveWorkflow(ctx, project)
	if err != nil {
		return err
	}
	if *jsonOutput || cfg.output == "json" {
		return writeJSON(a.stdout, result)
	}
	return writeWorkflowShow(a.stdout, result)
}

func (a app) resolveWorkflow(ctx context.Context, project entity.Project) (workflowShowResult, error) {
	filePath := filepath.Join(project.Location, workflowFileName)
	content, err := os.ReadFile(filePath)
	if err == nil {
		return workflowShowResult{
			Project: project,
			Source:  workflowShowSource{Type: "file", Path: filePath},
			Content: string(content),
		}, nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return workflowShowResult{}, fmt.Errorf("read workflow file: %w", err)
	}

	dbWorkflow, found, err := a.client.projectWorkflow(ctx, project.ID)
	if err != nil {
		return workflowShowResult{}, err
	}
	if found {
		content, err := formatStoredWorkflow(dbWorkflow)
		if err != nil {
			return workflowShowResult{}, err
		}
		return workflowShowResult{
			Project: project,
			Source:  workflowShowSource{Type: "db", ProjectID: project.ID},
			Content: content,
		}, nil
	}

	home, err := tqconfig.Home()
	if err != nil {
		return workflowShowResult{}, err
	}
	globalPath := tqconfig.WorkflowPath(home)
	content, err = os.ReadFile(globalPath)
	if err == nil {
		return workflowShowResult{
			Project: project,
			Source:  workflowShowSource{Type: "global", Path: globalPath},
			Content: string(content),
		}, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return workflowShowResult{}, cliError{message: "workflow not found", code: 1}
	}
	return workflowShowResult{}, fmt.Errorf("read global workflow file: %w", err)
}

func formatStoredWorkflow(workflow entity.ProjectWorkflow) (string, error) {
	if len(workflow.Frontmatter) == 0 {
		return workflow.Body, nil
	}
	frontmatter, err := yaml.Marshal(workflow.Frontmatter)
	if err != nil {
		return "", fmt.Errorf("format workflow front matter: %w", err)
	}
	return "---\n" + string(frontmatter) + "---\n" + workflow.Body, nil
}

func writeWorkflowShow(w io.Writer, result workflowShowResult) error {
	switch result.Source.Type {
	case "file", "global":
		if _, err := fmt.Fprintf(w, "# Source: %s (%s)\n\n", result.Source.Type, result.Source.Path); err != nil {
			return err
		}
	case "db":
		if _, err := fmt.Fprintf(w, "# Source: db (project %s, id %d)\n\n", result.Project.Key, result.Source.ProjectID); err != nil {
			return err
		}
	default:
		if _, err := fmt.Fprintf(w, "# Source: %s\n\n", result.Source.Type); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, result.Content)
	return err
}
