package tq

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	tqconfig "github.com/version-1/tasq/internal/config"
)

func TestServiceStartLockHelper(t *testing.T) {
	if os.Getenv("TQ_SERVICE_START_LOCK_HELPER") != "1" {
		return
	}
	home, err := tqconfig.EnsureHome()
	if err != nil {
		t.Fatalf("ensure home: %v", err)
	}
	lock, err := acquireServiceStartLock(home)
	if err != nil {
		t.Fatalf("acquire service start lock: %v", err)
	}
	defer lock.Close()
	fmt.Fprintln(os.Stdout, "locked")
	_, _ = io.Copy(io.Discard, os.Stdin)
}

func TestServiceStartAddressesUsesFallbackWhenDefaultPortIsUnavailable(t *testing.T) {
	defaults := defaultServiceAddresses()
	listener, err := net.Listen("tcp", defaults.issueTracker)
	if err == nil {
		defer listener.Close()
	}

	addresses, fallback, err := serviceStartAddresses()
	if err != nil {
		t.Fatalf("service start addresses: %v", err)
	}
	if !fallback {
		t.Fatal("fallback = false, want true")
	}
	for _, address := range []string{addresses.issueTracker, addresses.orchestrator, addresses.web} {
		if !strings.HasPrefix(address, "127.0.0.1:") {
			t.Fatalf("address = %q, want loopback address", address)
		}
	}
	if addresses.issueTracker == defaults.issueTracker || addresses.orchestrator == defaults.orchestrator || addresses.web == defaults.web {
		t.Fatalf("fallback addresses = %+v, want all service ports replaced", addresses)
	}
}

func TestServiceCommandEnvSetsManagedExecutionContract(t *testing.T) {
	home := t.TempDir()
	cliExecutable := filepath.Join(t.TempDir(), "tqdev")
	environment := serviceCommandEnv(home, cliExecutable, []string{
		"PATH=/usr/bin",
		"TQ_HOME=/other/home",
		"TQ_HOME_SUFFIX=preserved",
		"TQ_HOME=/another/home",
		"TQ_EXECUTABLE=/old/tq",
		"TQ_MANAGED_RUN=0",
	})

	var homes []string
	for _, entry := range environment {
		if strings.HasPrefix(entry, tqconfig.EnvHome+"=") {
			homes = append(homes, entry)
		}
	}
	if len(homes) != 1 || homes[0] != tqconfig.EnvHome+"="+home {
		t.Fatalf("TQ_HOME entries = %v, want exactly %s=%s", homes, tqconfig.EnvHome, home)
	}
	if !containsString(environment, "TQ_HOME_SUFFIX=preserved") {
		t.Fatalf("environment = %v, want TQ_HOME_SUFFIX preserved", environment)
	}
	for _, want := range []string{
		tqconfig.EnvExecutable + "=" + cliExecutable,
		tqconfig.EnvManagedRun + "=1",
	} {
		if !containsString(environment, want) {
			t.Fatalf("environment = %v, want %q", environment, want)
		}
	}
}

func TestValidateServiceExecutablesReportsAllUnavailableTargets(t *testing.T) {
	home := t.TempDir()
	err := validateServiceExecutables(home)
	if err == nil {
		t.Fatal("validate service executables error = nil")
	}
	message := err.Error()
	for _, name := range []serviceName{serviceIssueTracker, serviceOrchestrator, serviceWeb} {
		path := serviceExecutablePath(home, name)
		for _, want := range []string{string(name) + ": missing", path} {
			if !strings.Contains(message, want) {
				t.Errorf("error = %q, want %q", message, want)
			}
		}
	}
	if !strings.Contains(message, "resolved TQ_HOME "+home) || !strings.Contains(message, "curl -fsSLO https://raw.githubusercontent.com/version-1/tasq/main/scripts/install.sh") || !strings.Contains(message, "TQ_HOME="+home+" sh install.sh") {
		t.Fatalf("error = %q, want home and reinstall instructions", message)
	}
}

func TestCommandForServiceUsesHomeSystemBinExecutable(t *testing.T) {
	home := t.TempDir()
	path := serviceExecutablePath(home, serviceIssueTracker)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create service bin dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("write service executable: %v", err)
	}
	cliExecutable := filepath.Join(t.TempDir(), "tqdev")
	command, err := commandForService(context.Background(), home, cliExecutable, managedService{name: serviceIssueTracker})
	if err != nil {
		t.Fatalf("command for service: %v", err)
	}
	if command.Path != path {
		t.Fatalf("command path = %q, want %q", command.Path, path)
	}
	if !containsString(command.Env, tqconfig.EnvHome+"="+home) {
		t.Fatalf("command environment does not contain resolved home: %v", command.Env)
	}
	if !containsString(command.Env, tqconfig.EnvExecutable+"="+cliExecutable) || !containsString(command.Env, tqconfig.EnvManagedRun+"=1") {
		t.Fatalf("command environment does not contain managed execution contract: %v", command.Env)
	}
}

func TestValidateServiceExecutableRejectsNonRegularOrNonExecutableFiles(t *testing.T) {
	home := t.TempDir()
	path := serviceExecutablePath(home, serviceWeb)
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("create non-regular service path: %v", err)
	}
	if err := validateServiceExecutable(serviceWeb, path); err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("directory validation error = %v, want non-regular file", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatalf("remove service directory: %v", err)
	}
	if err := os.WriteFile(path, []byte("service"), 0o644); err != nil {
		t.Fatalf("write non-executable service: %v", err)
	}
	if err := validateServiceExecutable(serviceWeb, path); err == nil || !strings.Contains(err.Error(), "not executable") {
		t.Fatalf("non-executable validation error = %v, want non-executable", err)
	}
}

func TestServiceInstallDirUsesResolvedHomeSystemBin(t *testing.T) {
	home := t.TempDir()
	if got, want := serviceInstallDir(home), filepath.Join(home, "system", "bin"); got != want {
		t.Fatalf("service install dir = %q, want %q", got, want)
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func TestConfirmServicePortsAvailableFailsWhenCandidateIsTaken(t *testing.T) {
	addresses, err := allocateServiceAddresses()
	if err != nil {
		t.Fatalf("allocate service addresses: %v", err)
	}
	listener, err := net.Listen("tcp", addresses.web)
	if err != nil {
		t.Fatalf("take candidate port: %v", err)
	}
	defer listener.Close()

	if err := confirmServicePortsAvailable(addresses); err == nil {
		t.Fatal("confirm service ports available succeeded after candidate was taken")
	}
}

func TestConfirmServicePortFallbackRejectsNonInteractiveInput(t *testing.T) {
	var output bytes.Buffer
	application := app{
		stdin:  strings.NewReader("yes\n"),
		stdout: &output,
	}

	confirmed, err := application.confirmServicePortFallback(serviceAddresses{
		issueTracker: "127.0.0.1:41001",
		orchestrator: "127.0.0.1:41002",
		web:          "127.0.0.1:41003",
	})
	if err == nil {
		t.Fatal("confirmation error = nil, want non-interactive input error")
	}
	if confirmed {
		t.Fatal("confirmed = true, want false")
	}
	if output.Len() != 0 {
		t.Fatalf("prompt output = %q, want empty", output.String())
	}
}

func TestServiceStartFailsWhenAnotherStartHoldsHomeLock(t *testing.T) {
	home := t.TempDir()
	t.Setenv(tqconfig.EnvHome, home)

	command := exec.Command(os.Args[0], "-test.run=^TestServiceStartLockHelper$")
	command.Env = append(os.Environ(), "TQ_SERVICE_START_LOCK_HELPER=1", tqconfig.EnvHome+"="+home)
	stdin, err := command.StdinPipe()
	if err != nil {
		t.Fatalf("create helper stdin: %v", err)
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatalf("create helper stdout: %v", err)
	}
	if err := command.Start(); err != nil {
		t.Fatalf("start lock helper: %v", err)
	}
	t.Cleanup(func() {
		_ = stdin.Close()
		_ = command.Wait()
	})
	line, err := bufio.NewReader(stdout).ReadString('\n')
	if err != nil {
		t.Fatalf("wait for lock helper: %v", err)
	}
	if line != "locked\n" {
		t.Fatalf("lock helper output = %q, want locked", line)
	}

	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	code := run(context.Background(), []string{"service", "start", "-y"}, strings.NewReader(""), &stdoutBuffer, &stderrBuffer)
	if code != 1 {
		t.Fatalf("code = %d, want 1; stderr=%s", code, stderrBuffer.String())
	}
	if got := decodeCLIError(t, stderrBuffer.String()); got != "service start is already in progress" {
		t.Fatalf("error = %q, want service start lock error", got)
	}
	if stdoutBuffer.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdoutBuffer.String())
	}
	for _, logName := range []string{"issue-tracker.log", "orchestrator.log", "web.log"} {
		if _, err := os.Stat(filepath.Join(tqconfig.LogDir(home), logName)); !os.IsNotExist(err) {
			t.Fatalf("service log exists while start lock held: %s, err=%v", logName, err)
		}
	}
}
