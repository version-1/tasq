package workspace

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

var ErrHookTimeout = errors.New("workspace hook timed out")

type HookConfig struct {
	AfterCreate  string
	BeforeRun    string
	AfterRun     string
	BeforeRemove string
	Timeout      time.Duration
}

func (c HookConfig) timeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return 60 * time.Second
}

func RunHook(script string, workspacePath string, timeout time.Duration) error {
	if strings.TrimSpace(script) == "" {
		return nil
	}
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	log.Printf("workspace hook start cwd=%s timeout=%s", workspacePath, timeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-lc", script)
	cmd.Dir = workspacePath
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start workspace hook: %w", err)
	}
	err := cmd.Wait()
	if ctx.Err() == context.DeadlineExceeded {
		killProcessGroup(cmd.Process.Pid)
		log.Printf("workspace hook timeout cwd=%s timeout=%s", workspacePath, timeout)
		return fmt.Errorf("%w after %s", ErrHookTimeout, timeout)
	}
	if err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = err.Error()
		}
		log.Printf("workspace hook failed cwd=%s: %s", workspacePath, message)
		return fmt.Errorf("workspace hook failed: %s", message)
	}
	return nil
}

func killProcessGroup(pid int) {
	if pid <= 0 {
		return
	}
	_ = syscall.Kill(-pid, syscall.SIGKILL)
}
