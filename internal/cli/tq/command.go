package tq

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

const defaultAPIURL = "http://localhost:8080"

type config struct {
	apiURL string
	output string
}

type app struct {
	stdout io.Writer
	client *apiClient
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
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
	case "issue":
		return a.routeIssue(ctx, args[1:], cfg)
	case "comment":
		return a.routeComment(ctx, args[1:], cfg)
	default:
		return usageError("unknown resource %q", resource)
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
	case "get":
		return a.issueGet(ctx, args[1:], cfg)
	case "create":
		return a.issueCreate(ctx, args[1:], cfg)
	case "update":
		return a.issueUpdate(ctx, args[1:], cfg)
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
	if len(args) > 0 {
		return usageError("issue list does not accept positional arguments")
	}
	issues, err := a.client.listIssues(ctx)
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
	if err := fs.Parse(args); err != nil {
		return usageError(err.Error())
	}
	if fs.NArg() != 0 {
		return usageError("issue create does not accept positional arguments")
	}
	if *title == "" {
		return usageError("title is required")
	}

	input := entity.CreateIssueInput{
		Title:       *title,
		Description: *description,
		Status:      entity.Status(*status),
		Priority:    entity.Priority(*priority),
		Assignee:    *assignee,
	}
	issue, err := a.client.createIssue(ctx, input)
	if err != nil {
		return err
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
	if err := fs.Parse(args[1:]); err != nil {
		return usageError(err.Error())
	}
	if fs.NArg() != 0 {
		return usageError("issue update does not accept extra positional arguments")
	}

	patch := make(map[string]string)
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "title":
			patch["title"] = *title
		case "description":
			patch["description"] = *description
		case "status":
			patch["status"] = *status
		case "priority":
			patch["priority"] = *priority
		case "assignee":
			patch["assignee"] = *assignee
		}
	})
	if len(patch) == 0 {
		return usageError("at least one update field is required")
	}

	issue, err := a.client.updateIssue(ctx, id, patch)
	if err != nil {
		return err
	}
	return writeIssue(a.stdout, cfg.output, issue)
}

func parseCommon(args []string) (config, []string, error) {
	cfg := config{
		apiURL: strings.TrimSpace(os.Getenv("TQ_API_URL")),
		output: "text",
	}
	if cfg.apiURL == "" {
		cfg.apiURL = defaultAPIURL
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

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func defaultCommentAuthor() string {
	if value := strings.TrimSpace(os.Getenv("TQ_AUTHOR")); value != "" {
		return value
	}
	return strings.TrimSpace(os.Getenv("USER"))
}

func printRootHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq [--api-url URL] [--output text|json] <resource> <action> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Resources:")
	fmt.Fprintln(w, "  issue    create, get, list, and update issues")
	fmt.Fprintln(w, "  comment  add and list issue comments")
}

func printIssueHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq issue <action> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Actions:")
	fmt.Fprintln(w, "  create   Create an issue")
	fmt.Fprintln(w, "  get      Get an issue by ID")
	fmt.Fprintln(w, "  list     List issues")
	fmt.Fprintln(w, "  update   Update an issue")
}

func printCommentHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq comment <action> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Actions:")
	fmt.Fprintln(w, "  add      Add a comment to an issue")
	fmt.Fprintln(w, "  list     List comments for an issue")
}
