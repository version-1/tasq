package tq

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

type outputStory struct {
	name string
	run  func(io.Writer) error
}

// RunStory renders a deterministic tq output scenario for manual visual checks.
func RunStory(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		writeStoryUsage(stderr)
		return 2
	}
	for _, story := range outputStories() {
		if story.name != args[0] {
			continue
		}
		if err := story.run(stdout); err != nil {
			return writeCLIErrorForFormat(stderr, "text", err.Error(), 1)
		}
		return 0
	}
	writeStoryUsage(stderr)
	return 2
}

func outputStories() []outputStory {
	issue := storyIssue()
	return []outputStory{
		{name: "tq_issue_list", run: func(w io.Writer) error {
			return writeIssues(w, "text", []entity.Issue{issue, {ID: 43, ProjectKey: "tasq", Title: "Review CLI output", Status: entity.StatusReview, Priority: entity.PriorityHigh}})
		}},
		{name: "tq_issue_detail", run: func(w io.Writer) error { return writeIssue(w, "text", issue) }},
		{name: "tq_issue_action", run: func(w io.Writer) error { return writeIssueAction(w, "text", issue, "Issue #42 marked as done") }},
		{name: "tq_project_list", run: func(w io.Writer) error {
			return writeProjects(w, "text", []entity.Project{{ID: 7, Key: "tasq", Name: "Tasq", Location: "/workspace/tasq", UpdatedAt: storyTime()}, {ID: 8, Key: "docs", Name: "Documentation", Location: "/workspace/docs", UpdatedAt: storyTime()}})
		}},
		{name: "tq_empty", run: func(w io.Writer) error { return writeProjects(w, "text", nil) }},
		{name: "tq_project_check", run: func(w io.Writer) error {
			return writeProjectCheckItems(w, "text", []projectCheckItem{{Name: "WORKFLOW.md", Passed: true, Reason: "valid"}, {Name: "workflow.yaml", Passed: false, Reason: "missing required owner"}})
		}},
		{name: "tq_service_status", run: func(w io.Writer) error {
			return writeServiceStatuses(w, []serviceStatus{{Name: "issue-tracker", State: "running", PID: 4242, Port: 37651, Uptime: "2m15s"}, {Name: "orchestrator", State: "running", PID: 4243, Port: 37652, Uptime: "2m14s"}, {Name: "web", State: "stopped"}})
		}},
		{name: "tq_migration_status", run: func(w io.Writer) error {
			return writeMigrateResults(w, "text", "Migration status", []migrateResult{{Database: "issue-tracker", Path: "/workspace/.tasq/issues.sqlite", Statuses: []migrateStatus{{Version: "20260615000000", Name: "init", Applied: true}, {Version: "20260701000000", Name: "add_runs", Applied: false}}}, {Database: "orchestrator", Path: "/workspace/.tasq/orchestrator.sqlite"}})
		}},
		{name: "tq_warning", run: writeStoryWarning},
		{name: "tq_service_start_fail", run: func(w io.Writer) error {
			writeCLIErrorForFormat(w, "text", "issue-tracker health check failed: connection refused", 1)
			return nil
		}},
		{name: "tq_json_success", run: func(w io.Writer) error { return writeIssue(w, "json", issue) }},
		{name: "tq_json_error", run: func(w io.Writer) error { writeCLIErrorForFormat(w, "json", "issue not found", 1); return nil }},
		{name: "all", run: writeAllStories},
	}
}

func writeAllStories(w io.Writer) error {
	for _, story := range outputStories() {
		if story.name == "all" {
			continue
		}
		if _, err := fmt.Fprintf(w, "%s== %s ==%s\n", ansiBold, story.name, ansiReset); err != nil {
			return err
		}
		if err := story.run(w); err != nil {
			return err
		}
	}
	return nil
}

func writeStoryWarning(w io.Writer) error {
	_, err := fmt.Fprintf(w, "%sWARNING:%s This operation cannot be undone.\nProject to remove: %stasq%s (Tasq)\n", ansiRed, ansiReset, ansiCyan, ansiReset)
	return err
}

func writeStoryUsage(w io.Writer) {
	names := make([]string, 0, len(outputStories()))
	for _, story := range outputStories() {
		names = append(names, story.name)
	}
	sort.Strings(names)
	fmt.Fprintf(w, "Usage: tqstory <scenario>\n\nScenarios:\n  %s\n", strings.Join(names, "\n  "))
}

func storyIssue() entity.Issue {
	return entity.Issue{ID: 42, ProjectID: 7, ProjectKey: "tasq", Title: "Make tq output easier to scan", Description: "Color semantic states while preserving text output.", Status: entity.StatusDone, Priority: entity.PriorityHigh, Assignee: "codex", DependencyIDs: []int64{10, 11}, CreatedAt: storyTime(), UpdatedAt: storyTime()}
}

func storyTime() time.Time {
	return time.Date(2026, time.July, 26, 9, 30, 0, 0, time.UTC)
}
