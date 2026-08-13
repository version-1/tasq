package tq

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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

type serviceAddresses struct {
	issueTracker string
	orchestrator string
	web          string
}

type serviceStartLock struct {
	file *os.File
}

func (a app) routeService(ctx context.Context, args []string, cfg config) error {
	if len(args) == 0 || args[0] == "help" || args[0] == "-help" || args[0] == "--help" {
		printServiceHelp(a.stdout)
		return nil
	}
	action := args[0]
	if action == "stop" {
		if err := rejectManagedServiceMutation("stop"); err != nil {
			return err
		}
	}
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
	fs := newFlagSet("service start")
	yes := fs.Bool("y", false, "start without confirmation")
	if err := fs.Parse(args); err != nil {
		return usageError(err.Error())
	}
	if fs.NArg() != 0 {
		return usageError("usage: tq service start [-y]")
	}
	home, err := tqconfig.EnsureHome()
	if err != nil {
		return err
	}
	lock, err := acquireServiceStartLock(home)
	if err != nil {
		return err
	}
	defer lock.Close()
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
	if err := validateServiceExecutables(home); err != nil {
		return err
	}
	if err := checkMigrationTargetsNoPending(ctx); err != nil {
		return fmt.Errorf("migration pre-flight check failed: %w", err)
	}
	cliExecutable, err := prepareManagedCLI(home)
	if err != nil {
		return err
	}

	addresses, fallback, err := serviceStartAddresses()
	if err != nil {
		return err
	}
	if fallback {
		if !*yes {
			confirmed, err := a.confirmServicePortFallback(addresses)
			if err != nil {
				return err
			}
			if !confirmed {
				return errors.New("service start cancelled")
			}
		}
		if err := confirmServicePortsAvailable(addresses); err != nil {
			return err
		}
	}
	issueAddr := addresses.issueTracker
	orchestratorAddr := addresses.orchestrator
	webAddr := addresses.web
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
	if _, err := startManagedService(ctx, home, cliExecutable, issueService); err != nil {
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
			"-port", strconv.Itoa(servicePort(orchestratorAddr)),
		},
	}
	if _, err := startManagedService(ctx, home, cliExecutable, orchestratorService); err != nil {
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
	if _, err := startManagedService(ctx, home, cliExecutable, webService); err != nil {
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
	return writeServicesStarted(a.stdout)
}

func acquireServiceStartLock(home string) (*serviceStartLock, error) {
	path := filepath.Join(tqconfig.SystemDir(home), "service-start.lock")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open service start lock: %w", err)
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
			return nil, errors.New("service start is already in progress")
		}
		return nil, fmt.Errorf("lock service start: %w", err)
	}
	return &serviceStartLock{file: file}, nil
}

func (lock *serviceStartLock) Close() error {
	if lock == nil || lock.file == nil {
		return nil
	}
	unlockErr := syscall.Flock(int(lock.file.Fd()), syscall.LOCK_UN)
	closeErr := lock.file.Close()
	if unlockErr != nil {
		return fmt.Errorf("unlock service start: %w", unlockErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close service start lock: %w", closeErr)
	}
	return nil
}

func defaultServiceAddresses() serviceAddresses {
	return serviceAddresses{
		issueTracker: "127.0.0.1:" + strconv.Itoa(tqconfig.DefaultIssueTrackerPort),
		orchestrator: "127.0.0.1:" + strconv.Itoa(tqconfig.DefaultOrchestratorPort),
		web:          "127.0.0.1:" + strconv.Itoa(tqconfig.DefaultWebPort),
	}
}

func serviceStartAddresses() (serviceAddresses, bool, error) {
	addresses := defaultServiceAddresses()
	if err := confirmServicePortsAvailable(addresses); err == nil {
		return addresses, false, nil
	}

	fallback, err := allocateServiceAddresses()
	if err != nil {
		return serviceAddresses{}, false, fmt.Errorf("allocate fallback service ports: %w", err)
	}
	return fallback, true, nil
}

func allocateServiceAddresses() (serviceAddresses, error) {
	listeners := make([]net.Listener, 0, 3)
	for range 3 {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			for _, openListener := range listeners {
				_ = openListener.Close()
			}
			return serviceAddresses{}, err
		}
		listeners = append(listeners, listener)
	}
	addresses := serviceAddresses{
		issueTracker: listeners[0].Addr().String(),
		orchestrator: listeners[1].Addr().String(),
		web:          listeners[2].Addr().String(),
	}
	for _, listener := range listeners {
		if err := listener.Close(); err != nil {
			return serviceAddresses{}, err
		}
	}
	return addresses, nil
}

func confirmServicePortsAvailable(addresses serviceAddresses) error {
	listeners := make([]net.Listener, 0, 3)
	for _, address := range []string{addresses.issueTracker, addresses.orchestrator, addresses.web} {
		listener, err := net.Listen("tcp", address)
		if err != nil {
			for _, openListener := range listeners {
				_ = openListener.Close()
			}
			return fmt.Errorf("service port %s is unavailable: %w", address, err)
		}
		listeners = append(listeners, listener)
	}
	for index, listener := range listeners {
		if err := listener.Close(); err != nil {
			return fmt.Errorf("release service port %d: %w", index, err)
		}
	}
	return nil
}

func (a app) confirmServicePortFallback(addresses serviceAddresses) (bool, error) {
	if !interactiveInput(a.stdin) {
		return false, errors.New("service port fallback requires interactive confirmation; rerun with -y to continue")
	}
	fmt.Fprintln(a.stdout, "Default service ports are in use. The following loopback ports will be used:")
	fmt.Fprintf(a.stdout, "  issue-tracker: %s\n", addresses.issueTracker)
	fmt.Fprintf(a.stdout, "  orchestrator:  %s\n", addresses.orchestrator)
	fmt.Fprintf(a.stdout, "  web:           %s\n", addresses.web)
	fmt.Fprint(a.stdout, "Start services with these ports? [y/N] ")
	line, err := bufio.NewReader(a.stdin).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, fmt.Errorf("read confirmation: %w", err)
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}

func interactiveInput(input io.Reader) bool {
	file, ok := input.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
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
	return writeServicesStopped(a.stdout)
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
	return writeServiceStatuses(a.stdout, statuses)
}

func startManagedService(ctx context.Context, home string, cliExecutable string, service managedService) (int, error) {
	cmd, err := commandForService(ctx, home, cliExecutable, service)
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
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("start %s: %w", service.name, err)
	}
	return cmd.Process.Pid, nil
}

func commandForService(ctx context.Context, home string, cliExecutable string, service managedService) (*exec.Cmd, error) {
	executable := serviceExecutablePath(home, service.name)
	if err := validateServiceExecutable(service.name, executable); err != nil {
		return nil, err
	}
	command := exec.CommandContext(ctx, executable, service.args...)
	command.Env = serviceCommandEnv(home, cliExecutable, os.Environ())
	return command, nil
}

func validateServiceExecutables(home string) error {
	problems := make([]string, 0, 3)
	for _, name := range []serviceName{serviceIssueTracker, serviceOrchestrator, serviceWeb} {
		path := serviceExecutablePath(home, name)
		if err := validateServiceExecutable(name, path); err != nil {
			problems = append(problems, err.Error())
		}
	}
	if len(problems) == 0 {
		return nil
	}
	return fmt.Errorf("required service binaries are unavailable for resolved TQ_HOME %s:\n  %s\nReinstall Tasq with:\n  curl -fsSLO https://raw.githubusercontent.com/version-1/tasq/main/scripts/install.sh\n  TQ_HOME=%s sh install.sh", home, strings.Join(problems, "\n  "), home)
}

func validateServiceExecutable(name serviceName, path string) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("%s: missing (expected %s)", name, path)
	}
	if err != nil {
		return fmt.Errorf("%s: inspect expected %s: %w", name, path, err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s: not a regular file (expected %s)", name, path)
	}
	if info.Mode().Perm()&0o111 == 0 {
		return fmt.Errorf("%s: not executable (expected %s)", name, path)
	}
	return nil
}

func serviceExecutablePath(home string, name serviceName) string {
	return filepath.Join(tqconfig.SystemDir(home), "bin", string(name))
}

func serviceCommandEnv(home string, cliExecutable string, environment []string) []string {
	filtered := make([]string, 0, len(environment)+3)
	for _, entry := range environment {
		if !environmentEntryMatches(entry, tqconfig.EnvHome) &&
			!environmentEntryMatches(entry, tqconfig.EnvExecutable) &&
			!environmentEntryMatches(entry, tqconfig.EnvManagedRun) {
			filtered = append(filtered, entry)
		}
	}
	return append(filtered,
		tqconfig.EnvHome+"="+home,
		tqconfig.EnvExecutable+"="+cliExecutable,
		tqconfig.EnvManagedRun+"=1",
	)
}

func environmentEntryMatches(entry string, name string) bool {
	return strings.HasPrefix(entry, name+"=")
}

func currentExecutable() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("resolve tq executable: %w", err)
	}
	executable, err = filepath.Abs(executable)
	if err != nil {
		return "", fmt.Errorf("resolve absolute tq executable: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(executable); err == nil {
		executable = resolved
	}
	return filepath.Clean(executable), nil
}

func prepareManagedCLI(home string) (string, error) {
	source, err := currentExecutable()
	if err != nil {
		return "", err
	}
	return copyManagedCLI(home, source)
}

func copyManagedCLI(home string, source string) (string, error) {
	destination := filepath.Join(serviceInstallDir(home), "tq-managed")
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return "", fmt.Errorf("create managed tq directory: %w", err)
	}
	if err := installExecutable(source, destination); err != nil {
		return "", fmt.Errorf("prepare managed tq executable: %w", err)
	}
	return destination, nil
}

func rejectManagedServiceMutation(action string) error {
	if os.Getenv(tqconfig.EnvManagedRun) != "1" {
		return nil
	}
	return fmt.Errorf("tq service %s is unavailable inside an orchestrator-managed run; run it from a user shell", action)
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
	fmt.Fprintln(w, "  start    Start issue-tracker, orchestrator, and web (-y skips fallback port confirmation)")
	fmt.Fprintln(w, "  stop     Stop web, orchestrator, and issue-tracker")
	fmt.Fprintln(w, "  status   Show service status")
}
