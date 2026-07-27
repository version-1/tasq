package main

import (
	"os"

	tqcli "github.com/version-1/tasq/internal/cli/tq"
)

func main() {
	os.Exit(tqcli.RunStory(os.Args[1:], os.Stdout, os.Stderr))
}
