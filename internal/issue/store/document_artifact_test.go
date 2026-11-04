package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func TestDocumentArtifactsRetainMultipleDocumentsAndCascade(t *testing.T) {
	ctx := context.Background()
	s, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issues.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	project, err := s.CreateProject(ctx, entity.CreateProjectInput{Key: "docs", Name: "Docs", Location: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	issue, err := s.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Documents"})
	if err != nil {
		t.Fatal(err)
	}
	first, err := s.CreateDocumentArtifact(ctx, issue.ID, entity.CreateDocumentArtifactInput{Type: entity.DocumentArtifactTypeWorkReport, Title: "First", Body: "# one"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.CreateDocumentArtifact(ctx, issue.ID, entity.CreateDocumentArtifactInput{Type: entity.DocumentArtifactTypeWorkReport, Title: "Second", Body: "# two"})
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == second.ID {
		t.Fatal("document ids must differ")
	}
	summaries, err := s.DocumentArtifacts(ctx, issue.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 2 || summaries[0].Body != "" {
		t.Fatalf("summaries = %#v", summaries)
	}
	updated, err := s.UpdateDocumentArtifact(ctx, issue.ID, first.ID, entity.UpdateDocumentArtifactInput{Body: stringPointer("# revised")})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Body != "# revised" {
		t.Fatalf("body = %q", updated.Body)
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM issues WHERE id = ?`, issue.ID); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM issue_document_artifacts`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("document artifacts after cascade = %d", count)
	}
}

func TestDocumentArtifactRequiresMatchingIssue(t *testing.T) {
	ctx := context.Background()
	s, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issues.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	project, _ := s.CreateProject(ctx, entity.CreateProjectInput{Key: "docs", Name: "Docs", Location: t.TempDir()})
	first, _ := s.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "First"})
	second, _ := s.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Second"})
	item, err := s.CreateDocumentArtifact(ctx, first.ID, entity.CreateDocumentArtifactInput{Type: entity.DocumentArtifactTypeReferenceDocument, Title: "Reference", Body: "content"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.DocumentArtifact(ctx, second.ID, item.ID); err != sql.ErrNoRows {
		t.Fatalf("foreign issue get error = %v", err)
	}
}

func stringPointer(value string) *string { return &value }
