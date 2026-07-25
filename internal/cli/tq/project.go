package tq

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

const workflowFileName = "WORKFLOW.md"

type projectAddResult struct {
	Project entity.Project `json:"project"`
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
	if len(args) > 0 && (args[0] == "help" || args[0] == "-help" || args[0] == "--help") {
		printProjectRemoveHelp(a.stdout)
		return nil
	}
	fs := newFlagSet("project remove")
	yes := fs.Bool("y", false, "remove without confirmation")
	if err := fs.Parse(args); err != nil {
		return usageError(err.Error())
	}
	if fs.NArg() != 1 {
		return usageError("usage: tq project remove [-y] <project-key>")
	}
	project, err := a.projectByKey(ctx, fs.Arg(0))
	if err != nil {
		return err
	}
	if !*yes {
		ok, err := a.confirmProjectRemoval(project)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(a.stdout, "Project removal cancelled")
			return nil
		}
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

func (a app) confirmProjectRemoval(project entity.Project) (bool, error) {
	fmt.Fprintf(a.stdout, "WARNING: This operation cannot be undone.\n")
	fmt.Fprintf(a.stdout, "Project to remove: %s (%s)\n", project.Key, project.Name)
	fmt.Fprintln(a.stdout, "This deletes the project and descendant data, including issues, comments, attachments, workflow overrides, and run data.")
	fmt.Fprintf(a.stdout, "Type the project key to confirm: ")
	line, err := bufio.NewReader(a.stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("read confirmation: %w", err)
	}
	return strings.TrimSpace(line) == project.Key, nil
}

func (a app) workflowRemove(ctx context.Context, args []string, cfg config) error {
	fs := newFlagSet("workflow remove")
	projectKey := fs.String("project", "", "project key")
	if err := fs.Parse(args); err != nil {
		return usageError(err.Error())
	}
	if fs.NArg() != 0 {
		return usageError("workflow remove does not accept positional arguments")
	}
	if *projectKey == "" {
		return usageError("project is required")
	}
	project, err := a.projectByKey(ctx, *projectKey)
	if err != nil {
		return err
	}
	if err := a.client.deleteProjectWorkflow(ctx, project.ID); err != nil {
		return err
	}
	if cfg.output == "json" {
		return writeJSON(a.stdout, map[string]any{"removed": true, "project": project})
	}
	fmt.Fprintf(a.stdout, "Removed workflow override for project %s\n", project.Key)
	return nil
}

func (a app) workflowAdd(ctx context.Context, args []string, cfg config) error {
	fs := newFlagSet("workflow add")
	projectKey := fs.String("project", "", "project key")
	filePath := fs.String("file", "", "workflow file path")
	body := fs.String("body", "", "workflow content")
	if err := fs.Parse(args); err != nil {
		return usageError(err.Error())
	}
	if fs.NArg() != 0 {
		return usageError("workflow add does not accept positional arguments")
	}
	if *projectKey == "" {
		return usageError("project is required")
	}
	if (*filePath == "" && *body == "") || (*filePath != "" && *body != "") {
		return usageError("exactly one of file or body is required")
	}

	content := *body
	if *filePath != "" {
		read, err := os.ReadFile(*filePath)
		if err != nil {
			return fmt.Errorf("read workflow file: %w", err)
		}
		content = string(read)
	}

	parsed, err := parseWorkflowAddContent(content)
	if err != nil {
		return err
	}
	project, err := a.projectByKey(ctx, *projectKey)
	if err != nil {
		return err
	}
	workflow, err := a.client.upsertProjectWorkflow(ctx, project.ID, upsertProjectWorkflowInput{
		Frontmatter: parsed.Frontmatter,
		Body:        parsed.Body,
		Checksum:    parsed.Checksum,
	})
	if err != nil {
		return err
	}
	return writeWorkflowAddResult(a.stdout, cfg.output, project, workflow)
}

func (a app) projectAdd(ctx context.Context, args []string, cfg config) error {
	fs := newFlagSet("project add")
	key := fs.String("key", "", "project key")
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

	local, err := prepareProjectFiles(projectRoot)
	if err != nil {
		local.rollback()
		return err
	}
	rollbackAPI = false
	return writeProjectAddResult(a.stdout, cfg.output, projectAddResult{Project: project})
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
	return nil
}

var invalidKeyChars = regexp.MustCompile(`[^a-z0-9]+`)
