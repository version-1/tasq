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
