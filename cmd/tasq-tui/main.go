package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	tqconfig "github.com/version-1/tasq/internal/config"
	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func main() {
	apiURL := flag.String("api", "", "issue-tracker API base URL")
	watch := flag.Bool("watch", false, "refresh until interrupted")
	interval := flag.Duration("interval", 3*time.Second, "refresh interval when watch is enabled")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	resolvedAPIURL := *apiURL
	if resolvedAPIURL == "" {
		if stateURL, ok, err := tqconfig.IssueTrackerURLFromState(); err != nil {
			fmt.Fprintf(os.Stderr, "tasq-tui: %v\n", err)
			os.Exit(1)
		} else if ok {
			resolvedAPIURL = stateURL
		}
	}
	if resolvedAPIURL == "" {
		resolvedAPIURL = "http://localhost:8080"
	}
	for {
		if err := render(ctx, resolvedAPIURL); err != nil {
			fmt.Fprintf(os.Stderr, "tasq-tui: %v\n", err)
			os.Exit(1)
		}
		if !*watch {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(*interval):
		}
	}
}

func render(ctx context.Context, apiURL string) error {
	summary, err := fetchSummary(ctx, apiURL)
	if err != nil {
		return err
	}

	clearScreen()
	fmt.Printf("Tasq Issues  generated=%s\n", summary.GeneratedAt.Local().Format(time.RFC3339))
	fmt.Println()

	for _, column := range summary.Columns {
		fmt.Printf("== %s (%d) ==\n", column.Title, len(column.Issues))
		if len(column.Issues) == 0 {
			fmt.Println("  -")
			fmt.Println()
			continue
		}
		for _, item := range column.Issues {
			assignee := item.Assignee
			if assignee == "" {
				assignee = "unassigned"
			}
			fmt.Printf("  #%d [%s] %s\n", item.ID, item.Priority, item.Title)
			fmt.Printf("      issue=%s assignee=%s updated=%s\n", item.Status, assignee, item.UpdatedAt.Local().Format("01-02 15:04"))
		}
		fmt.Println()
	}
	return nil
}

func fetchSummary(ctx context.Context, apiURL string) (entity.Summary, error) {
	endpoint := strings.TrimRight(apiURL, "/") + "/api/v1/summary"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return entity.Summary{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return entity.Summary{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return entity.Summary{}, fmt.Errorf("GET %s returned %s", endpoint, resp.Status)
	}
	var payload struct {
		Data entity.Summary `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return entity.Summary{}, err
	}
	return payload.Data, nil
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
