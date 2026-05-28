package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/version-1/tasq/internal/orchestrator"
)

func main() {
	dbPath := flag.String("db", "tasq-orchestrator.sqlite", "SQLite database path")
	issueTrackerURL := flag.String("issue-tracker", "http://localhost:8080", "issue-tracker API base URL")
	orchestratorID := flag.String("id", "local-orchestrator", "orchestrator instance id")
	pollInterval := flag.Duration("poll-interval", 3*time.Second, "work item polling interval")
	leaseSeconds := flag.Int("lease-seconds", 30, "work item claim lease duration")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := orchestrator.Open(ctx, *dbPath)
	if err != nil {
		log.Fatalf("open orchestrator store: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("close orchestrator store: %v", err)
		}
	}()

	client := orchestrator.NewIssueTrackerClient(*issueTrackerURL)
	worker := orchestrator.NewWorker(store, client, *orchestratorID, *pollInterval, *leaseSeconds)
	log.Printf("orchestrator %s polling %s every %s", *orchestratorID, *issueTrackerURL, pollInterval.String())
	worker.Run(ctx)
}
