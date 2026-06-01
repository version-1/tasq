package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/version-1/tasq/internal/orchestrator/coordinator"
	"github.com/version-1/tasq/internal/orchestrator/httpserver"
	"github.com/version-1/tasq/internal/orchestrator/runstore"
	"github.com/version-1/tasq/internal/orchestrator/tracker"
	"github.com/version-1/tasq/internal/orchestrator/workflow"
	"github.com/version-1/tasq/internal/orchestrator/workspace"
)

func main() {
	dbPath := flag.String("db", "tasq-orchestrator.sqlite", "SQLite database path")
	workflowPath := flag.String("workflow", "WORKFLOW.md", "Symphony workflow file path")
	issueTrackerURL := flag.String("issue-tracker", "", "issue-tracker API base URL; enables polling when set")
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

	definition, err := workflow.Load(*workflowPath)
	if err != nil {
		log.Fatalf("load workflow: %v", err)
	}
	var refresher httpserver.Refresher
	if *issueTrackerURL != "" {
		workspaceManager, err := workspace.NewManagerWithSourceAndHooks(definition.Config.WorkspaceRoot, definition.Config.WorkspaceSource, workspace.HookConfig{
			AfterCreate:  definition.Config.HookAfterCreate,
			BeforeRun:    definition.Config.HookBeforeRun,
			AfterRun:     definition.Config.HookAfterRun,
			BeforeRemove: definition.Config.HookBeforeRemove,
			Timeout:      definition.Config.HookTimeout,
		})
		if err != nil {
			log.Fatalf("create workspace manager: %v", err)
		}
		poller, err := coordinator.NewPoller(coordinator.PollerConfig{
			Tracker:        tracker.NewClient(*issueTrackerURL),
			Store:          store,
			Workspaces:     workspaceManager,
			Interval:       definition.Config.PollInterval,
			MaxActiveRuns:  definition.Config.MaxConcurrentRuns,
			OrchestratorID: "tasq-orchestrator",
		})
		if err != nil {
			log.Fatalf("create orchestrator poller: %v", err)
		}
		poller.Start(ctx)
		refresher = poller
		log.Printf("orchestrator polling issue-tracker=%s interval=%s", *issueTrackerURL, definition.Config.PollInterval)
	} else {
		log.Printf("orchestrator polling disabled; set --issue-tracker to enable")
	}
	port := definition.Config.ServerPort
	if *httpPort >= 0 {
		port = *httpPort
	}
	if port >= 0 {
		address, err := httpserver.ListenAndServe(ctx, port, httpserver.New(store, refresher).Handler())
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("start orchestrator http server: %v", err)
		}
		log.Printf("orchestrator http server listening on %s", address)
	} else {
		log.Printf("orchestrator http server disabled; set --port or workflow server.port to enable")
	}
	<-ctx.Done()
}
