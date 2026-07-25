package tq

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func (a app) routeIssue(ctx context.Context, args []string, cfg config) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-help" || args[0] == "--help" {
		printIssueHelp(a.stdout)
		return nil
	}
	action := args[0]
	switch action {
	case "list":
		return a.issueList(ctx, args[1:], cfg)
	case "watch":
		return a.issueWatch(ctx, args[1:])
	case "get":
		return a.issueGet(ctx, args[1:], cfg)
	case "create":
		return a.issueCreate(ctx, args[1:], cfg)
	case "update":
		return a.issueUpdate(ctx, args[1:], cfg)
	case "close":
		return a.issueSetStatus(ctx, args[1:], cfg, entity.StatusDone, "closed", "usage: tq issue close <id>")
	case "cancel":
		return a.issueSetStatus(ctx, args[1:], cfg, entity.StatusCancelled, "cancelled", "usage: tq issue cancel <id>")
	case "ready":
		return a.issueSetStatus(ctx, args[1:], cfg, entity.StatusReady, "marked as ready", "usage: tq issue ready <id>")
	case "draft":
		return a.issueSetStatus(ctx, args[1:], cfg, entity.StatusBacklog, "moved to backlog", "usage: tq issue draft <id>")
	case "rename":
		return a.issueRename(ctx, args[1:], cfg)
	case "edit":
		return a.issueEdit(ctx, args[1:], cfg)
	default:
		return usageError("unknown issue action %q", action)
	}
}

func (a app) issueList(ctx context.Context, args []string, cfg config) error {
	fs := newFlagSet("issue list")
	projectKey := fs.String("project", "", "project key")
	if err := fs.Parse(args); err != nil {
		return usageError(err.Error())
	}
	if fs.NArg() > 0 {
		return usageError("issue list does not accept positional arguments")
	}
	var projectID *int64
	if *projectKey != "" {
		project, err := a.client.projectByKey(ctx, *projectKey)
		if err != nil {
			return err
		}
		projectID = &project.ID
	}
	issues, err := a.client.listIssues(ctx, projectID)
	if err != nil {
		return err
	}
	return writeIssues(a.stdout, cfg.output, issues)
}

func (a app) issueGet(ctx context.Context, args []string, cfg config) error {
	if len(args) != 1 {
		return usageError("usage: tq issue get <id>")
	}
	id, err := parseID(args[0])
	if err != nil {
		return err
	}
	issue, err := a.client.getIssue(ctx, id)
	if err != nil {
		return err
	}
	return writeIssue(a.stdout, cfg.output, issue)
}

func (a app) issueCreate(ctx context.Context, args []string, cfg config) error {
	fs := newFlagSet("issue create")
	title := fs.String("title", "", "issue title")
	description := fs.String("description", "", "issue description")
	status := fs.String("status", "", "issue status")
	priority := fs.String("priority", "", "issue priority")
	assignee := fs.String("assignee", "", "issue assignee")
	attach := fs.String("attach", "", "image attachment path")
	dependency := fs.String("dependency", "", "comma-separated dependency issue IDs")
	projectKey := fs.String("project", "", "project key")
	if err := fs.Parse(args); err != nil {
		return usageError(err.Error())
	}
	if fs.NArg() != 0 {
		return usageError("issue create does not accept positional arguments")
	}
	if *title == "" {
		return usageError("title is required")
	}
	if *projectKey == "" {
		return usageError("project is required")
	}
	dependencySet := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "dependency" {
			dependencySet = true
		}
	})
	var dependencyIDs []int64
	if dependencySet {
		var err error
		dependencyIDs, err = parseDependencyIDs(*dependency)
		if err != nil {
			return err
		}
	}
	project, err := a.client.projectByKey(ctx, *projectKey)
	if err != nil {
		return err
	}

	input := entity.CreateIssueInput{
		ProjectID:   project.ID,
		Title:       *title,
		Description: *description,
		Status:      entity.Status(*status),
		Priority:    entity.Priority(*priority),
		Assignee:    *assignee,
	}
	if dependencySet {
		input.DependencyIDs = dependencyIDs
	}
	issue, err := a.client.createIssue(ctx, input)
	if err != nil {
		return err
	}
	if *attach != "" {
		attachment, err := a.client.uploadAttachment(ctx, attachmentUploadInput{
			EntityType: entity.AttachmentEntityIssue,
			EntityID:   strconv.FormatInt(issue.ID, 10),
			Path:       *attach,
		})
		if err != nil {
			return err
		}
		description := appendAttachmentMarkdown(issue.Description, attachment)
		issue, err = a.client.updateIssue(ctx, issue.ID, updateIssueDescriptionInput(description))
		if err != nil {
			_ = a.client.deleteAttachment(ctx, attachment.ID)
			return err
		}
	}
	return writeIssue(a.stdout, cfg.output, issue)
}

func (a app) issueUpdate(ctx context.Context, args []string, cfg config) error {
	if len(args) == 0 {
		return usageError("usage: tq issue update <id> [flags]")
	}
	id, err := parseID(args[0])
	if err != nil {
		return err
	}

	fs := newFlagSet("issue update")
	title := fs.String("title", "", "issue title")
	description := fs.String("description", "", "issue description")
	status := fs.String("status", "", "issue status")
	priority := fs.String("priority", "", "issue priority")
	assignee := fs.String("assignee", "", "issue assignee")
	dependency := fs.String("dependency", "", "comma-separated dependency issue IDs; fully replaces dependencies")
	clearDependencies := fs.Bool("clear-dependencies", false, "clear all dependency issue IDs")
	attach := fs.String("attach", "", "image attachment path")
	if err := fs.Parse(args[1:]); err != nil {
		return usageError(err.Error())
	}
	if fs.NArg() != 0 {
		return usageError("issue update does not accept extra positional arguments")
	}

	var patch updateIssueInput
	updated := false
	dependencySet := false
	clearDependenciesSet := false
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "title":
			patch.Title = title
			updated = true
		case "description":
			patch.Description = description
			updated = true
		case "status":
			statusValue := entity.Status(*status)
			patch.Status = &statusValue
			updated = true
		case "priority":
			priorityValue := entity.Priority(*priority)
			patch.Priority = &priorityValue
			updated = true
		case "assignee":
			patch.Assignee = assignee
			updated = true
		case "dependency":
			dependencySet = true
		case "clear-dependencies":
			clearDependenciesSet = *clearDependencies
		case "attach":
		}
	})
	if dependencySet && clearDependenciesSet {
		return usageError("dependency and clear-dependencies cannot be used together")
	}
	if dependencySet {
		if *dependency == "" {
			return usageError("dependency must not be empty; use --clear-dependencies to clear dependencies")
		}
		dependencyIDs, err := parseDependencyIDs(*dependency)
		if err != nil {
			return err
		}
		patch.DependencyIDs = &dependencyIDs
		updated = true
	}
	if clearDependenciesSet {
		dependencyIDs := []int64{}
		patch.DependencyIDs = &dependencyIDs
		updated = true
	}
	if !updated && *attach == "" {
		return usageError("at least one update field is required")
	}
	var uploadedAttachmentID string
	if *attach != "" {
		attachment, err := a.client.uploadAttachment(ctx, attachmentUploadInput{
			EntityType: entity.AttachmentEntityIssue,
			EntityID:   strconv.FormatInt(id, 10),
			Path:       *attach,
		})
		if err != nil {
			return err
		}
		uploadedAttachmentID = attachment.ID
		var descriptionText string
		if patch.Description != nil {
			descriptionText = *patch.Description
		} else {
			current, err := a.client.getIssue(ctx, id)
			if err != nil {
				_ = a.client.deleteAttachment(ctx, uploadedAttachmentID)
				return err
			}
			descriptionText = current.Description
		}
		updatedDescription := appendAttachmentMarkdown(descriptionText, attachment)
		patch.Description = &updatedDescription
	}

	issue, err := a.client.updateIssue(ctx, id, patch)
	if err != nil {
		if uploadedAttachmentID != "" {

			_ = a.client.deleteAttachment(ctx, uploadedAttachmentID)
		}
		return err
	}
	return writeIssue(a.stdout, cfg.output, issue)
}

func (a app) issueSetStatus(ctx context.Context, args []string, cfg config, status entity.Status, action string, usage string) error {
	if len(args) != 1 {
		return usageError(usage)
	}
	id, err := parseID(args[0])
	if err != nil {
		return err
	}
	statusValue := status
	issue, err := a.client.updateIssue(ctx, id, updateIssueInput{Status: &statusValue})
	if err != nil {
		return err
	}
	return writeIssueAction(a.stdout, cfg.output, issue, fmt.Sprintf("Issue #%d %s", id, action))
}

func (a app) issueRename(ctx context.Context, args []string, cfg config) error {
	if len(args) != 2 {
		return usageError("usage: tq issue rename <id> <title>")
	}
	id, err := parseID(args[0])
	if err != nil {
		return err
	}
	issue, err := a.client.updateIssue(ctx, id, updateIssueTitleInput(args[1]))
	if err != nil {
		return err
	}
	return writeIssueAction(a.stdout, cfg.output, issue, fmt.Sprintf("Issue #%d renamed", id))
}

func (a app) issueEdit(ctx context.Context, args []string, cfg config) error {
	if len(args) != 2 {
		return usageError("usage: tq issue edit <id> <description>")
	}
	id, err := parseID(args[0])
	if err != nil {
		return err
	}
	issue, err := a.client.updateIssue(ctx, id, updateIssueDescriptionInput(args[1]))
	if err != nil {
		return err
	}
	return writeIssueAction(a.stdout, cfg.output, issue, fmt.Sprintf("Issue #%d description updated", id))
}

func updateIssueTitleInput(title string) updateIssueInput {
	return updateIssueInput{Title: &title}
}

func updateIssueDescriptionInput(description string) updateIssueInput {
	return updateIssueInput{Description: &description}
}
