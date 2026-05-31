package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/version-1/tasq/internal/orchestrator/httpserver"
	"github.com/version-1/tasq/internal/orchestrator/runner"
	"github.com/version-1/tasq/internal/orchestrator/runstore"
	"github.com/version-1/tasq/internal/orchestrator/tracker"
	"github.com/version-1/tasq/internal/orchestrator/worker"
	"github.com/version-1/tasq/internal/orchestrator/workflow"
)

func main() {
	dbPath := flag.String("db", "tasq-orchestrator.sqlite", "SQLite database path")
	issueTrackerURL := flag.String("issue-tracker", "http://localhost:8080", "issue-tracker API base URL")
	orchestratorID := flag.String("id", "local-orchestrator", "orchestrator instance id")
	pollInterval := flag.Duration("poll-interval", 0, "override work item polling interval")
	leaseSeconds := flag.Int("lease-seconds", 30, "work item claim lease duration")
	workflowPath := flag.String("workflow", "WORKFLOW.md", "Symphony workflow file path")
	httpPort := flag.Int("port", -1, "orchestrator HTTP server port; overrides workflow server.port when >= 0")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	store, err := runstore.Open(ctx, *dbPath)
	if err != nil {
		log.Fatalf("open orchestrator store: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("close orchestrator store: %v", err)
		}
	}()

	client := tracker.NewClient(*issueTrackerURL)
	definition, err := workflow.Load(*workflowPath)
	if err != nil {
		log.Fatalf("load workflow: %v", err)
	}
	if *pollInterval > 0 {
		definition.Config.PollInterval = *pollInterval
	}
	orchestratorWorker, err := worker.NewWithConfig(store, client, *orchestratorID, *leaseSeconds, definition, runner.CodexRunner{})
	if err != nil {
		log.Fatalf("create orchestrator worker: %v", err)
	}
	port := definition.Config.ServerPort
	if *httpPort >= 0 {
		port = *httpPort
	}
	if port >= 0 {
		address, err := httpserver.ListenAndServe(ctx, port, httpserver.New(store, orchestratorWorker).Handler())
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("start orchestrator http server: %v", err)
		}
		log.Printf("orchestrator http server listening on %s", address)
	}
	log.Printf("orchestrator %s polling %s every %s workspace_root=%s", *orchestratorID, *issueTrackerURL, definition.Config.PollInterval.String(), definition.Config.WorkspaceRoot)
	orchestratorWorker.Run(ctx)
}
