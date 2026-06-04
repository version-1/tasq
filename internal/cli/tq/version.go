package tq

import (
	"fmt"
	"io"
	"runtime/debug"
	"strings"
)

var (
	buildVersion = "dev"
	buildCommit  = "unknown"
)

func versionInfo() (version, commit string) {
	version = buildVersion
	commit = buildCommit

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return version, commit
	}

	if version == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" && !isPseudoVersion(info.Main.Version) {
		version = info.Main.Version
	}

	modified := false
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			if commit != "unknown" {
				continue
			}
			if len(setting.Value) > 7 {
				commit = setting.Value[:7]
			} else if setting.Value != "" {
				commit = setting.Value
			}
		case "vcs.modified":
			modified = strings.EqualFold(setting.Value, "true")
		}
	}
	if modified && commit != "unknown" {
		commit += "-dirty"
	}
	return version, commit
}

func printVersion(w io.Writer) {
	version, commit := versionInfo()
	fmt.Fprintf(w, "tq %s (commit: %s)\n", version, commit)
}

func isPseudoVersion(version string) bool {
	version = strings.TrimSuffix(version, "+dirty")
	parts := strings.Split(version, "-")
	if len(parts) < 3 {
		return false
	}
	timestamp := parts[len(parts)-2]
	revision := parts[len(parts)-1]
	return isDigits(timestamp, 14) && isHex(revision, 12)
}

func isDigits(value string, length int) bool {
	if len(value) != length {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func isHex(value string, length int) bool {
	if len(value) != length {
		return false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}
