package tq

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	tqconfig "github.com/version-1/tasq/internal/config"
)

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
