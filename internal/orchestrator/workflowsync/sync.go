package workflowsync

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"gopkg.in/yaml.v3"
)

const workflowFilename = "WORKFLOW.md"

type Tracker interface {
	Projects(ctx context.Context) ([]entity.Project, error)
	Workflow(ctx context.Context, projectID int64) (entity.ProjectWorkflow, bool, error)
	UpsertWorkflow(ctx context.Context, projectID int64, input entity.UpsertProjectWorkflowInput) (entity.ProjectWorkflow, error)
}

type Result struct {
	Projects int
	Missing  int
	Skipped  int
	Updated  int
}

type LocalWorkflow struct {
	Frontmatter json.RawMessage
	Body        string
	Checksum    string
}

func SyncProjectWorkflows(ctx context.Context, client Tracker) (Result, error) {
	projects, err := client.Projects(ctx)
	if err != nil {
		return Result{}, fmt.Errorf("list projects: %w", err)
	}
	result := Result{Projects: len(projects)}
	for _, project := range projects {
		local, ok, err := ReadProjectWorkflow(project.Location)
		if err != nil {
			return result, fmt.Errorf("read workflow for project %s: %w", project.Key, err)
		}
		if !ok {
			result.Missing++
			continue
		}
		stored, found, err := client.Workflow(ctx, project.ID)
		if err != nil {
			return result, fmt.Errorf("read stored workflow for project %s: %w", project.Key, err)
		}
		if found && strings.EqualFold(stored.Checksum, local.Checksum) {
			result.Skipped++
			continue
		}
		if _, err := client.UpsertWorkflow(ctx, project.ID, entity.UpsertProjectWorkflowInput{
			ProjectID:       project.ID,
			FrontmatterJSON: string(local.Frontmatter),
			Body:            local.Body,
			Checksum:        local.Checksum,
		}); err != nil {
			return result, fmt.Errorf("upsert workflow for project %s: %w", project.Key, err)
		}
		result.Updated++
	}
	return result, nil
}

func ReadProjectWorkflow(projectLocation string) (LocalWorkflow, bool, error) {
	if strings.TrimSpace(projectLocation) == "" {
		return LocalWorkflow{}, false, nil
	}
	path := filepath.Join(projectLocation, workflowFilename)
	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return LocalWorkflow{}, false, nil
	}
	if err != nil {
		return LocalWorkflow{}, false, err
	}
	local, err := parseLocalWorkflow(raw)
	if err != nil {
		return LocalWorkflow{}, false, err
	}
	return local, true, nil
}

func parseLocalWorkflow(raw []byte) (LocalWorkflow, error) {
	frontmatter := json.RawMessage(`{}`)
	body := string(raw)
	text := string(raw)
	if strings.HasPrefix(text, "---\n") || strings.TrimSpace(text) == "---" {
		frontmatterText, parsedBody, ok := splitFrontmatter(strings.TrimPrefix(text, "---\n"))
		if !ok {
			return LocalWorkflow{}, errors.New("workflow_parse_error: unterminated front matter")
		}
		encoded, err := encodeFrontmatter(frontmatterText)
		if err != nil {
			return LocalWorkflow{}, err
		}
		frontmatter = encoded
		body = parsedBody
	}
	sum := sha256.Sum256(raw)
	return LocalWorkflow{
		Frontmatter: frontmatter,
		Body:        strings.TrimSpace(body),
		Checksum:    hex.EncodeToString(sum[:]),
	}, nil
}

func splitFrontmatter(rest string) (string, string, bool) {
	if strings.HasSuffix(rest, "\n---") {
		return strings.TrimSuffix(rest, "\n---"), "", true
	}
	frontmatter, body, ok := strings.Cut(rest, "\n---\n")
	return frontmatter, body, ok
}

func encodeFrontmatter(raw string) (json.RawMessage, error) {
	if strings.TrimSpace(raw) == "" {
		return json.RawMessage(`{}`), nil
	}
	var frontmatter map[string]any
	if err := yaml.Unmarshal([]byte(raw), &frontmatter); err != nil {
		return nil, fmt.Errorf("workflow_parse_error: %w", err)
	}
	if frontmatter == nil {
		frontmatter = map[string]any{}
	}
	encoded, err := json.Marshal(frontmatter)
	if err != nil {
		return nil, fmt.Errorf("workflow_parse_error: encode frontmatter: %w", err)
	}
	return encoded, nil
}
