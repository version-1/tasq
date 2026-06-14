package workflow

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
	"sync"

	tqconfig "github.com/version-1/tasq/internal/config"
	"github.com/version-1/tasq/internal/issue/domain/entity"
	"gopkg.in/yaml.v3"
)

const workflowFileName = "WORKFLOW.md"

type ProjectWorkflowClient interface {
	Workflow(ctx context.Context, projectID int64) (entity.ProjectWorkflow, bool, error)
	UpsertWorkflow(ctx context.Context, projectID int64, input entity.UpsertProjectWorkflowInput) (entity.ProjectWorkflow, error)
}

type Resolver struct {
	client ProjectWorkflowClient

	mu            sync.Mutex
	projectCache  map[int64]cachedDefinition
	globalCache   cachedDefinition
	hasGlobalItem bool
}

type cachedDefinition struct {
	checksum   string
	definition Definition
}

func NewResolver(client ProjectWorkflowClient) *Resolver {
	return &Resolver{
		client:       client,
		projectCache: make(map[int64]cachedDefinition),
	}
}

func (r *Resolver) Resolve(ctx context.Context, project entity.Project) (Definition, error) {
	fileWorkflow, hasFile, err := readWorkflowFile(filepath.Join(project.Location, workflowFileName))
	if err != nil {
		return Definition{}, err
	}
	if hasFile && fileWorkflow.Checksum != project.WorkflowChecksum {
		updated, err := r.client.UpsertWorkflow(ctx, project.ID, entity.UpsertProjectWorkflowInput{
			ProjectID:       project.ID,
			FrontmatterJSON: fileWorkflow.FrontmatterJSON,
			Body:            fileWorkflow.Body,
			Checksum:        fileWorkflow.Checksum,
		})
		if err != nil {
			return Definition{}, fmt.Errorf("update project workflow %d: %w", project.ID, err)
		}
		definition, err := parseProjectWorkflowDefinition(updated, project.Location, "")
		if err != nil {
			return Definition{}, err
		}
		r.storeProject(project.ID, updated.Checksum, definition)
		return definition, nil
	}

	if project.WorkflowChecksum != "" {
		if definition, ok := r.cachedProject(project.ID, project.WorkflowChecksum); ok {
			return definition, nil
		}
		remoteWorkflow, found, err := r.client.Workflow(ctx, project.ID)
		if err != nil {
			return Definition{}, fmt.Errorf("fetch project workflow %d: %w", project.ID, err)
		}
		if found && remoteWorkflow.Checksum != "" {
			definition, err := parseProjectWorkflowDefinition(remoteWorkflow, project.Location, "")
			if err != nil {
				return Definition{}, err
			}
			r.storeProject(project.ID, remoteWorkflow.Checksum, definition)
			return definition, nil
		}
	}

	return r.globalDefinition()
}

func (r *Resolver) cachedProject(projectID int64, checksum string) (Definition, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cached, ok := r.projectCache[projectID]
	if !ok || cached.checksum != checksum {
		return Definition{}, false
	}
	return cached.definition, true
}

func (r *Resolver) storeProject(projectID int64, checksum string, definition Definition) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.projectCache[projectID] = cachedDefinition{checksum: checksum, definition: definition}
}

func (r *Resolver) globalDefinition() (Definition, error) {
	home, err := tqconfig.Home()
	if err != nil {
		return Definition{}, err
	}
	path := filepath.Join(home, workflowFileName)
	fileWorkflow, ok, err := readWorkflowFile(path)
	if err != nil {
		return Definition{}, err
	}
	if !ok {
		return Definition{}, fmt.Errorf("missing_global_workflow_file: %s", path)
	}
	r.mu.Lock()
	if r.hasGlobalItem && r.globalCache.checksum == fileWorkflow.Checksum {
		definition := r.globalCache.definition
		r.mu.Unlock()
		return definition, nil
	}
	r.mu.Unlock()

	definition, err := parseDefinition(fileWorkflow.Content, filepath.Dir(path), path)
	if err != nil {
		return Definition{}, err
	}
	r.mu.Lock()
	r.globalCache = cachedDefinition{checksum: fileWorkflow.Checksum, definition: definition}
	r.hasGlobalItem = true
	r.mu.Unlock()
	return definition, nil
}

func parseDefinition(content string, workflowDir string, path string) (Definition, error) {
	config, body, err := parse([]byte(content), workflowDir)
	if err != nil {
		return Definition{}, err
	}
	return Definition{
		Path:           path,
		Config:         config,
		PromptTemplate: strings.TrimSpace(body),
	}, nil
}

type workflowFile struct {
	Content         string
	FrontmatterJSON string
	Body            string
	Checksum        string
}

func parseProjectWorkflowDefinition(projectWorkflow entity.ProjectWorkflow, workflowDir string, path string) (Definition, error) {
	frontmatter, err := yaml.Marshal(projectWorkflow.Frontmatter)
	if err != nil {
		return Definition{}, fmt.Errorf("encode workflow frontmatter: %w", err)
	}
	content := "---\n" + string(frontmatter) + "---\n" + projectWorkflow.Body
	return parseDefinition(content, workflowDir, path)
}

func readWorkflowFile(path string) (workflowFile, bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return workflowFile{}, false, nil
		}
		return workflowFile{}, false, fmt.Errorf("read workflow file %s: %w", path, err)
	}
	frontmatterJSON, body, err := splitWorkflowContent(raw)
	if err != nil {
		return workflowFile{}, false, fmt.Errorf("parse workflow file %s: %w", path, err)
	}
	return workflowFile{
		Content:         string(raw),
		FrontmatterJSON: frontmatterJSON,
		Body:            body,
		Checksum:        checksum(raw),
	}, true, nil
}

func splitWorkflowContent(raw []byte) (string, string, error) {
	text := string(raw)
	if !strings.HasPrefix(text, "---\n") && strings.TrimSpace(text) != "---" {
		return "{}", text, nil
	}
	frontmatter, body, ok := splitFrontMatter(strings.TrimPrefix(text, "---\n"))
	if !ok {
		return "", "", errors.New("unterminated front matter")
	}
	var parsed map[string]any
	if err := yaml.Unmarshal([]byte(frontmatter), &parsed); err != nil {
		return "", "", err
	}
	if parsed == nil {
		parsed = map[string]any{}
	}
	encoded, err := json.Marshal(parsed)
	if err != nil {
		return "", "", err
	}
	return string(encoded), body, nil
}

func checksum(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
