package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHomeUsesTQHome(t *testing.T) {
	originalProfile := defaultHomeProfile
	defaultHomeProfile = "preview"
	t.Cleanup(func() { defaultHomeProfile = originalProfile })
	dir := t.TempDir()
	t.Setenv(EnvHome, filepath.Join(dir, "relative"))
	home, err := Home()
	if err != nil {
		t.Fatalf("home: %v", err)
	}
	if !filepath.IsAbs(home) {
		t.Fatalf("home is not absolute: %s", home)
	}
	if home != filepath.Join(dir, "relative") {
		t.Fatalf("home=%q", home)
	}
}

func TestHomeUsesDefaultProfile(t *testing.T) {
	originalProfile := defaultHomeProfile
	defaultHomeProfile = "preview"
	t.Cleanup(func() { defaultHomeProfile = originalProfile })
	t.Setenv(EnvHome, "")

	home, err := Home()
	if err != nil {
		t.Fatalf("home: %v", err)
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("user home: %v", err)
	}
	if want := filepath.Join(userHome, ".tasq-preview"); home != want {
		t.Fatalf("home=%q, want %q", home, want)
	}
}

func TestDefaultHomeProfileRejectsInvalidValue(t *testing.T) {
	originalProfile := defaultHomeProfile
	defaultHomeProfile = "Preview"
	t.Cleanup(func() { defaultHomeProfile = originalProfile })

	if _, err := DefaultHomeProfile(); err == nil {
		t.Fatal("DefaultHomeProfile() error = nil")
	}
}

func TestEnsureHomeCreatesLayout(t *testing.T) {
	home := filepath.Join(t.TempDir(), ".tasq")
	t.Setenv(EnvHome, home)
	resolved, err := EnsureHome()
	if err != nil {
		t.Fatalf("ensure home: %v", err)
	}
	if resolved != home {
		t.Fatalf("resolved=%q, want %q", resolved, home)
	}
	for _, path := range []string{ConfigDir(home), SystemDir(home), DataDir(home), LogDir(home)} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if !info.IsDir() {
			t.Fatalf("%s is not dir", path)
		}
	}
	content, err := os.ReadFile(WorkflowPath(home))
	if err != nil {
		t.Fatalf("read workflow: %v", err)
	}
	if string(content) != DefaultWorkflowTemplate() {
		t.Fatalf("workflow content=%q", content)
	}
}

func TestEnsureHomePreservesExistingWorkflow(t *testing.T) {
	home := filepath.Join(t.TempDir(), ".tasq")
	t.Setenv(EnvHome, home)
	const existing = "custom workflow\n"
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatalf("mkdir home: %v", err)
	}
	if err := os.WriteFile(WorkflowPath(home), []byte(existing), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}
	if _, err := EnsureHome(); err != nil {
		t.Fatalf("ensure home: %v", err)
	}
	content, err := os.ReadFile(WorkflowPath(home))
	if err != nil {
		t.Fatalf("read workflow: %v", err)
	}
	if string(content) != existing {
		t.Fatalf("workflow content=%q, want %q", content, existing)
	}
}

func TestDefaultWorkflowTemplateIncludesAgentTaskInstructions(t *testing.T) {
	template := DefaultWorkflowTemplate()
	for _, want := range []string{
		"  root: .worktrees/agents\n",
		"  max_concurrent_agents: 5\n",
		"  command: codex --sandbox workspace-write app-server\n",
		"  read_timeout_ms: 15000\n",
		"# Task\n",
		"Issue ID: {{ issue.id }}\n",
		"Write agent plan files, pull request summary drafts, and other temporary\n",
		"temporary artifacts under `~/codex`, `$CODEX_HOME`, or other external home\n",
		"2. Before branch creation, run `git fetch origin`, then create the task branch from `origin/main` or the repository's default branch.\n",
		"8. Move the issue to `review` when the pull request is ready for human review.\n",
		"## Complete the issue\n",
		"1. Run a review and resolve every High or higher severity finding.\n",
		"2. Leave an issue comment that summarizes the review result, including whether any High or higher severity findings were found and how they were handled.\n",
		"3. Create a pull request.\n",
		"4. After steps 1 through 3 are complete, run `tq issue update {{ issue.id }} --status review`\n",
	} {
		if !strings.Contains(template, want) {
			t.Fatalf("default workflow template missing %q", want)
		}
	}
}

func TestLoadReturnsDefaultsWhenConfigMissing(t *testing.T) {
	t.Setenv(EnvHome, t.TempDir())
	t.Setenv("USER", "agent")
	config, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if config.Author != "agent" {
		t.Fatalf("author=%q", config.Author)
	}
	if config.MaxConcurrentAgents != 10 {
		t.Fatalf("max=%d", config.MaxConcurrentAgents)
	}
}

func TestLoadReadsConfigYAML(t *testing.T) {
	home := t.TempDir()
	t.Setenv(EnvHome, home)
	if err := os.MkdirAll(ConfigDir(home), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ConfigPath(home), []byte("author: jiro\nmax_concurrent_agents: 3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	config, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if config.Author != "jiro" || config.MaxConcurrentAgents != 3 {
		t.Fatalf("config=%+v", config)
	}
}

func TestConfigResolveAppliesOverrides(t *testing.T) {
	author := "override"
	max := 2
	config := Config{Author: "file", MaxConcurrentAgents: 3}.Resolve(Overrides{
		Author:              &author,
		MaxConcurrentAgents: &max,
	})
	if config.Author != "override" || config.MaxConcurrentAgents != 2 {
		t.Fatalf("config=%+v", config)
	}
}
