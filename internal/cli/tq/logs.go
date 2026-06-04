package tq

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	tqconfig "github.com/version-1/tasq/internal/config"
)

const defaultLogLines = 1000

type logService struct {
	name string
	file string
}

func (a app) routeLogs(ctx context.Context, args []string, cfg config) error {
	if len(args) == 0 {
		return usageError("usage: tq logs <service> [-n lines] [-f]")
	}
	if args[0] == "help" || args[0] == "-help" || args[0] == "--help" {
		printLogsHelp(a.stdout)
		return nil
	}
	if cfg.output == "json" {
		return usageError("logs does not support json output")
	}

	service, remaining, err := parseLogsService(args)
	if err != nil {
		return err
	}
	fs := newFlagSet("logs")
	lines := fs.Int("n", defaultLogLines, "number of lines from the end")
	follow := fs.Bool("f", false, "follow appended log output")
	if err := fs.Parse(remaining); err != nil {
		return usageError(err.Error())
	}
	if fs.NArg() != 0 {
		return usageError("logs does not accept extra positional arguments")
	}
	if *lines < 0 {
		return usageError("lines must be non-negative")
	}

	home, err := tqconfig.Home()
	if err != nil {
		return err
	}
	path := filepath.Join(tqconfig.LogDir(home), service.file)
	return writeServiceLog(ctx, a.stdout, path, *lines, *follow)
}

func parseLogsService(args []string) (logService, []string, error) {
	if len(args) == 0 {
		return logService{}, nil, usageError("usage: tq logs <service> [-n lines] [-f]")
	}
	service, err := resolveLogService(args[0])
	if err != nil {
		return logService{}, nil, err
	}
	return service, args[1:], nil
}

func resolveLogService(value string) (logService, error) {
	switch strings.TrimSpace(value) {
	case "tracker", "issue-tracker":
		return logService{name: "issue-tracker", file: "issue-tracker.log"}, nil
	case "orchestrator":
		return logService{name: "orchestrator", file: "orchestrator.log"}, nil
	case "web":
		return logService{name: "web", file: "web.log"}, nil
	default:
		return logService{}, usageError("unknown log service %q", value)
	}
}

func writeServiceLog(ctx context.Context, w io.Writer, path string, lines int, follow bool) error {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("log file not found: %s", path)
		}
		return fmt.Errorf("open log file: %w", err)
	}
	defer file.Close()

	if lines > 0 {
		if err := writeLastLogLines(w, file, lines); err != nil {
			return err
		}
	}
	if !follow {
		return nil
	}
	offset, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("seek log file: %w", err)
	}
	return followLog(ctx, w, file, offset, 200*time.Millisecond)
}

func writeLastLogLines(w io.Writer, file *os.File, lines int) error {
	raw, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("read log file: %w", err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek log file: %w", err)
	}
	_, err = w.Write(lastLogLines(raw, lines))
	return err
}

func lastLogLines(raw []byte, lines int) []byte {
	if lines <= 0 || len(raw) == 0 {
		return nil
	}
	parts := strings.SplitAfter(string(raw), "\n")
	if parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	if lines >= len(parts) {
		return raw
	}
	return []byte(strings.Join(parts[len(parts)-lines:], ""))
}

func followLog(ctx context.Context, w io.Writer, file *os.File, offset int64, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			nextOffset, err := copyAppendedLog(w, file, offset)
			if err != nil {
				return err
			}
			offset = nextOffset
		}
	}
}

func copyAppendedLog(w io.Writer, file *os.File, offset int64) (int64, error) {
	info, err := file.Stat()
	if err != nil {
		return offset, fmt.Errorf("stat log file: %w", err)
	}
	size := info.Size()
	if size < offset {
		offset = 0
	}
	if size == offset {
		return offset, nil
	}
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return offset, fmt.Errorf("seek log file: %w", err)
	}
	if _, err := io.CopyN(w, file, size-offset); err != nil {
		return offset, fmt.Errorf("read appended log: %w", err)
	}
	return size, nil
}

func printLogsHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq logs <service> [-n lines] [-f]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Services:")
	fmt.Fprintln(w, "  tracker, issue-tracker")
	fmt.Fprintln(w, "  orchestrator")
	fmt.Fprintln(w, "  web")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  -n lines  Number of lines from the end (default 1000)")
	fmt.Fprintln(w, "  -f        Follow appended log output")
}
