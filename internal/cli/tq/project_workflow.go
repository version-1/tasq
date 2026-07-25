package tq

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type workflowAddContent struct {
	Frontmatter map[string]any
	Body        string
	Checksum    string
}

func checkWorkflowExists(err error) projectCheckItem {
	if err == nil {
		return projectCheckItem{Name: "workflow.exists", Passed: true, Reason: "WORKFLOW.md exists"}
	}
	return projectCheckItem{Name: "workflow.exists", Passed: false, Reason: "WORKFLOW.md is missing"}
}

func checkWorkflowFrontMatter(content []byte) projectCheckItem {
	frontMatter, ok := workflowFrontMatter(string(content))
	if !ok {
		return projectCheckItem{Name: "workflow.front_matter", Passed: false, Reason: "WORKFLOW.md front matter is missing"}
	}
	var raw map[string]any
	if err := yaml.Unmarshal([]byte(frontMatter), &raw); err != nil {
		return projectCheckItem{Name: "workflow.front_matter", Passed: false, Reason: err.Error()}
	}
	missing := missingWorkflowFields(raw)
	if len(missing) > 0 {
		return projectCheckItem{Name: "workflow.front_matter", Passed: false, Reason: "missing fields: " + strings.Join(missing, ", ")}
	}
	return projectCheckItem{Name: "workflow.front_matter", Passed: true, Reason: "required front matter fields are present"}
}

func parseWorkflowAddContent(content string) (workflowAddContent, error) {
	frontMatter, body, ok := splitWorkflowContent(content)
	if !ok {
		return workflowAddContent{}, usageError("workflow front matter is required")
	}
	var parsed map[string]any
	if err := yaml.Unmarshal([]byte(frontMatter), &parsed); err != nil {
		return workflowAddContent{}, fmt.Errorf("parse workflow front matter: %w", err)
	}
	if parsed == nil {
		return workflowAddContent{}, usageError("workflow front matter must be a YAML object")
	}
	sum := sha256.Sum256([]byte(content))
	return workflowAddContent{
		Frontmatter: parsed,
		Body:        body,
		Checksum:    fmt.Sprintf("%x", sum),
	}, nil
}

func workflowFrontMatter(content string) (string, bool) {
	frontMatter, _, ok := splitWorkflowContent(content)
	return frontMatter, ok
}

func splitWorkflowContent(content string) (string, string, bool) {
	if !strings.HasPrefix(content, "---\n") {
		return "", "", false
	}
	rest := strings.TrimPrefix(content, "---\n")
	frontMatter, body, ok := strings.Cut(rest, "\n---\n")
	return frontMatter, body, ok
}

func missingWorkflowFields(raw map[string]any) []string {
	required := []string{
		"polling.interval_ms",
		"workspace.root",
		"agent.max_concurrent_agents",
		"agent.max_turns",
		"agent.continuation_turns_enabled",
		"agent.max_retry_attempts",
		"agent.max_retry_backoff_ms",
		"codex.command",
		"codex.read_timeout_ms",
		"codex.turn_timeout_ms",
		"codex.stall_timeout_ms",
	}
	missing := []string{}
	for _, field := range required {
		if !hasNestedField(raw, strings.Split(field, ".")) {
			missing = append(missing, field)
		}
	}
	return missing
}

func hasNestedField(raw map[string]any, path []string) bool {
	current := any(raw)
	for _, key := range path {
		asMap, ok := current.(map[string]any)
		if !ok {
			return false
		}
		next, ok := asMap[key]
		if !ok {
			return false
		}
		current = next
	}
	return true
}
