package issue

import (
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/version-1/tasq/db/schema"
)

func TestEntitiesMatchIssueTrackerSchemaColumns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		tableName  string
		entityType any
	}{
		{name: "issue", tableName: "issues", entityType: IssueEntity{}},
		{name: "work item", tableName: "work_items", entityType: WorkItemEntity{}},
		{name: "orchestrator event", tableName: "orchestrator_events", entityType: OrchestratorEventEntity{}},
		{name: "run snapshot", tableName: "run_snapshots", entityType: RunSnapshotEntity{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			want := tableColumns(t, schema.IssueTracker, tt.tableName)
			got := entityColumns(t, tt.entityType)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("columns mismatch\n got: %v\nwant: %v", got, want)
			}
		})
	}
}

func entityColumns(t *testing.T, value any) []string {
	t.Helper()

	entityType := reflect.TypeOf(value)
	columns := make([]string, 0, entityType.NumField())
	for i := range entityType.NumField() {
		column := entityType.Field(i).Tag.Get("db")
		if column == "" {
			t.Fatalf("%s.%s has no db tag", entityType.Name(), entityType.Field(i).Name)
		}
		columns = append(columns, column)
	}
	return columns
}

func tableColumns(t *testing.T, schemaSQL string, tableName string) []string {
	t.Helper()

	pattern := regexp.MustCompile(`(?is)CREATE TABLE IF NOT EXISTS ` + regexp.QuoteMeta(tableName) + ` \((.*?)\);`)
	matches := pattern.FindStringSubmatch(schemaSQL)
	if len(matches) != 2 {
		t.Fatalf("table %q not found in schema", tableName)
	}

	lines := strings.Split(matches[1], "\n")
	columns := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(strings.TrimSuffix(line, ","))
		if line == "" || strings.HasPrefix(line, "FOREIGN KEY") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		columns = append(columns, fields[0])
	}
	return columns
}
