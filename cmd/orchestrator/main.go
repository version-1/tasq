package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	tqconfig "github.com/version-1/tasq/internal/config"
	"github.com/version-1/tasq/internal/orchestrator/coordinator"
	"github.com/version-1/tasq/internal/orchestrator/httpserver"
	"github.com/version-1/tasq/internal/orchestrator/runner"
	"github.com/version-1/tasq/internal/orchestrator/runstore"
	"github.com/version-1/tasq/internal/orchestrator/tracker"
	"github.com/version-1/tasq/internal/orchestrator/workflow"
	"github.com/version-1/tasq/internal/orchestrator/workspace"
)

func main() {
	dbPath := flag.String("db", "", "SQLite database path")
	workflowPath := flag.String("workflow", "WORKFLOW.md", "Symphony workflow file path")
	issueTrackerURL := flag.String("issue-tracker", "", "issue-tracker API base URL; enables polling when set")
	httpPort := flag.Int("port", -1, "orchestrator HTTP server port; overrides workflow server.port when >= 0")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	home, err := tqconfig.EnsureHome()
	if err != nil {
		log.Fatalf("ensure TQ_HOME: %v", err)
	}
	appConfig, err := tqconfig.Load()
	if err != nil {
		log.Fatalf("load TQ_HOME config: %v", err)
	}
	resolvedDBPath := *dbPath
	if resolvedDBPath == "" {
		resolvedDBPath = tqconfig.OrchestratorDBPath(home)
	}
	store, err := runstore.Open(ctx, resolvedDBPath)
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
	effectiveMaxConcurrentRuns := min(definition.Config.MaxConcurrentRuns, appConfig.MaxConcurrentAgents)
	resolvedIssueTrackerURL := *issueTrackerURL
	if resolvedIssueTrackerURL == "" {
		if apiURL, ok, err := tqconfig.IssueTrackerURLFromState(); err != nil {
			log.Fatalf("read issue-tracker state: %v", err)
		} else if ok {
			resolvedIssueTrackerURL = apiURL
		}
	}
	var refresher httpserver.Refresher
	var dispatcher *coordinator.Dispatcher
	if resolvedIssueTrackerURL != "" {
		workspaceManager, err := workspace.NewManagerWithHooks(definition.Config.WorkspaceRoot, workspace.HookConfig{
			AfterCreate:  definition.Config.HookAfterCreate,
			BeforeRun:    definition.Config.HookBeforeRun,
			AfterRun:     definition.Config.HookAfterRun,
			BeforeRemove: definition.Config.HookBeforeRemove,
			Timeout:      definition.Config.HookTimeout,
		})
		if err != nil {
			log.Fatalf("create workspace manager: %v", err)
		}
		if err := workspaceManager.Prune(); err != nil {
			log.Fatalf("prune workspace manager: %v", err)
		}
		trackerClient := tracker.NewClient(resolvedIssueTrackerURL)
		dispatcher, err = coordinator.NewDispatcher(coordinator.DispatcherConfig{
			Tracker:           trackerClient,
			Store:             store,
			Runner:            runner.CodexRunner{},
			WorkflowConfig:    definition.Config,
			PromptTemplate:    definition.PromptTemplate,
			MaxConcurrentRuns: effectiveMaxConcurrentRuns,
		})
		if err != nil {
			log.Fatalf("create orchestrator dispatcher: %v", err)
		}
		poller, err := coordinator.NewPoller(coordinator.PollerConfig{
			Tracker:        trackerClient,
			Store:          store,
			Workspaces:     workspaceManager,
			Dispatcher:     dispatcher,
			Interval:       definition.Config.PollInterval,
			MaxActiveRuns:  effectiveMaxConcurrentRuns,
			OrchestratorID: "tasq-orchestrator",
		})
		if err != nil {
			log.Fatalf("create orchestrator poller: %v", err)
		}
		poller.Start(ctx)
		refresher = poller
		log.Printf("orchestrator polling issue-tracker=%s interval=%s max_concurrent=%d", resolvedIssueTrackerURL, definition.Config.PollInterval, effectiveMaxConcurrentRuns)
	} else {
		log.Printf("orchestrator polling disabled; set --issue-tracker or start issue-tracker with TQ_HOME state")
	}
	port := definition.Config.ServerPort
	if *httpPort >= 0 {
		port = *httpPort
	}
	orchestratorAddr := ""
	if port >= 0 {
		address, err := httpserver.ListenAndServe(ctx, port, httpserver.New(store, refresher).Handler())
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("start orchestrator http server: %v", err)
		}
		orchestratorAddr = address
		log.Printf("orchestrator http server listening on %s", address)
	} else {
		log.Printf("orchestrator http server disabled; set --port or workflow server.port to enable")
	}
	if err := tqconfig.UpdateState(func(state *tqconfig.State) error {
		state.Orchestrator = &tqconfig.ServiceState{
			PID:       os.Getpid(),
			Addr:      orchestratorAddr,
			DB:        resolvedDBPath,
			StartedAt: time.Now().UTC(),
		}
		return nil
	}); err != nil {
		log.Fatalf("write orchestrator state: %v", err)
	}
	defer func() {
		if err := tqconfig.UpdateState(func(state *tqconfig.State) error {
			state.Orchestrator = nil
			return nil
		}); err != nil {
			log.Printf("clear orchestrator state: %v", err)
		}
	}()
	<-ctx.Done()
	if dispatcher != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := dispatcher.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown orchestrator dispatcher: %v", err)
		}
	}
}
