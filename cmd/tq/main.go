package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	tqcli "github.com/version-1/tasq/internal/cli/tq"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(tqcli.Run(ctx, os.Args[1:], os.Stdout, os.Stderr))
}
