package tq

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	tqconfig "github.com/version-1/tasq/internal/config"
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
		return writeCLIErrorForFormat(stderr, cfg.output, err.Error(), 2)
	}

	client, err := newAPIClient(cfg.apiURL)
	if err != nil {
		return writeCLIErrorForFormat(stderr, cfg.output, err.Error(), 2)
	}

	application := app{
		stdout: stdout,
		stdin:  stdin,
		client: client,
	}
	if err := application.route(ctx, remaining, cfg); err != nil {
		var statusErr apiStatusError
		if errors.As(err, &statusErr) {
			return statusErr.code
		}
		var ce cliError
		if errors.As(err, &ce) {
			return writeCLIErrorForFormat(stderr, cfg.output, ce.message, ce.code)
		}
		return writeCLIErrorForFormat(stderr, cfg.output, err.Error(), 1)
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
	case "config":
		return a.config(args[1:], cfg)
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
	case "api":
		if len(args) == 1 || args[1] == "help" || args[1] == "-help" || args[1] == "--help" {
			printAPIHelp(a.stdout)
			return nil
		}
		return a.api(ctx, args[1:])
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
