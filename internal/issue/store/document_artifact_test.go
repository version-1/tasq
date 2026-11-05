package store

import (
	"context"
	"database/sql"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/version-1/tasq/db/migrations"
	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/migration"
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
	if updated.Title != "First" || updated.Body != "# revised" {
		t.Fatalf("body-only update = %#v", updated)
	}
	updated, err = s.UpdateDocumentArtifact(ctx, issue.ID, first.ID, entity.UpdateDocumentArtifactInput{Title: stringPointer("Renamed")})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "Renamed" || updated.Body != "# revised" {
		t.Fatalf("title-only update = %#v", updated)
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
	if _, err := s.UpdateDocumentArtifact(ctx, second.ID, item.ID, entity.UpdateDocumentArtifactInput{Title: stringPointer("Wrong issue")}); err != sql.ErrNoRows {
		t.Fatalf("foreign issue update error = %v", err)
	}
}

func TestDocumentArtifactMigrationPreservesExistingIssuesAndPullRequestArtifacts(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "issues.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys = ON`); err != nil {
		t.Fatal(err)
	}
	if _, err := migration.NewManager(db, issueTrackerMigrationsBeforeDocumentArtifacts(t), "issue-tracker").Apply(ctx); err != nil {
		t.Fatalf("apply old migrations: %v", err)
	}
	oldStore := &Store{db: db}
	project, err := oldStore.CreateProject(ctx, entity.CreateProjectInput{Key: "legacy", Name: "Legacy", Location: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	issue, err := oldStore.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Existing issue"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := oldStore.UpsertArtifact(ctx, issue.ID, entity.ArtifactTypePullRequest, entity.UpsertArtifactInput{DataValue: "https://example.com/pr/legacy"}); err != nil {
		t.Fatal(err)
	}

	if _, err := migration.NewManager(db, migrations.Files, "issue-tracker").Apply(ctx); err != nil {
		t.Fatalf("apply document artifact migration: %v", err)
	}
	read, err := oldStore.Issue(ctx, issue.ID)
	if err != nil {
		t.Fatal(err)
	}
	if read.ID != issue.ID || len(read.Artifacts) != 1 || read.Artifacts[0].DataValue != "https://example.com/pr/legacy" {
		t.Fatalf("existing data after migration = %+v", read)
	}
	document, err := oldStore.CreateDocumentArtifact(ctx, issue.ID, entity.CreateDocumentArtifactInput{Type: entity.DocumentArtifactTypeReferenceDocument, Title: "Reference", Body: "# preserved issue"})
	if err != nil {
		t.Fatal(err)
	}
	if document.IssueID != issue.ID {
		t.Fatalf("document issue id=%d want %d", document.IssueID, issue.ID)
	}
}

func issueTrackerMigrationsBeforeDocumentArtifacts(t *testing.T) fs.FS {
	t.Helper()
	entries, err := fs.ReadDir(migrations.Files, "issue-tracker")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	directory := filepath.Join(root, "issue-tracker")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "20261104000000_create_issue_document_artifacts.up.sql" || entry.Name() == "20261104000000_create_issue_document_artifacts.down.sql" {
			continue
		}
		path := "issue-tracker/" + entry.Name()
		contents, err := fs.ReadFile(migrations.Files, path)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, path), contents, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return os.DirFS(root)
}

func stringPointer(value string) *string { return &value }
