package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

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

type apiClient struct {
	baseURL    string
	httpClient *http.Client
}

type apiResponse[T any] struct {
	Data T `json:"data"`
}

type apiErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type cliError struct {
	message string
	code    int
}

func main() {
	code := run(context.Background(), os.Args[1:], os.Stdout, os.Stderr)
	os.Exit(code)
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
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

func newAPIClient(baseURL string) (*apiClient, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, errors.New("api-url is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("api-url is invalid: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("api-url must include scheme and host")
	}
	return &apiClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *apiClient) listIssues(ctx context.Context) ([]entity.Issue, error) {
	var issues []entity.Issue
	if err := c.do(ctx, http.MethodGet, "/api/v1/issues", nil, &issues); err != nil {
		return nil, err
	}
	return issues, nil
}

func (c *apiClient) getIssue(ctx context.Context, id int64) (entity.Issue, error) {
	var issue entity.Issue
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/issues/%d", id), nil, &issue); err != nil {
		return entity.Issue{}, err
	}
	return issue, nil
}

func (c *apiClient) createIssue(ctx context.Context, input entity.CreateIssueInput) (entity.Issue, error) {
	var issue entity.Issue
	if err := c.do(ctx, http.MethodPost, "/api/v1/issues", input, &issue); err != nil {
		return entity.Issue{}, err
	}
	return issue, nil
}

func (c *apiClient) updateIssue(ctx context.Context, id int64, patch map[string]string) (entity.Issue, error) {
	var issue entity.Issue
	if err := c.do(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/issues/%d", id), patch, &issue); err != nil {
		return entity.Issue{}, err
	}
	return issue, nil
}

func (c *apiClient) createComment(ctx context.Context, issueID int64, input entity.CreateCommentInput) (entity.Comment, error) {
	var comment entity.Comment
	if err := c.do(ctx, http.MethodPost, fmt.Sprintf("/api/v1/issues/%d/comments", issueID), input, &comment); err != nil {
		return entity.Comment{}, err
	}
	return comment, nil
}

func (c *apiClient) listComments(ctx context.Context, issueID int64) ([]entity.Comment, error) {
	var comments []entity.Comment
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/issues/%d/comments", issueID), nil, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

func (c *apiClient) do(ctx context.Context, method, path string, body any, output any) error {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return readAPIError(resp)
	}
	var payload apiResponse[json.RawMessage]
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if err := json.Unmarshal(payload.Data, output); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func readAPIError(resp *http.Response) error {
	var payload apiErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return cliError{message: fmt.Sprintf("%s returned %s", resp.Request.URL.Path, resp.Status), code: 1}
	}
	message := payload.Error.Message
	if message == "" {
		message = resp.Status
	}
	return cliError{message: message, code: 1}
}

func writeIssues(w io.Writer, format string, issues []entity.Issue) error {
	if format == "json" {
		return writeJSON(w, issues)
	}
	for _, issue := range issues {
		fmt.Fprintf(w, "#%d\t%s\t%s\n", issue.ID, issue.Status, issue.Title)
	}
	return nil
}

func writeIssue(w io.Writer, format string, issue entity.Issue) error {
	if format == "json" {
		return writeJSON(w, issue)
	}
	fmt.Fprintf(w, "ID: %d\n", issue.ID)
	fmt.Fprintf(w, "Title: %s\n", issue.Title)
	fmt.Fprintf(w, "Description: %s\n", issue.Description)
	fmt.Fprintf(w, "Status: %s\n", issue.Status)
	fmt.Fprintf(w, "Priority: %s\n", issue.Priority)
	fmt.Fprintf(w, "Assignee: %s\n", valueOrDefault(issue.Assignee, "unassigned"))
	fmt.Fprintf(w, "Created: %s\n", formatTime(issue.CreatedAt))
	fmt.Fprintf(w, "Updated: %s\n", formatTime(issue.UpdatedAt))
	return nil
}

func writeComments(w io.Writer, format string, comments []entity.Comment) error {
	if format == "json" {
		return writeJSON(w, comments)
	}
	for _, comment := range comments {
		fmt.Fprintf(w, "#%d\tissue:%d\t%s\t%s\t%s\n", comment.ID, comment.IssueID, comment.Type, comment.Author, comment.Body)
	}
	return nil
}

func writeComment(w io.Writer, format string, comment entity.Comment) error {
	if format == "json" {
		return writeJSON(w, comment)
	}
	fmt.Fprintf(w, "ID: %d\n", comment.ID)
	fmt.Fprintf(w, "Issue: %d\n", comment.IssueID)
	fmt.Fprintf(w, "Author: %s\n", comment.Author)
	fmt.Fprintf(w, "Type: %s\n", comment.Type)
	fmt.Fprintf(w, "Body: %s\n", comment.Body)
	fmt.Fprintf(w, "Created: %s\n", formatTime(comment.CreatedAt))
	return nil
}

func writeJSON(w io.Writer, value any) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func writeCLIError(w io.Writer, message string, code int) int {
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
	return code
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

func usageError(format string, args ...any) error {
	return cliError{message: fmt.Sprintf(format, args...), code: 2}
}

func (e cliError) Error() string {
	return e.message
}

func valueOrDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func defaultCommentAuthor() string {
	if value := strings.TrimSpace(os.Getenv("TQ_AUTHOR")); value != "" {
		return value
	}
	return strings.TrimSpace(os.Getenv("USER"))
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Local().Format(time.RFC3339)
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
