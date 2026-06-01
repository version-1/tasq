package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultMaxConcurrentAgents = 10

type Config struct {
	Author              string `yaml:"author"`
	MaxConcurrentAgents int    `yaml:"max_concurrent_agents"`
}

type Overrides struct {
	Author              *string
	MaxConcurrentAgents *int
}

func DefaultConfig() Config {
	return Config{
		Author:              strings.TrimSpace(os.Getenv("USER")),
		MaxConcurrentAgents: defaultMaxConcurrentAgents,
	}
}

func Load() (Config, error) {
	config := DefaultConfig()
	home, err := Home()
	if err != nil {
		return Config{}, err
	}
	raw, err := os.ReadFile(ConfigPath(home))
	if errors.Is(err, os.ErrNotExist) {
		return config, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config.yaml: %w", err)
	}
	if err := yaml.Unmarshal(raw, &config); err != nil {
		return Config{}, fmt.Errorf("parse config.yaml: %w", err)
	}
	if config.MaxConcurrentAgents <= 0 {
		config.MaxConcurrentAgents = defaultMaxConcurrentAgents
	}
	return config, nil
}

func (c Config) Resolve(overrides Overrides) Config {
	resolved := c
	if resolved.MaxConcurrentAgents <= 0 {
		resolved.MaxConcurrentAgents = defaultMaxConcurrentAgents
	}
	if overrides.Author != nil {
		resolved.Author = *overrides.Author
	}
	if overrides.MaxConcurrentAgents != nil {
		resolved.MaxConcurrentAgents = *overrides.MaxConcurrentAgents
	}
	return resolved
}
