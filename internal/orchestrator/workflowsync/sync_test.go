package workflowsync

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func TestSyncProjectWorkflowsUpdatesMissingAndChangedWorkflows(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	changedDir := t.TempDir()
	missingStoredDir := t.TempDir()
	noWorkflowDir := t.TempDir()
	changedRaw := []byte("---\ntracker:\n  kind: tasq\n---\nChanged prompt\n")
	missingStoredRaw := []byte("Prompt without frontmatter\n")
	writeWorkflow(t, changedDir, changedRaw)
	writeWorkflow(t, missingStoredDir, missingStoredRaw)
	client := &fakeTracker{
		projects: []entity.Project{
			{ID: 1, Key: "CHANGED", Location: changedDir},
			{ID: 2, Key: "MISSINGDB", Location: missingStoredDir},
			{ID: 3, Key: "MISSINGFILE", Location: noWorkflowDir},
		},
		workflows: map[int64]entity.ProjectWorkflow{
			1: {Checksum: strings.Repeat("0", 64)},
		},
	}

	result, err := SyncProjectWorkflows(ctx, client)
	if err != nil {
		t.Fatalf("sync project workflows: %v", err)
	}

	if result.Projects != 3 || result.Updated != 2 || result.Skipped != 0 || result.Missing != 1 {
		t.Fatalf("result = %+v", result)
	}
	if got := client.upserts[1]; got.Checksum != checksum(changedRaw) || got.FrontmatterJSON != `{"tracker":{"kind":"tasq"}}` || got.Body != "Changed prompt" {
		t.Fatalf("changed upsert = %+v", got)
	}
	if got := client.upserts[2]; got.Checksum != checksum(missingStoredRaw) || got.FrontmatterJSON != `{}` || got.Body != "Prompt without frontmatter" {
		t.Fatalf("missing stored upsert = %+v", got)
	}
	if _, ok := client.upserts[3]; ok {
		t.Fatalf("missing workflow was upserted: %+v", client.upserts[3])
	}
}

func TestSyncProjectWorkflowsSkipsMatchingChecksum(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	raw := []byte("---\ntasq:\n  task_work_prompt: false\n---\nPrompt\n")
	writeWorkflow(t, dir, raw)
	client := &fakeTracker{
		projects:  []entity.Project{{ID: 1, Key: "MATCH", Location: dir}},
		workflows: map[int64]entity.ProjectWorkflow{1: {Checksum: checksum(raw)}},
	}

	result, err := SyncProjectWorkflows(context.Background(), client)
	if err != nil {
		t.Fatalf("sync project workflows: %v", err)
	}

	if result.Projects != 1 || result.Updated != 0 || result.Skipped != 1 || result.Missing != 0 {
		t.Fatalf("result = %+v", result)
	}
	if len(client.upserts) != 0 {
		t.Fatalf("upserts = %+v", client.upserts)
	}
}

func TestReadProjectWorkflowRejectsUnterminatedFrontmatter(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeWorkflow(t, dir, []byte("---\ntracker:\n  kind: tasq\nPrompt\n"))

	_, _, err := ReadProjectWorkflow(dir)
	if err == nil || !strings.Contains(err.Error(), "unterminated front matter") {
		t.Fatalf("error = %v", err)
	}
}

type fakeTracker struct {
	projects  []entity.Project
	workflows map[int64]entity.ProjectWorkflow
	upserts   map[int64]entity.UpsertProjectWorkflowInput
}

func (f *fakeTracker) Projects(context.Context) ([]entity.Project, error) {
	return f.projects, nil
}

func (f *fakeTracker) Workflow(_ context.Context, projectID int64) (entity.ProjectWorkflow, bool, error) {
	workflow, ok := f.workflows[projectID]
	return workflow, ok, nil
}

func (f *fakeTracker) UpsertWorkflow(_ context.Context, projectID int64, input entity.UpsertProjectWorkflowInput) (entity.ProjectWorkflow, error) {
	if f.upserts == nil {
		f.upserts = map[int64]entity.UpsertProjectWorkflowInput{}
	}
	copied := entity.UpsertProjectWorkflowInput{
		ProjectID:       input.ProjectID,
		FrontmatterJSON: input.FrontmatterJSON,
		Body:            input.Body,
		Checksum:        input.Checksum,
	}
	f.upserts[projectID] = copied
	return entity.ProjectWorkflow{
		ProjectID: projectID,
		Body:      copied.Body,
		Checksum:  copied.Checksum,
	}, nil
}

func writeWorkflow(t *testing.T, dir string, raw []byte) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, workflowFilename), raw, 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}
}

func checksum(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
