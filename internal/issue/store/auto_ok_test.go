package store

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func TestAutoOKPersistsCountsAndStopsAtFive(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "issues.sqlite")
	s, err := OpenMigrated(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	project := createTestProject(t, s, "AUTOOK")
	issue, err := s.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Auto reply", Status: entity.StatusBlocked})
	if err != nil {
		t.Fatal(err)
	}
	if issue.AutoOK || issue.AutoOKCount != 0 || issue.AutoOKPending {
		t.Fatalf("defaults: %+v", issue)
	}
	enabled := true
	issue, err = s.UpdateIssue(ctx, issue.ID, entity.UpdateIssueInput{AutoOK: &enabled})
	if err != nil {
		t.Fatal(err)
	}
	if issue.Status != entity.StatusReady || !issue.AutoOKPending {
		t.Fatalf("enable blocked: %+v", issue)
	}
	for count := 1; count <= 5; count++ {
		generation := issue.AutoOKGeneration
		issue, err = s.CompleteAutoOK(ctx, issue.ID, entity.AutoOKResultInput{Generation: generation})
		if err != nil {
			t.Fatal(err)
		}
		if issue.AutoOKCount != count || issue.AutoOKPending || issue.AutoOK != (count < 5) {
			t.Fatalf("reply %d: %+v", count, issue)
		}
		// Duplicate acceptance must never consume another reply.
		issue, err = s.CompleteAutoOK(ctx, issue.ID, entity.AutoOKResultInput{Generation: generation})
		if err != nil || issue.AutoOKCount != count {
			t.Fatalf("duplicate: %+v %v", issue, err)
		}
		if err := s.Close(); err != nil {
			t.Fatal(err)
		}
		s, err = Open(ctx, path)
		if err != nil {
			t.Fatal(err)
		}
		persisted, err := s.Issue(ctx, issue.ID)
		if err != nil || persisted.AutoOKCount != count || persisted.AutoOK != issue.AutoOK {
			t.Fatalf("restart: %+v %v", persisted, err)
		}
		blocked := entity.StatusBlocked
		issue, err = s.UpdateIssue(ctx, issue.ID, entity.UpdateIssueInput{Status: &blocked})
		if err != nil {
			t.Fatal(err)
		}
	}
	if issue.Status != entity.StatusBlocked || issue.AutoOKStopReason != "limit" || issue.AutoOKPending {
		t.Fatalf("limit: %+v", issue)
	}
	var commentCount int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM comments WHERE issue_id = ? AND author = 'tasq-auto-ok'`, issue.ID).Scan(&commentCount); err != nil || commentCount != 5 {
		t.Fatalf("history count=%d err=%v", commentCount, err)
	}
	issue, err = s.UpdateIssue(ctx, issue.ID, entity.UpdateIssueInput{AutoOK: &enabled})
	if err != nil || issue.AutoOKCount != 0 || !issue.AutoOKPending || issue.AutoOKStopReason != "" {
		t.Fatalf("reenable: %+v %v", issue, err)
	}
}

func TestAutoOKFailureAndCancellation(t *testing.T) {
	ctx := context.Background()
	s, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issues.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	project := createTestProject(t, s, "AUTOFAIL")
	issue, err := s.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Failure"})
	if err != nil {
		t.Fatal(err)
	}
	enabled, disabled := true, false
	issue, err = s.UpdateIssue(ctx, issue.ID, entity.UpdateIssueInput{AutoOK: &enabled})
	if err != nil || issue.AutoOKPending || issue.Status != entity.StatusBacklog {
		t.Fatalf("enable backlog: %+v %v", issue, err)
	}
	blocked := entity.StatusBlocked
	issue, err = s.UpdateIssue(ctx, issue.ID, entity.UpdateIssueInput{Status: &blocked})
	if err != nil {
		t.Fatal(err)
	}
	oldGeneration := issue.AutoOKGeneration
	issue, err = s.UpdateIssue(ctx, issue.ID, entity.UpdateIssueInput{AutoOK: &disabled})
	if err != nil || issue.AutoOKPending || issue.Status != entity.StatusBlocked {
		t.Fatalf("cancel: %+v %v", issue, err)
	}
	issue, err = s.UpdateIssue(ctx, issue.ID, entity.UpdateIssueInput{AutoOK: &enabled})
	if err != nil {
		t.Fatal(err)
	}
	issue, err = s.CompleteAutoOK(ctx, issue.ID, entity.AutoOKResultInput{Generation: oldGeneration})
	if err != nil || issue.AutoOKCount != 0 || !issue.AutoOKPending {
		t.Fatalf("stale: %+v %v", issue, err)
	}
	issue, err = s.CompleteAutoOK(ctx, issue.ID, entity.AutoOKResultInput{Generation: issue.AutoOKGeneration, Error: "thread resume failed"})
	if err != nil || issue.AutoOK || issue.AutoOKPending || issue.AutoOKCount != 0 || issue.Status != entity.StatusBlocked || issue.AutoOKStopReason != "error" {
		t.Fatalf("failure: %+v %v", issue, err)
	}
	var body string
	if err := s.db.QueryRow(`SELECT body FROM comments WHERE issue_id = ?`, issue.ID).Scan(&body); err != nil || !strings.Contains(body, "thread resume failed") {
		t.Fatalf("failure history=%q err=%v", body, err)
	}
}
