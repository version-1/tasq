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
	"github.com/version-1/tasq/internal/orchestrator/runstore"
	"github.com/version-1/tasq/internal/orchestrator/workflow"
)

func main() {
	dbPath := flag.String("db", "tasq-orchestrator.sqlite", "SQLite database path")
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

	definition, err := workflow.Load(*workflowPath)
	if err != nil {
		log.Fatalf("load workflow: %v", err)
	}
	port := definition.Config.ServerPort
	if *httpPort >= 0 {
		port = *httpPort
	}
	if port >= 0 {
		address, err := httpserver.ListenAndServe(ctx, port, httpserver.New(store, nil).Handler())
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("start orchestrator http server: %v", err)
		}
		log.Printf("orchestrator http server listening on %s", address)
	} else {
		log.Printf("orchestrator http server disabled; set --port or workflow server.port to enable")
	}
	<-ctx.Done()
}
