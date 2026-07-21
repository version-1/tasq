package tq

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	tqconfig "github.com/version-1/tasq/internal/config"
	"github.com/version-1/tasq/internal/issue/domain/entity"
)

var defaultAPIURL = "http://localhost:" + strconv.Itoa(tqconfig.DefaultIssueTrackerPort)

type config struct {
	apiURL string
	output string
}

type app struct {
	stdout io.Writer
	stdin  io.Reader
	client *apiClient
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	return run(ctx, args, os.Stdin, stdout, stderr)
}

func run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	cfg, remaining, err := parseCommon(args)
	if err != nil {
		return writeCLIError(stderr, err.Error(), 2)
	}

	client, err := newAPIClient(cfg.apiURL)
	if err != nil {
		return writeCLIError(stderr, err.Error(), 2)
	}

	application := app{
		stdout: stdout,
		stdin:  stdin,
		client: client,
	}
	if err := application.route(ctx, remaining, cfg); err != nil {
		var ce cliError
		if errors.As(err, &ce) {
			return writeCLIError(stderr, ce.message, ce.code)
		}
		return writeCLIError(stderr, err.Error(), 1)
	}
	return 0
}

func (a app) route(ctx context.Context, args []string, cfg config) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-help" || args[0] == "--help" {
		printRootHelp(a.stdout)
		return nil
	}
	resource := args[0]
	switch resource {
	case "version":
		printVersion(a.stdout)
		return nil
	case "update":
		return a.update(ctx, args[1:], cfg)
	case "issue":
		return a.routeIssue(ctx, args[1:], cfg)
	case "comment":
		return a.routeComment(ctx, args[1:], cfg)
	case "project":
		return a.routeProject(ctx, args[1:], cfg)
	case "workflow":
		return a.routeWorkflow(ctx, args[1:], cfg)
	case "migrate":
		return a.routeMigrate(ctx, args[1:], cfg)
	case "web":
		return a.web(args[1:])
	case "service":
		return a.routeService(ctx, args[1:], cfg)
	case "logs":
		return a.routeLogs(ctx, args[1:], cfg)
	default:
		return usageError("unknown resource %q", resource)
	}
}

func (a app) web(args []string) error {
	if len(args) != 0 {
		return usageError("usage: tq web")
	}
	state, err := tqconfig.ReadState()
	if err != nil {
		return err
	}
	if state.Web == nil || !processAlive(state.Web.PID) {
		return errors.New("web UI is not running; run `tq service start` first")
	}
	webURL, ok, err := tqconfig.WebURLFromState()
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("web UI is not running; run `tq service start` first")
	}
	if err := openBrowser(webURL); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	fmt.Fprintf(a.stdout, "Opening %s\n", webURL)
	return nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default:
		return fmt.Errorf("unsupported OS %q", runtime.GOOS)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

func (a app) routeProject(ctx context.Context, args []string, cfg config) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-help" || args[0] == "--help" {
		printProjectHelp(a.stdout)
		return nil
	}
	action := args[0]
	switch action {
	case "add":
		return a.projectAdd(ctx, args[1:], cfg)
	case "remove":
		return a.projectRemove(ctx, args[1:], cfg)
	case "check":
		return a.projectCheck(ctx, args[1:], cfg)
	case "list":
		return a.projectList(ctx, args[1:], cfg)
	default:
		return usageError("unknown project action %q", action)
	}
}

func (a app) routeWorkflow(ctx context.Context, args []string, cfg config) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-help" || args[0] == "--help" {
		printWorkflowHelp(a.stdout)
		return nil
	}
	action := args[0]
	switch action {
	case "add":
		return a.workflowAdd(ctx, args[1:], cfg)
	case "remove":
		return a.workflowRemove(ctx, args[1:], cfg)
	case "show":
		return a.workflowShow(ctx, args[1:], cfg)
	default:
		return usageError("unknown workflow action %q", action)
	}
}

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

func (a app) routeComment(ctx context.Context, args []string, cfg config) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-help" || args[0] == "--help" {
		printCommentHelp(a.stdout)
		return nil
	}
	action := args[0]
	switch action {
	case "add":
		return a.commentAdd(ctx, args[1:], cfg)
	case "list":
		return a.commentList(ctx, args[1:], cfg)
	default:
		return usageError("unknown comment action %q", action)
	}
}

func (a app) commentAdd(ctx context.Context, args []string, cfg config) error {
	if len(args) == 0 {
		return usageError("usage: tq comment add <issue-id> [flags]")
	}
	issueID, err := parseID(args[0])
	if err != nil {
		return err
	}
	fs := newFlagSet("comment add")
	author := fs.String("author", defaultCommentAuthor(), "comment author")
	commentType := fs.String("type", string(entity.CommentGeneral), "comment type")
	body := fs.String("body", "", "comment body")
	attach := fs.String("attach", "", "image attachment path")
	if err := fs.Parse(args[1:]); err != nil {
		return usageError(err.Error())
	}
	if fs.NArg() != 0 {
		return usageError("comment add does not accept extra positional arguments")
	}
	if *author == "" {
		return usageError("author is required")
	}
	if *body == "" {
		return usageError("body is required")
	}
	input := entity.CreateCommentInput{
		Author: *author,
		Type:   entity.CommentType(*commentType),
		Body:   *body,
	}
	comment, err := a.client.createComment(ctx, issueID, input)
	if err != nil {
		return err
	}
	if *attach != "" {
		attachment, err := a.client.uploadAttachment(ctx, attachmentUploadInput{
			EntityType: entity.AttachmentEntityComment,
			EntityID:   strconv.FormatInt(comment.ID, 10),
			Path:       *attach,
		})
		if err != nil {
			return err
		}
		body := appendAttachmentMarkdown(comment.Body, attachment)
		comment, err = a.client.updateComment(ctx, comment.ID, entity.UpdateCommentInput{Body: &body})
		if err != nil {
			_ = a.client.deleteAttachment(ctx, attachment.ID)
			return err
		}
	}
	return writeComment(a.stdout, cfg.output, comment)
}

func (a app) commentList(ctx context.Context, args []string, cfg config) error {
	if len(args) != 1 {
		return usageError("usage: tq comment list <issue-id>")
	}
	issueID, err := parseID(args[0])
	if err != nil {
		return err
	}
	comments, err := a.client.listComments(ctx, issueID)
	if err != nil {
		return err
	}
	return writeComments(a.stdout, cfg.output, comments)
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
	changedReason := fs.String("changed-reason", "", "status change reason")
	changedBy := fs.String("changed-by", defaultCommentAuthor(), "status change actor")
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
	changedBySet := false
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
		case "changed-reason":
			patch.ChangedReason = changedReason
			updated = true
		case "changed-by":
			patch.ChangedBy = changedBy
			changedBySet = true
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
	if changedBySet && patch.Status == nil {
		return usageError("changed-by requires status")
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
			// The issue update is the step that links the uploaded file from Markdown.
			// Delete the uploaded file on failure so the API does not keep an orphan.
			_ = a.client.deleteAttachment(ctx, uploadedAttachmentID)
		}
		return err
	}
	return writeIssue(a.stdout, cfg.output, issue)
}

func (a app) issueSetStatus(ctx context.Context, args []string, cfg config, status entity.Status, action string, usage string) error {
	if len(args) == 0 {
		return usageError(usage)
	}
	id, err := parseID(args[0])
	if err != nil {
		return err
	}
	fs := newFlagSet("issue status")
	changedReason := fs.String("changed-reason", "", "status change reason")
	changedBy := fs.String("changed-by", defaultCommentAuthor(), "status change actor")
	if err := fs.Parse(args[1:]); err != nil {
		return usageError(err.Error())
	}
	if fs.NArg() != 0 {
		return usageError("issue status action does not accept extra positional arguments")
	}
	statusValue := status
	input := updateIssueInput{Status: &statusValue}
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "changed-reason":
			input.ChangedReason = changedReason
		case "changed-by":
			input.ChangedBy = changedBy
		}
	})
	issue, err := a.client.updateIssue(ctx, id, input)
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

func parseCommon(args []string) (config, []string, error) {
	cfg := config{
		output: "text",
	}

	var remaining []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--api-url" || arg == "-api-url":
			if i+1 >= len(args) {
				return cfg, nil, errors.New("api-url requires a value")
			}
			i++
			cfg.apiURL = args[i]
		case strings.HasPrefix(arg, "--api-url="):
			cfg.apiURL = strings.TrimPrefix(arg, "--api-url=")
		case strings.HasPrefix(arg, "-api-url="):
			cfg.apiURL = strings.TrimPrefix(arg, "-api-url=")
		case arg == "--output" || arg == "-output":
			if i+1 >= len(args) {
				return cfg, nil, errors.New("output requires a value")
			}
			i++
			cfg.output = args[i]
		case strings.HasPrefix(arg, "--output="):
			cfg.output = strings.TrimPrefix(arg, "--output=")
		case strings.HasPrefix(arg, "-output="):
			cfg.output = strings.TrimPrefix(arg, "-output=")
		default:
			remaining = append(remaining, arg)
		}
	}
	if cfg.apiURL == "" {
		cfg.apiURL = strings.TrimSpace(os.Getenv("TQ_API_URL"))
	}
	if cfg.apiURL == "" {
		apiURL, ok, err := tqconfig.IssueTrackerURLFromState()
		if err != nil {
			return cfg, nil, err
		}
		if ok {
			cfg.apiURL = apiURL
		}
	}
	if cfg.apiURL == "" {
		cfg.apiURL = defaultAPIURL
	}
	if cfg.output == "text" {
		return cfg, remaining, nil
	}
	if cfg.output == "json" {
		return cfg, remaining, nil
	}
	return cfg, nil, fmt.Errorf("unsupported output %q", cfg.output)
}

func parseID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, usageError("id must be a positive integer")
	}
	return id, nil
}

func parseDependencyIDs(value string) ([]int64, error) {
	if value == "" {
		return nil, usageError("dependency must not be empty")
	}
	parts := strings.Split(value, ",")
	ids := make([]int64, 0, len(parts))
	seen := map[int64]struct{}{}
	for _, part := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
		if err != nil || id <= 0 {
			return nil, usageError("dependency must be a comma-separated list of positive integers")
		}
		if _, ok := seen[id]; ok {
			return nil, usageError("dependency contains duplicate issue id")
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids, nil
}

func updateIssueTitleInput(title string) updateIssueInput {
	return updateIssueInput{Title: &title}
}

func updateIssueDescriptionInput(description string) updateIssueInput {
	return updateIssueInput{Description: &description}
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func defaultCommentAuthor() string {
	if value := strings.TrimSpace(os.Getenv("TQ_AUTHOR")); value != "" {
		return value
	}
	if config, err := tqconfig.Load(); err == nil && strings.TrimSpace(config.Author) != "" {
		return strings.TrimSpace(config.Author)
	}
	return strings.TrimSpace(os.Getenv("USER"))
}

func appendAttachmentMarkdown(content string, attachment entity.Attachment) string {
	markdown := fmt.Sprintf("![%s](attachment://%s)", markdownAltText(attachment.Filename), attachment.ID)
	content = strings.TrimRight(content, "\n")
	if content == "" {
		return markdown
	}
	return content + "\n\n" + markdown
}

func markdownAltText(value string) string {
	value = strings.ReplaceAll(value, "[", "")
	value = strings.ReplaceAll(value, "]", "")
	value = strings.ReplaceAll(value, "\n", " ")
	return strings.TrimSpace(value)
}

func printRootHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq [--api-url URL] [--output text|json] <resource> <action> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Resources:")
	fmt.Fprintln(w, "  issue    create, get, list, update, and shortcut issue actions")
	fmt.Fprintln(w, "  comment  add and list issue comments")
	fmt.Fprintln(w, "  project  add, remove, check, and list projects")
	fmt.Fprintln(w, "  workflow add, remove, and show project workflow resolution")
	fmt.Fprintln(w, "  migrate  apply, roll back, and inspect local database migrations")
	fmt.Fprintln(w, "  web      open the running Web UI in the default browser")
	fmt.Fprintln(w, "  service  start, stop, and inspect local services")
	fmt.Fprintln(w, "  logs     show and follow service logs")
	fmt.Fprintln(w, "  version  show version information")
	fmt.Fprintln(w, "  update   update tq from a GitHub Release and restart services")
}

func printIssueHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq issue <action> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Actions:")
	fmt.Fprintln(w, "  create   Create an issue (--project KEY required)")
	fmt.Fprintln(w, "  get      Get an issue by ID")
	fmt.Fprintln(w, "  list     List issues (--project KEY optional)")
	fmt.Fprintln(w, "  watch    Poll ready issues and emit JSON events (--interval, --seen-ttl, --verbose)")
	fmt.Fprintln(w, "  update   Update an issue")
	fmt.Fprintln(w, "  close    Mark an issue as done")
	fmt.Fprintln(w, "  cancel   Mark an issue as cancelled")
	fmt.Fprintln(w, "  ready    Mark an issue as ready")
	fmt.Fprintln(w, "  draft    Move an issue to backlog")
	fmt.Fprintln(w, "  rename   Rename an issue")
	fmt.Fprintln(w, "  edit     Update an issue description")
}

func printCommentHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq comment <action> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Actions:")
	fmt.Fprintln(w, "  add      Add a comment to an issue")
	fmt.Fprintln(w, "  list     List comments for an issue")
}

func printProjectHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq project <action> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Actions:")
	fmt.Fprintln(w, "  add      Add a project for a repository")
	fmt.Fprintln(w, "  remove   Remove a project by key (-y skips confirmation)")
	fmt.Fprintln(w, "  check    Check local project workflow files")
	fmt.Fprintln(w, "  list     List projects")
}

func printProjectRemoveHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq project remove [-y] <project-key>")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  -y  Remove without confirmation")
}

func printWorkflowHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq workflow <action> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Actions:")
	fmt.Fprintln(w, "  add      Add or replace a project workflow override (--project KEY and --file PATH or --body TEXT required)")
	fmt.Fprintln(w, "  remove   Remove a project workflow override (--project KEY required)")
	fmt.Fprintln(w, "  show     Show the resolved project workflow")
}
