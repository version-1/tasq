package tq

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	tqconfig "github.com/version-1/tasq/internal/config"
)

const (
	serviceHealthTimeout = 10 * time.Second
	serviceStopGrace     = 5 * time.Second
)

type serviceName string

const (
	serviceIssueTracker serviceName = "issue-tracker"
	serviceOrchestrator serviceName = "orchestrator"
	serviceWeb          serviceName = "web"
)

type serviceStatus struct {
	Name      string `json:"name"`
	State     string `json:"state"`
	PID       int    `json:"pid,omitempty"`
	Addr      string `json:"addr,omitempty"`
	Port      int    `json:"port,omitempty"`
	Uptime    string `json:"uptime,omitempty"`
	StartedAt string `json:"started_at,omitempty"`
}

type managedService struct {
	name    serviceName
	logName string
	args    []string
}

func (a app) routeService(ctx context.Context, args []string, cfg config) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-help" || args[0] == "--help" {
		printServiceHelp(a.stdout)
		return nil
	}
	action := args[0]
	switch action {
	case "start":
		return a.serviceStart(ctx, args[1:], cfg)
	case "stop":
		return a.serviceStop(ctx, args[1:], cfg)
	case "status":
		return a.serviceStatus(ctx, args[1:], cfg)
	default:
		return usageError("unknown service action %q", action)
	}
}

func (a app) serviceStart(ctx context.Context, args []string, cfg config) error {
	if len(args) != 0 {
		return usageError("service start does not accept positional arguments")
	}
	home, err := tqconfig.EnsureHome()
	if err != nil {
		return err
	}
	if err := cleanupStaleServices(); err != nil {
		return err
	}
	state, err := tqconfig.ReadState()
	if err != nil {
		return err
	}
	if state.IssueTracker != nil && processAlive(state.IssueTracker.PID) {
		return usageError("issue-tracker is already running")
	}
	if state.Orchestrator != nil && processAlive(state.Orchestrator.PID) {
		return usageError("orchestrator is already running")
	}
	if state.Web != nil && processAlive(state.Web.PID) {
		return usageError("web is already running")
	}

	issueAddr := "127.0.0.1:" + strconv.Itoa(tqconfig.DefaultIssueTrackerPort)
	orchestratorAddr := "127.0.0.1:" + strconv.Itoa(tqconfig.DefaultOrchestratorPort)
	webAddr := "127.0.0.1:" + strconv.Itoa(tqconfig.DefaultWebPort)
	issueDB := tqconfig.IssueTrackerDBPath(home)
	orchestratorDB := tqconfig.OrchestratorDBPath(home)

	issueService := managedService{
		name:    serviceIssueTracker,
		logName: "issue-tracker.log",
		args: []string{
			"-addr", issueAddr,
			"-db", issueDB,
		},
	}
	if _, err := startManagedService(ctx, home, issueService); err != nil {
		return err
	}
	if err := waitIssueTrackerHealthy(ctx, "http://"+issueAddr); err != nil {
		_ = stopServiceByName(context.Background(), serviceIssueTracker)
		return fmt.Errorf("issue-tracker health check failed: %w", err)
	}

	orchestratorService := managedService{
		name:    serviceOrchestrator,
		logName: "orchestrator.log",
		args: []string{
			"-db", orchestratorDB,
			"-issue-tracker", "http://" + issueAddr,
			"-port", strconv.Itoa(tqconfig.DefaultOrchestratorPort),
		},
	}
	if _, err := startManagedService(ctx, home, orchestratorService); err != nil {
		_ = stopServiceByName(context.Background(), serviceIssueTracker)
		return err
	}
	if err := waitServiceRunning(ctx, serviceOrchestrator); err != nil {
		_ = stopServiceByName(context.Background(), serviceOrchestrator)
		_ = stopServiceByName(context.Background(), serviceIssueTracker)
		return fmt.Errorf("orchestrator startup failed: %w", err)
	}

	webService := managedService{
		name:    serviceWeb,
		logName: "web.log",
		args: []string{
			"-addr", webAddr,
			"-tracker-url", "http://" + issueAddr,
			"-orchestrator-url", "http://" + orchestratorAddr,
		},
	}
	if _, err := startManagedService(ctx, home, webService); err != nil {
		_ = stopServiceByName(context.Background(), serviceOrchestrator)
		_ = stopServiceByName(context.Background(), serviceIssueTracker)
		return err
	}
	if err := waitWebHealthy(ctx, "http://"+webAddr); err != nil {
		_ = stopServiceByName(context.Background(), serviceWeb)
		_ = stopServiceByName(context.Background(), serviceOrchestrator)
		_ = stopServiceByName(context.Background(), serviceIssueTracker)
		return fmt.Errorf("web health check failed: %w", err)
	}

	if cfg.output == "json" {
		return a.serviceStatus(ctx, nil, cfg)
	}
	_, err = fmt.Fprintf(a.stdout, "%s✓%s Services started\n", ansiGreen, ansiReset)
	return err
}

func (a app) serviceStop(ctx context.Context, args []string, cfg config) error {
	if len(args) != 0 {
		return usageError("service stop does not accept positional arguments")
	}
	if err := stopServiceByName(ctx, serviceWeb); err != nil {
		return err
	}
	if err := stopServiceByName(ctx, serviceOrchestrator); err != nil {
		return err
	}
	if err := stopServiceByName(ctx, serviceIssueTracker); err != nil {
		return err
	}
	if err := cleanupServiceState(serviceWeb, true); err != nil {
		return err
	}
	if err := cleanupServiceState(serviceOrchestrator, true); err != nil {
		return err
	}
	if err := cleanupServiceState(serviceIssueTracker, true); err != nil {
		return err
	}
	if cfg.output == "json" {
		return a.serviceStatus(ctx, nil, cfg)
	}
	_, err := fmt.Fprintf(a.stdout, "%s✓%s Services stopped\n", ansiGreen, ansiReset)
	return err
}

func (a app) serviceStatus(ctx context.Context, args []string, cfg config) error {
	if len(args) != 0 {
		return usageError("service status does not accept positional arguments")
	}
	if err := cleanupStaleServices(); err != nil {
		return err
	}
	state, err := tqconfig.ReadState()
	if err != nil {
		return err
	}
	statuses := []serviceStatus{
		statusForService(serviceIssueTracker, state.IssueTracker),
		statusForService(serviceOrchestrator, state.Orchestrator),
		statusForService(serviceWeb, state.Web),
	}
	if cfg.output == "json" {
		return writeJSON(a.stdout, statuses)
	}
	for _, status := range statuses {
		if status.State == "running" {
			fmt.Fprintf(a.stdout, "%s\trunning\tpid=%d\tport=%d\tuptime=%s\n", status.Name, status.PID, status.Port, status.Uptime)
			continue
		}
		fmt.Fprintf(a.stdout, "%s\tstopped\n", status.Name)
	}
	return nil
}

func startManagedService(ctx context.Context, home string, service managedService) (int, error) {
	cmd, err := commandForService(ctx, service)
	if err != nil {
		return 0, err
	}
	logFile, err := openServiceLog(home, service.logName)
	if err != nil {
		return 0, err
	}
	defer logFile.Close()
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Dir = serviceWorkingDir()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("start %s: %w", service.name, err)
	}
	return cmd.Process.Pid, nil
}

func commandForService(ctx context.Context, service managedService) (*exec.Cmd, error) {
	if executable, ok := siblingServiceExecutable(service.name); ok {
		return exec.CommandContext(ctx, executable, service.args...), nil
	}
	if _, err := os.Stat(filepath.Join("cmd", string(service.name), "main.go")); err != nil {
		return nil, fmt.Errorf("%s executable not found next to tq and ./cmd/%s is unavailable", service.name, service.name)
	}
	args := append([]string{"run", "./cmd/" + string(service.name)}, service.args...)
	return exec.CommandContext(ctx, "go", args...), nil
}

func siblingServiceExecutable(name serviceName) (string, bool) {
	executable, err := os.Executable()
	if err != nil {
		return "", false
	}
	path := filepath.Join(filepath.Dir(executable), string(name))
	if runtime.GOOS == "windows" {
		path += ".exe"
	}
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return path, true
	}
	return "", false
}

func serviceWorkingDir() string {
	if cwd, err := os.Getwd(); err == nil {
		return cwd
	}
	return "."
}

func openServiceLog(home string, name string) (*os.File, error) {
	dir := tqconfig.LogDir(home)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create log dir: %w", err)
	}
	path := filepath.Join(dir, name)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	return file, nil
}

func waitIssueTrackerHealthy(ctx context.Context, baseURL string) error {
	return waitHTTPHealthy(ctx, strings.TrimRight(baseURL, "/")+"/api/v1/health")
}

func waitWebHealthy(ctx context.Context, baseURL string) error {
	return waitHTTPHealthy(ctx, strings.TrimRight(baseURL, "/")+"/health")
}

func waitHTTPHealthy(ctx context.Context, url string) error {
	deadline := time.Now().Add(serviceHealthTimeout)
	client := http.Client{Timeout: time.Second}
	var lastErr error
	for time.Now().Before(deadline) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		response, err := client.Do(req)
		if err == nil {
			_ = response.Body.Close()
			if response.StatusCode >= 200 && response.StatusCode < 300 {
				return nil
			}
			lastErr = fmt.Errorf("unexpected status %d", response.StatusCode)
		} else {
			lastErr = err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
	if lastErr == nil {
		lastErr = errors.New("timeout")
	}
	return lastErr
}

func waitServiceRunning(ctx context.Context, name serviceName) error {
	deadline := time.Now().Add(serviceHealthTimeout)
	for time.Now().Before(deadline) {
		state, err := tqconfig.ReadState()
		if err != nil {
			return err
		}
		if service := serviceStateByName(state, name); service != nil && processAlive(service.PID) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
	return errors.New("timeout waiting for service state")
}

func stopServiceByName(ctx context.Context, name serviceName) error {
	state, err := tqconfig.ReadState()
	if err != nil {
		return err
	}
	service := serviceStateByName(state, name)
	if service == nil || service.PID <= 0 {
		return nil
	}
	return terminatePID(ctx, service.PID)
}

func terminatePID(ctx context.Context, pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil
	}
	if !processAlive(pid) {
		return nil
	}
	if err := process.Signal(syscall.SIGTERM); err != nil && !isProcessDone(err) {
		return err
	}
	deadline := time.Now().Add(serviceStopGrace)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	if err := process.Kill(); err != nil && !isProcessDone(err) {
		return err
	}
	return nil
}

func cleanupStaleServices() error {
	return tqconfig.UpdateState(func(state *tqconfig.State) error {
		if state.IssueTracker != nil && !processAlive(state.IssueTracker.PID) {
			state.IssueTracker = nil
		}
		if state.Orchestrator != nil && !processAlive(state.Orchestrator.PID) {
			state.Orchestrator = nil
		}
		if state.Web != nil && !processAlive(state.Web.PID) {
			state.Web = nil
		}
		return nil
	})
}

func cleanupServiceState(name serviceName, force bool) error {
	return tqconfig.UpdateState(func(state *tqconfig.State) error {
		switch name {
		case serviceIssueTracker:
			if force || state.IssueTracker == nil || !processAlive(state.IssueTracker.PID) {
				state.IssueTracker = nil
			}
		case serviceOrchestrator:
			if force || state.Orchestrator == nil || !processAlive(state.Orchestrator.PID) {
				state.Orchestrator = nil
			}
		case serviceWeb:
			if force || state.Web == nil || !processAlive(state.Web.PID) {
				state.Web = nil
			}
		}
		return nil
	})
}

func serviceStateByName(state tqconfig.State, name serviceName) *tqconfig.ServiceState {
	switch name {
	case serviceIssueTracker:
		return state.IssueTracker
	case serviceOrchestrator:
		return state.Orchestrator
	case serviceWeb:
		return state.Web
	default:
		return nil
	}
}

func statusForService(name serviceName, service *tqconfig.ServiceState) serviceStatus {
	status := serviceStatus{Name: string(name), State: "stopped"}
	if service == nil || !processAlive(service.PID) {
		return status
	}
	status.State = "running"
	status.PID = service.PID
	status.Addr = service.Addr
	status.Port = servicePort(service.Addr)
	status.StartedAt = service.StartedAt.Format(time.RFC3339)
	status.Uptime = time.Since(service.StartedAt).Round(time.Second).String()
	return status
}

func servicePort(addr string) int {
	addr = strings.TrimRight(strings.TrimSpace(addr), "/")
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		if parsed, err := http.NewRequest(http.MethodGet, addr, nil); err == nil && parsed.URL.Port() != "" {
			port, _ := strconv.Atoi(parsed.URL.Port())
			return port
		}
	}
	if index := strings.LastIndex(addr, ":"); index >= 0 {
		port, _ := strconv.Atoi(addr[index+1:])
		return port
	}
	return 0
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return process.Signal(syscall.Signal(0)) == nil
}

func isProcessDone(err error) bool {
	return errors.Is(err, os.ErrProcessDone)
}

func printServiceHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq service <action>")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Actions:")
	fmt.Fprintln(w, "  start    Start issue-tracker, orchestrator, and web")
	fmt.Fprintln(w, "  stop     Stop web, orchestrator, and issue-tracker")
	fmt.Fprintln(w, "  status   Show service status")
}
