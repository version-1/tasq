package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	tqconfig "github.com/version-1/tasq/internal/config"
	"github.com/version-1/tasq/internal/issue/api"
	"github.com/version-1/tasq/internal/issue/store"
	"github.com/version-1/tasq/internal/orchestrator/runstore"
)

func main() {
	addr := flag.String("addr", ":"+strconv.Itoa(tqconfig.DefaultIssueTrackerPort), "HTTP listen address")
	dbPath := flag.String("db", "", "SQLite database path")
	flag.Parse()

	ctx := context.Background()
	home, err := tqconfig.EnsureHome()
	if err != nil {
		log.Fatalf("ensure TQ_HOME: %v", err)
	}
	resolvedDBPath := *dbPath
	if resolvedDBPath == "" {
		resolvedDBPath = tqconfig.IssueTrackerDBPath(home)
	}
	issueStore, err := store.Open(ctx, resolvedDBPath)
	if err != nil {
		log.Fatalf("open issue store: %v", err)
	}
	defer func() {
		if err := issueStore.Close(); err != nil {
			log.Printf("close issue store: %v", err)
		}
	}()
	orchestratorStore, err := runstore.Open(ctx, tqconfig.OrchestratorDBPath(home))
	if err != nil {
		log.Fatalf("open orchestrator store: %v", err)
	}
	defer func() {
		if err := orchestratorStore.Close(); err != nil {
			log.Printf("close orchestrator store: %v", err)
		}
	}()

	server := &http.Server{
		Handler:           api.NewServerWithOrchestratorStore(issueStore, orchestratorStore).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	serviceAddr := clientAddr(listener.Addr().String())
	serviceState, err := tqconfig.NewCurrentServiceState(serviceAddr, resolvedDBPath)
	if err != nil {
		log.Fatalf("identify issue-tracker process: %v", err)
	}
	if err := tqconfig.UpdateState(func(state *tqconfig.State) error {
		state.IssueTracker = &serviceState
		return nil
	}); err != nil {
		log.Fatalf("write issue-tracker state: %v", err)
	}
	defer func() {
		if err := tqconfig.UpdateState(func(state *tqconfig.State) error {
			if state.IssueTracker != nil && state.IssueTracker.HasSameProcessIdentity(serviceState) {
				state.IssueTracker = nil
			}
			return nil
		}); err != nil {
			log.Printf("clear issue-tracker state: %v", err)
		}
	}()

	go func() {
		log.Printf("issue-tracker listening on %s db=%s", serviceAddr, resolvedDBPath)
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
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

func clientAddr(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	if host == "" || host == "::" || host == "[::]" || host == "0.0.0.0" {
		host = "127.0.0.1"
	}
	return net.JoinHostPort(host, port)
}
