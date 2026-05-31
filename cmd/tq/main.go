package main

import (
	"context"
	"os"

	tqcli "github.com/version-1/tasq/internal/cli/tq"
)

func main() {
	os.Exit(tqcli.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
}
