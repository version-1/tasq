package store

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

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
	first, err := s.CreateDocumentArtifact(ctx, project.ID, entity.CreateDocumentArtifactInput{Type: entity.DocumentArtifactTypeWorkReport, Title: "First", Body: "# one"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.CreateDocumentArtifact(ctx, project.ID, entity.CreateDocumentArtifactInput{Type: entity.DocumentArtifactTypeWorkReport, Title: "Second", Body: "# two"})
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == second.ID {
		t.Fatal("document ids must differ")
	}
	summaries, err := s.DocumentArtifacts(ctx, project.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 2 || summaries[0].Body != "" {
		t.Fatalf("summaries = %#v", summaries)
	}
	updated, err := s.UpdateDocumentArtifact(ctx, project.ID, first.ID, entity.UpdateDocumentArtifactInput{Body: stringPointer("# revised")})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "First" || updated.Body != "# revised" {
		t.Fatalf("body-only update = %#v", updated)
	}
	updated, err = s.UpdateDocumentArtifact(ctx, project.ID, first.ID, entity.UpdateDocumentArtifactInput{Title: stringPointer("Renamed")})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Title != "Renamed" || updated.Body != "# revised" {
		t.Fatalf("title-only update = %#v", updated)
	}
	if err := s.LinkDocumentArtifactToIssue(ctx, project.ID, first.ID, issue.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM issues WHERE id = ?`, issue.ID); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM document_artifacts`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("document artifacts after issue deletion = %d", count)
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
	item, err := s.CreateDocumentArtifact(ctx, project.ID, entity.CreateDocumentArtifactInput{Type: entity.DocumentArtifactTypeReferenceDocument, Title: "Reference", Body: "content"})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.LinkDocumentArtifactToIssue(ctx, project.ID, item.ID, first.ID); err != nil {
		t.Fatal(err)
	}
	if items, err := s.DocumentArtifactsForIssue(ctx, second.ID); err != nil || len(items) != 0 {
		t.Fatalf("unlinked issue artifacts=%+v err=%v", items, err)
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
	document, err := oldStore.CreateDocumentArtifact(ctx, project.ID, entity.CreateDocumentArtifactInput{Type: entity.DocumentArtifactTypeReferenceDocument, Title: "Reference", Body: "# preserved issue"})
	if err != nil {
		t.Fatal(err)
	}
	if document.ProjectID != project.ID {
		t.Fatalf("document project id=%d want %d", document.ProjectID, project.ID)
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
		if entry.IsDir() || entry.Name() == "20261104000000_create_issue_document_artifacts.up.sql" || entry.Name() == "20261104000000_create_issue_document_artifacts.down.sql" || entry.Name() == "20261115000000_project_document_artifacts.up.sql" || entry.Name() == "20261115000000_project_document_artifacts.down.sql" {
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

func TestProjectDocumentArtifactMigrationPreservesLegacyDocument(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "migration.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys = ON`); err != nil {
		t.Fatal(err)
	}
	legacy := migration.NewManager(db, issueTrackerMigrationsBeforeProjectDocumentArtifacts(t), "issue-tracker")
	if _, err := legacy.Apply(ctx); err != nil {
		t.Fatal(err)
	}
	s := &Store{db: db}
	project, err := s.CreateProject(ctx, entity.CreateProjectInput{Key: "owner", Name: "Owner", Location: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	issue, err := s.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Legacy"})
	if err != nil {
		t.Fatal(err)
	}
	const id, body, created, updated = "preserved-id", "# preserved", "2026-11-01T01:02:03Z", "2026-11-02T01:02:03Z"
	if _, err := db.ExecContext(ctx, `INSERT INTO issue_document_artifacts (id,issue_id,type,title,body,created_at,updated_at) VALUES (?,?,?,?,?,?,?)`, id, issue.ID, "work_report", "Legacy", body, created, updated); err != nil {
		t.Fatal(err)
	}
	manager := migration.NewManager(db, migrations.Files, "issue-tracker")
	if _, err := manager.Apply(ctx); err != nil {
		t.Fatal(err)
	}
	var gotID, gotBody, gotCreated, gotUpdated string
	var owner, target int64
	if err := db.QueryRowContext(ctx, `SELECT a.id,a.body,a.created_at,a.updated_at,a.project_id,l.target_id FROM document_artifacts a JOIN document_artifact_links l ON l.document_artifact_id=a.id WHERE a.id=?`, id).Scan(&gotID, &gotBody, &gotCreated, &gotUpdated, &owner, &target); err != nil {
		t.Fatal(err)
	}
	if gotID != id || gotBody != body || gotCreated != created || gotUpdated != updated || owner != project.ID || target != issue.ID {
		t.Fatalf("migration lost document: %q %q %q %q owner=%d target=%d", gotID, gotBody, gotCreated, gotUpdated, owner, target)
	}
	if _, err := manager.Down(ctx); err != nil {
		t.Fatal(err)
	}
	var rollbackBody string
	if err := db.QueryRowContext(ctx, `SELECT body FROM issue_document_artifacts WHERE id=? AND issue_id=?`, id, issue.ID).Scan(&rollbackBody); err != nil || rollbackBody != body {
		t.Fatalf("rollback body=%q err=%v", rollbackBody, err)
	}
}

func TestProjectDocumentArtifactRollbackRejectsLossyLinks(t *testing.T) {
	ctx := context.Background()
	s, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "rollback.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	owner, _ := s.CreateProject(ctx, entity.CreateProjectInput{Key: "owner", Name: "Owner", Location: t.TempDir()})
	other, _ := s.CreateProject(ctx, entity.CreateProjectInput{Key: "other", Name: "Other", Location: t.TempDir()})
	issue, _ := s.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: other.ID, Title: "External"})
	item, _ := s.CreateDocumentArtifact(ctx, owner.ID, entity.CreateDocumentArtifactInput{Type: entity.DocumentArtifactTypeWorkReport, Title: "Unlinked", Body: "body"})
	manager := migration.NewManager(s.db, migrations.Files, "issue-tracker")
	if _, err := manager.Down(ctx); err == nil {
		t.Fatal("lossy rollback succeeded")
	}
	if _, err := s.DocumentArtifact(ctx, owner.ID, item.ID); err != nil {
		t.Fatalf("artifact lost after rejected rollback: %v", err)
	}
	if err := s.LinkDocumentArtifactToIssue(ctx, owner.ID, item.ID, issue.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Down(ctx); err == nil {
		t.Fatal("cross-project rollback succeeded")
	}
}

func TestProjectDocumentArtifactRollbackRejectsMultipleSameProjectLinks(t *testing.T) {
	ctx := context.Background()
	s, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "multiple-links.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	project, _ := s.CreateProject(ctx, entity.CreateProjectInput{Key: "owner", Name: "Owner", Location: t.TempDir()})
	first, _ := s.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "first"})
	second, _ := s.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "second"})
	item, _ := s.CreateDocumentArtifact(ctx, project.ID, entity.CreateDocumentArtifactInput{Type: entity.DocumentArtifactTypeWorkReport, Title: "Shared", Body: "body"})
	if err := s.LinkDocumentArtifactToIssue(ctx, project.ID, item.ID, first.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.LinkDocumentArtifactToIssue(ctx, project.ID, item.ID, second.ID); err != nil {
		t.Fatal(err)
	}
	manager := migration.NewManager(s.db, migrations.Files, "issue-tracker")
	if _, err := manager.Down(ctx); err == nil {
		t.Fatal("multiple-link rollback succeeded")
	}
	if _, err := s.DocumentArtifact(ctx, project.ID, item.ID); err != nil {
		t.Fatalf("artifact lost after rejected rollback: %v", err)
	}
}

func issueTrackerMigrationsBeforeProjectDocumentArtifacts(t *testing.T) fs.FS {
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
		if entry.IsDir() || entry.Name() == "20261115000000_project_document_artifacts.up.sql" || entry.Name() == "20261115000000_project_document_artifacts.down.sql" {
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

func TestDeleteProjectWithCleanupRejectsExternalLinkBeforeCleanup(t *testing.T) {
	ctx := context.Background()
	s, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "delete.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	owner, _ := s.CreateProject(ctx, entity.CreateProjectInput{Key: "owner", Name: "Owner", Location: t.TempDir()})
	other, _ := s.CreateProject(ctx, entity.CreateProjectInput{Key: "other", Name: "Other", Location: t.TempDir()})
	// Ensure issue ID differs from the owner project ID.
	_, _ = s.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: owner.ID, Title: "owner first"})
	external, _ := s.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: other.ID, Title: "external"})
	item, _ := s.CreateDocumentArtifact(ctx, owner.ID, entity.CreateDocumentArtifactInput{Type: entity.DocumentArtifactTypeWorkReport, Title: "Shared", Body: "body"})
	if err := s.LinkDocumentArtifactToIssue(ctx, owner.ID, item.ID, external.ID); err != nil {
		t.Fatal(err)
	}
	called := false
	err = s.DeleteProjectWithCleanup(ctx, owner.ID, func(context.Context, []int64) error { called = true; return nil })
	if !errors.Is(err, ErrProjectHasExternalDocumentArtifactLinks) || called {
		t.Fatalf("delete err=%v cleanup=%v", err, called)
	}
	if _, err := s.Project(ctx, owner.ID); err != nil {
		t.Fatalf("owner deleted: %v", err)
	}
}

func TestDeleteProjectWithCleanupBlocksConcurrentLinkUntilCommit(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "race.sqlite")
	owner, err := OpenMigrated(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer owner.Close()
	linker, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer linker.Close()
	if _, err := linker.db.ExecContext(ctx, `PRAGMA busy_timeout = 1000`); err != nil {
		t.Fatal(err)
	}
	project, _ := owner.CreateProject(ctx, entity.CreateProjectInput{Key: "owner", Name: "Owner", Location: t.TempDir()})
	// The target ID is deliberately different from its owner project ID.
	_, _ = owner.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "first"})
	target, _ := owner.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "target"})
	item, _ := owner.CreateDocumentArtifact(ctx, project.ID, entity.CreateDocumentArtifactInput{Type: entity.DocumentArtifactTypeWorkReport, Title: "Report", Body: "body"})
	cleanupStarted := make(chan struct{})
	allowCleanup := make(chan struct{})
	var releaseCleanup sync.Once
	release := func() { releaseCleanup.Do(func() { close(allowCleanup) }) }
	defer release()
	deleteDone := make(chan error, 1)
	go func() {
		deleteDone <- owner.DeleteProjectWithCleanup(ctx, project.ID, func(context.Context, []int64) error { close(cleanupStarted); <-allowCleanup; return nil })
	}()
	select {
	case <-cleanupStarted:
	case <-time.After(time.Second):
		t.Fatal("cleanup did not start")
	}
	linkDone := make(chan error, 1)
	go func() { linkDone <- linker.LinkDocumentArtifactToIssue(ctx, project.ID, item.ID, target.ID) }()
	select {
	case err := <-linkDone:
		t.Fatalf("link completed while deletion lock held: %v", err)
	case <-time.After(100 * time.Millisecond):
	}
	release()
	select {
	case err := <-deleteDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("delete did not commit")
	}
	select {
	case err := <-linkDone:
		if err == nil {
			t.Fatal("link succeeded after project deletion")
		}
	case <-time.After(time.Second):
		t.Fatal("blocked link did not return")
	}
	var links int
	if err := linker.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM document_artifact_links WHERE document_artifact_id = ?`, item.ID).Scan(&links); err != nil {
		t.Fatal(err)
	}
	if links != 0 {
		t.Fatalf("orphan links=%d", links)
	}
}
