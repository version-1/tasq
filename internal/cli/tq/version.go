package tq

import (
	"fmt"
	"io"

	"github.com/version-1/tasq/internal/buildinfo"
)

func versionInfo() (version, commit string) {
	info := buildinfo.Current()
	return info.Version, info.Commit
}

func printVersion(w io.Writer) {
	version, commit := versionInfo()
	fmt.Fprintf(w, "tq %s (commit: %s)\n", version, commit)
}
