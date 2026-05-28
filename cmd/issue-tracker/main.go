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

	"github.com/version-1/tasq/internal/issueapi"
	"github.com/version-1/tasq/internal/issuestore"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	dbPath := flag.String("db", "tasq-issues.sqlite", "SQLite database path")
	flag.Parse()

	ctx := context.Background()
	store, err := issuestore.Open(ctx, *dbPath)
	if err != nil {
		log.Fatalf("open issue store: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.Printf("close issue store: %v", err)
		}
	}()

	server := &http.Server{
		Addr:              *addr,
		Handler:           issueapi.NewServer(store).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("issue-tracker listening on %s", *addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown server: %v", err)
	}
}
