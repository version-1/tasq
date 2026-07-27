package tq

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func TestWriteIssueTextIncludesDependencies(t *testing.T) {
	tests := []struct {
		name string
		ids  []int64
		want string
	}{
		{name: "none", ids: nil, want: "Dependencies: none\n"},
		{name: "multiple", ids: []int64{2, 11, 42}, want: "Dependencies: 2,11,42\n"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := writeIssue(&buf, "text", entity.Issue{
				ID:            7,
				ProjectKey:    "CLI",
				Title:         "Show dependencies",
				Status:        entity.StatusDone,
				Priority:      entity.PriorityNormal,
				DependencyIDs: test.ids,
			})
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(buf.String(), test.want) {
				t.Fatalf("stdout does not contain %q: %s", test.want, buf.String())
			}
		})
	}
}

func TestWriteIssueJSONKeepsDependencyIDs(t *testing.T) {
	var buf bytes.Buffer
	err := writeIssue(&buf, "json", entity.Issue{
		ID:            7,
		ProjectKey:    "CLI",
		Title:         "Show dependencies",
		DependencyIDs: []int64{2, 11, 42},
	})
	if err != nil {
		t.Fatal(err)
	}

	var got entity.Issue
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	want := []int64{2, 11, 42}
	if len(got.DependencyIDs) != len(want) {
		t.Fatalf("dependency_ids=%v, want %v", got.DependencyIDs, want)
	}
	for i, id := range want {
		if got.DependencyIDs[i] != id {
			t.Fatalf("dependency_ids=%v, want %v", got.DependencyIDs, want)
		}
	}
}

func TestWriteIssuesTextUsesSemanticColors(t *testing.T) {
	var buf bytes.Buffer
	if err := writeIssues(&buf, "text", []entity.Issue{{ID: 7, ProjectKey: "tasq", Title: "Ready work", Status: entity.StatusReady}}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{ansiBold + "#7" + ansiReset, ansiCyan + "tasq" + ansiReset, ansiCyan + "ready" + ansiReset} {
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("output does not contain %q: %s", want, buf.String())
		}
	}
}

func TestWriteProjectCheckItemsTextUsesPassFailColors(t *testing.T) {
	var buf bytes.Buffer
	if err := writeProjectCheckItems(&buf, "text", []projectCheckItem{{Name: "valid", Passed: true}, {Name: "invalid", Passed: false}}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{ansiGreen + "PASS" + ansiReset, ansiRed + "FAIL" + ansiReset} {
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("output does not contain %q: %s", want, buf.String())
		}
	}
}

func TestWriteServiceStatusesTextUsesStateColors(t *testing.T) {
	var buf bytes.Buffer
	if err := writeServiceStatuses(&buf, []serviceStatus{
		{Name: "issue-tracker", State: "running", PID: 42, Port: 37651, Uptime: "1m"},
		{Name: "web", State: "stopped"},
	}); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		ansiCyan + "issue-tracker" + ansiReset,
		ansiGreen + "running" + ansiReset,
		ansiCyan + "web" + ansiReset,
		ansiFaint + "stopped" + ansiReset,
	} {
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("output does not contain %q: %s", want, buf.String())
		}
	}
}

func TestWriteMigrateResultsDistinguishesStatusAndNoChanges(t *testing.T) {
	var status bytes.Buffer
	if err := writeMigrateResults(&status, "text", "Migration status", []migrateResult{{Database: "issue-tracker", Statuses: []migrateStatus{}}}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(status.String(), "no changes") {
		t.Fatalf("status output must not render no changes: %s", status.String())
	}

	var noChanges bytes.Buffer
	if err := writeMigrateResults(&noChanges, "text", "Applied migrations", []migrateResult{{Database: "issue-tracker"}}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(noChanges.String(), ansiFaint+"no changes"+ansiReset) {
		t.Fatalf("output does not contain no changes: %s", noChanges.String())
	}
}
