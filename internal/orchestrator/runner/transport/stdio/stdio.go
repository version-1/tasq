package stdio

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

type Connection struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
	stderr io.Reader
	done   chan error
	mu     sync.Mutex
}

func Start(ctx context.Context, command string, cwd string) (*Connection, error) {
	cmd := exec.CommandContext(ctx, "bash", "-lc", command)
	cmd.Dir = cwd
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("create stderr pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start codex app-server: %w", err)
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	conn := &Connection{
		cmd:    cmd,
		stdin:  stdin,
		stdout: scanner,
		stderr: stderr,
		done:   make(chan error, 1),
	}
	go func() {
		conn.done <- cmd.Wait()
		close(conn.done)
	}()
	return conn, nil
}

func (c *Connection) Identifier() string {
	if c.cmd.Process == nil {
		return "pid=0"
	}
	return fmt.Sprintf("pid=%d", c.cmd.Process.Pid)
}

func (c *Connection) FrameSource() string {
	return "stdout"
}

func (c *Connection) Send(_ context.Context, frame []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.stdin.Write(append(frame, '\n')); err != nil {
		return fmt.Errorf("write app-server request: %w", err)
	}
	return nil
}

func (c *Connection) Receive(_ context.Context) ([]byte, error) {
	if c.stdout.Scan() {
		return append([]byte(nil), c.stdout.Bytes()...), nil
	}
	if err := c.stdout.Err(); err != nil {
		return nil, err
	}
	return nil, io.EOF
}

func (c *Connection) Close() error {
	_ = c.stdin.Close()
	if c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
	}
	<-c.done
	return nil
}

func (c *Connection) Done() <-chan error {
	return c.done
}

func (c *Connection) Stderr() io.Reader {
	return c.stderr
}
