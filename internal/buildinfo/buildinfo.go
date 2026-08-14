package buildinfo

import (
	"runtime/debug"
	"strings"
)

var (
	version = "dev"
	commit  = "unknown"
)

type Info struct {
	Version string
	Commit  string
}

func Current() Info {
	info := Info{Version: version, Commit: commit}
	build, ok := debug.ReadBuildInfo()
	if !ok {
		return info
	}
	if info.Version == "dev" && build.Main.Version != "" && build.Main.Version != "(devel)" && !isPseudoVersion(build.Main.Version) {
		info.Version = build.Main.Version
	}

	modified := false
	for _, setting := range build.Settings {
		switch setting.Key {
		case "vcs.revision":
			if info.Commit != "unknown" {
				continue
			}
			if len(setting.Value) > 7 {
				info.Commit = setting.Value[:7]
			} else if setting.Value != "" {
				info.Commit = setting.Value
			}
		case "vcs.modified":
			modified = strings.EqualFold(setting.Value, "true")
		}
	}
	if modified && info.Commit != "unknown" {
		info.Commit += "-dirty"
	}
	return info
}

func isPseudoVersion(value string) bool {
	value = strings.TrimSuffix(value, "+dirty")
	parts := strings.Split(value, "-")
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
