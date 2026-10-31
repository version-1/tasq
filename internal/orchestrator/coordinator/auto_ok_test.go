package coordinator

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	issueapi "github.com/version-1/tasq/internal/issue/api"
	"github.com/version-1/tasq/internal/issue/domain/entity"
	issuestore "github.com/version-1/tasq/internal/issue/store"
	"github.com/version-1/tasq/internal/orchestrator/run"
	"github.com/version-1/tasq/internal/orchestrator/runner"
	"github.com/version-1/tasq/internal/orchestrator/tracker"
)

func TestDispatcherAutoOKThroughTrackerAPI(t *testing.T) {
	for _, test := range []struct {
		name string
		fail bool
	}{{"accepted", false}, {"failed before delivery", true}} {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			issues, err := issuestore.OpenMigrated(ctx, filepath.Join(t.TempDir(), "tracker.sqlite"))
			if err != nil {
				t.Fatal(err)
			}
			defer issues.Close()
			project, err := issues.CreateProject(ctx, entity.CreateProjectInput{Key: "AUTO", Name: "Auto", Location: t.TempDir()})
			if err != nil {
				t.Fatal(err)
			}
			issue, err := issues.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Auto", Status: entity.StatusBlocked})
			if err != nil {
				t.Fatal(err)
			}
			enabled := true
			issue, err = issues.UpdateIssue(ctx, issue.ID, entity.UpdateIssueInput{AutoOK: &enabled})
			if err != nil {
				t.Fatal(err)
			}
			server := httptest.NewServer(issueapi.NewServer(issues).Handler())
			defer server.Close()
			store := openTestStore(t)
			previous := createQueuedRun(t, store, issue.ID)
			if _, err := store.UpdateRunThreadID(ctx, previous.RunID, "previous-thread"); err != nil {
				t.Fatal(err)
			}
			if _, err := store.UpdateRunStatus(ctx, previous.RunID, run.StatusFailed, "blocked"); err != nil {
				t.Fatal(err)
			}
			queued := createQueuedRun(t, store, issue.ID)
			var testRunner runner.Runner = runner.SimulatedRunner{Duration: time.Millisecond}
			if test.fail {
				testRunner = &recordingRunner{result: runner.Result{Status: run.StatusFailed, Error: "could not resume thread"}}
			}
			d, err := NewDispatcher(DispatcherConfig{Tracker: tracker.NewClient(server.URL), Store: store, Runner: testRunner, WorkflowResolver: staticWorkflowResolver(testWorkflowDefinition()), MaxConcurrentRuns: 1})
			if err != nil {
				t.Fatal(err)
			}
			if err := d.Dispatch(ctx, []run.Run{queued}); err != nil {
				t.Fatal(err)
			}
			deadline := time.Now().Add(3 * time.Second)
			for time.Now().Before(deadline) {
				stored, err := store.RunByRunID(ctx, queued.RunID)
				if err != nil {
					t.Fatal(err)
				}
				if stored.Status == run.StatusSucceeded || stored.Status == run.StatusFailed {
					break
				}
				time.Sleep(time.Millisecond)
			}
			shutdownDispatcher(t, d)
			updated, err := issues.Issue(ctx, issue.ID)
			if err != nil {
				t.Fatal(err)
			}
			if test.fail {
				if updated.AutoOK || updated.AutoOKPending || updated.AutoOKCount != 0 || updated.Status != entity.StatusBlocked || updated.AutoOKStopReason != "error" {
					t.Fatalf("failure=%+v", updated)
				}
			} else if !updated.AutoOK || updated.AutoOKPending || updated.AutoOKCount != 1 || updated.Status != entity.StatusInProgress {
				t.Fatalf("accepted=%+v", updated)
			}
		})
	}
}
