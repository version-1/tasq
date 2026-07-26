package tq

import (
	"fmt"
	"os"
	"strings"

	tqconfig "github.com/version-1/tasq/internal/config"
)

type configInfo struct {
	Version    string        `json:"version"`
	Commit     string        `json:"commit"`
	Profile    string        `json:"profile"`
	TQHome     *string       `json:"tqHome"`
	Home       string        `json:"home"`
	ConfigPath string        `json:"configPath"`
	Config     configSummary `json:"config"`
}

type configSummary struct {
	Author              string `json:"author"`
	MaxConcurrentAgents int    `json:"maxConcurrentAgents"`
}

func (a app) config(args []string, cfg config) error {
	if len(args) > 0 && (args[0] == "help" || args[0] == "-help" || args[0] == "--help") {
		printConfigHelp(a.stdout)
		return nil
	}
	if len(args) != 0 {
		return usageError("usage: tq config")
	}

	info, err := readConfigInfo()
	if err != nil {
		return err
	}
	if cfg.output == "json" {
		return writeJSON(a.stdout, info)
	}
	writeConfigInfo(a.stdout, info)
	return nil
}

func readConfigInfo() (configInfo, error) {
	profile, err := tqconfig.DefaultHomeProfile()
	if err != nil {
		return configInfo{}, err
	}
	home, err := tqconfig.Home()
	if err != nil {
		return configInfo{}, err
	}
	loaded, err := tqconfig.Load()
	if err != nil {
		return configInfo{}, err
	}
	version, commit := versionInfo()
	info := configInfo{
		Version:    version,
		Commit:     commit,
		Profile:    profile,
		Home:       home,
		ConfigPath: tqconfig.ConfigPath(home),
		Config: configSummary{
			Author:              loaded.Author,
			MaxConcurrentAgents: loaded.MaxConcurrentAgents,
		},
	}
	if value := strings.TrimSpace(os.Getenv(tqconfig.EnvHome)); value != "" {
		info.TQHome = &value
	}
	return info, nil
}

func writeConfigInfo(w interface{ Write([]byte) (int, error) }, info configInfo) {
	profile := info.Profile
	if profile == "" {
		profile = "default"
	}
	tqHome := "unset"
	if info.TQHome != nil {
		tqHome = *info.TQHome
	}
	fmt.Fprintf(w, "Version: %s\n", info.Version)
	fmt.Fprintf(w, "Commit: %s\n", info.Commit)
	fmt.Fprintf(w, "Profile: %s\n", profile)
	fmt.Fprintf(w, "TQ_HOME: %s\n", tqHome)
	fmt.Fprintf(w, "Home: %s\n", info.Home)
	fmt.Fprintf(w, "Config path: %s\n", info.ConfigPath)
	fmt.Fprintf(w, "Author: %s\n", info.Config.Author)
	fmt.Fprintf(w, "Max concurrent agents: %d\n", info.Config.MaxConcurrentAgents)
}
