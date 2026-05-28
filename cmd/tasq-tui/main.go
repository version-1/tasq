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

	issue "github.com/version-1/tasq/internal/issue/domain/entity"
)

func main() {
	apiURL := flag.String("api", "http://localhost:8080", "issue-tracker API base URL")
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
	fmt.Printf("Tasq Issues  generated=%s\n", summary.GeneratedAt.Local().Format(time.RFC3339))
	fmt.Printf("Runs: %d tracked\n\n", len(summary.Runs))

	if len(summary.Runs) == 0 {
		fmt.Println("Run status: no orchestrator runs")
	} else {
		fmt.Println("Run status:")
		for _, run := range summary.Runs {
			fmt.Printf("  issue=%d work_item=%d %-18s %s\n", run.IssueID, run.WorkItemID, run.Status, run.Workspace)
		}
	}
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
			runStatus := "no_run"
			if item.Run != nil {
				runStatus = string(item.Run.Status)
			}
			fmt.Printf("      issue=%s run=%s assignee=%s updated=%s\n", item.Status, runStatus, assignee, item.UpdatedAt.Local().Format("01-02 15:04"))
			if item.Run != nil && item.Run.Error != "" {
				fmt.Printf("      error=%s\n", item.Run.Error)
			}
		}
		fmt.Println()
	}
	return nil
}

func fetchSummary(ctx context.Context, apiURL string) (issue.Summary, error) {
	endpoint := strings.TrimRight(apiURL, "/") + "/api/v1/summary"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return issue.Summary{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return issue.Summary{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return issue.Summary{}, fmt.Errorf("GET %s returned %s", endpoint, resp.Status)
	}
	var summary issue.Summary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		return issue.Summary{}, err
	}
	return summary, nil
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
