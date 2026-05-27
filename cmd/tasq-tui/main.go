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

	"github.com/version-1/tasq/internal/task"
)

func main() {
	apiURL := flag.String("api", "http://localhost:8080", "orchestrator API base URL")
	watch := flag.Bool("watch", false, "refresh until interrupted")
	interval := flag.Duration("interval", 3*time.Second, "refresh interval when watch is enabled")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	for {
		if err := render(ctx, *apiURL); err != nil {
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
	fmt.Printf("Tasq Kanban  generated=%s\n", summary.GeneratedAt.Local().Format(time.RFC3339))
	fmt.Printf("Agents: %d active | concurrency=%d | poll=%ds\n\n", len(summary.Agents), summary.Settings.MaxConcurrentRuns, summary.Settings.PollIntervalSeconds)

	if len(summary.Agents) == 0 {
		fmt.Println("Agent status: no queued or running tasks")
	} else {
		fmt.Println("Agent status:")
		for _, item := range summary.Agents {
			fmt.Printf("  #%d %-18s %-22s %s\n", item.ID, item.AgentStatus, item.Title, item.Workspace)
		}
	}
	fmt.Println()

	for _, column := range summary.Columns {
		fmt.Printf("== %s (%d) ==\n", column.Title, len(column.Tasks))
		if len(column.Tasks) == 0 {
			fmt.Println("  -")
			fmt.Println()
			continue
		}
		for _, item := range column.Tasks {
			assignee := item.Assignee
			if assignee == "" {
				assignee = "unassigned"
			}
			fmt.Printf("  #%d [%s] %s\n", item.ID, item.Priority, item.Title)
			fmt.Printf("      agent=%s assignee=%s updated=%s\n", item.AgentStatus, assignee, item.UpdatedAt.Local().Format("01-02 15:04"))
			if item.LastError != "" {
				fmt.Printf("      error=%s\n", item.LastError)
			}
		}
		fmt.Println()
	}
	return nil
}

func fetchSummary(ctx context.Context, apiURL string) (task.Summary, error) {
	endpoint := strings.TrimRight(apiURL, "/") + "/api/v1/summary"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return task.Summary{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return task.Summary{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return task.Summary{}, fmt.Errorf("GET %s returned %s", endpoint, resp.Status)
	}
	var summary task.Summary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return task.Summary{}, err
	}
	return summary, nil
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
