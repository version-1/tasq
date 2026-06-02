package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

func TestOpenAppliesIssueTrackerSchema(t *testing.T) {
	t.Parallel()

	store, err := Open(context.Background(), filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	for _, name := range []string{
		"issues",
		"issues_project_id_idx",
		"issues_status_idx",
		"comments",
		"comments_issue_id_idx",
		"projects",
		"projects_key_idx",
		"workspaces",
		"workspaces_project_id_idx",
		"workspaces_status_idx",
		"attachments",
		"attachments_entity_idx",
	} {
		if !schemaObjectExists(t, store, name) {
			t.Fatalf("schema object %q does not exist", name)
		}
	}
}

func TestOpenDeletesLegacyIssuesWhenProjectIDIsMissing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "issue-tracker.sqlite")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE issues (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			priority TEXT NOT NULL,
			assignee TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		);
		CREATE TABLE comments (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			issue_id INTEGER NOT NULL,
			author TEXT NOT NULL,
			type TEXT NOT NULL,
			body TEXT NOT NULL,
			created_at TEXT NOT NULL
		);
		CREATE TABLE attachments (
			id TEXT PRIMARY KEY,
			entity_type TEXT NOT NULL,
			entity_id TEXT NOT NULL,
			filename TEXT NOT NULL,
			path TEXT NOT NULL,
			content_type TEXT NOT NULL,
			size INTEGER NOT NULL,
			created_at TEXT NOT NULL
		);
		INSERT INTO issues (title, status, priority, created_at, updated_at) VALUES ('Legacy', 'backlog', 'normal', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z');
		INSERT INTO comments (issue_id, author, type, body, created_at) VALUES (1, 'codex', 'general', 'legacy', '2026-01-01T00:00:00Z');
		INSERT INTO attachments (id, entity_type, entity_id, filename, path, content_type, size, created_at) VALUES
			('issue_att', 'issue', '1', 'issue.png', 'issue.png', 'image/png', 1, '2026-01-01T00:00:00Z'),
			('comment_att', 'comment', '1', 'comment.png', 'comment.png', 'image/png', 1, '2026-01-01T00:00:00Z'),
			('other_att', 'run', '1', 'run.png', 'run.png', 'image/png', 1, '2026-01-01T00:00:00Z');
	`); err != nil {
		t.Fatalf("seed legacy sqlite: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close sqlite: %v", err)
	}

	store, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	if exists, err := store.hasColumn(ctx, "issues", "project_id"); err != nil || !exists {
		t.Fatalf("issues.project_id exists = %v err = %v", exists, err)
	}
	issues, err := store.Issues(ctx)
	if err != nil {
		t.Fatalf("list issues: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("issues = %+v, want empty", issues)
	}
	var commentCount int
	if err := store.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM comments`).Scan(&commentCount); err != nil {
		t.Fatalf("count comments: %v", err)
	}
	if commentCount != 0 {
		t.Fatalf("comment count = %d, want 0", commentCount)
	}
	var remainingAttachmentID string
	if err := store.db.QueryRowContext(ctx, `SELECT id FROM attachments`).Scan(&remainingAttachmentID); err != nil {
		t.Fatalf("read remaining attachment: %v", err)
	}
	if remainingAttachmentID != "other_att" {
		t.Fatalf("remaining attachment = %q", remainingAttachmentID)
	}
}

func TestAttachmentStoreCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	created, err := store.CreateAttachment(ctx, entity.CreateAttachmentInput{
		ID:          "att_test",
		EntityType:  entity.AttachmentEntityIssue,
		EntityID:    "42",
		Filename:    "screenshot.png",
		Path:        "system/data/attachments/issue/42/att_test.png",
		ContentType: "image/png",
		Size:        8,
	})
	if err != nil {
		t.Fatalf("create attachment: %v", err)
	}
	if created.ID != "att_test" || created.Path == "" || created.CreatedAt.IsZero() {
		t.Fatalf("attachment = %+v", created)
	}

	attachments, err := store.AttachmentsByEntity(ctx, entity.AttachmentEntityIssue, "42")
	if err != nil {
		t.Fatalf("list attachments: %v", err)
	}
	if len(attachments) != 1 || attachments[0].ID != created.ID {
		t.Fatalf("attachments = %+v", attachments)
	}

	deleted, err := store.DeleteAttachment(ctx, created.ID)
	if err != nil {
		t.Fatalf("delete attachment: %v", err)
	}
	if deleted.ID != created.ID {
		t.Fatalf("deleted = %+v", deleted)
	}
	if _, err := store.Attachment(ctx, created.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}
}

func TestAttachmentStorageSaveResolveAndDelete(t *testing.T) {
	t.Parallel()

	storage := NewAttachmentStorage(t.TempDir())
	input, err := storage.Save(entity.AttachmentEntityIssue, "7", "screenshot.png", "image/png", []byte("png-data"))
	if err != nil {
		t.Fatalf("save attachment: %v", err)
	}
	if input.ID == "" || input.Size != int64(len("png-data")) {
		t.Fatalf("input = %+v", input)
	}
	file, err := storage.Open(input.Path)
	if err != nil {
		t.Fatalf("open attachment: %v", err)
	}
	file.Close()
	if err := storage.Delete(input.Path); err != nil {
		t.Fatalf("delete attachment: %v", err)
	}
	if _, err := storage.Resolve("../escape.png"); err == nil {
		t.Fatal("expected escape path error")
	}
}

func TestReadAttachmentBytesRejectsTooLargeFile(t *testing.T) {
	t.Parallel()

	_, err := ReadAttachmentBytes(strings.NewReader("abcdef"), 5)
	if err == nil {
		t.Fatal("expected size error")
	}
}

func TestCreateComment(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "COMMENTS")
	issue, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Add comments", Status: entity.StatusBacklog})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	comment, err := store.CreateComment(ctx, entity.CreateCommentInput{
		IssueID: issue.ID,
		Author:  "codex",
		Type:    entity.CommentProgress,
		Body:    "Implemented storage.",
	})
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	if comment.ID == 0 || comment.IssueID != issue.ID || comment.Author != "codex" || comment.Type != entity.CommentProgress || comment.Body != "Implemented storage." {
		t.Fatalf("comment = %+v", comment)
	}
	if comment.CreatedAt.IsZero() {
		t.Fatal("comment created_at is zero")
	}

	defaultType, err := store.CreateComment(ctx, entity.CreateCommentInput{
		IssueID: issue.ID,
		Author:  "codex",
		Body:    "Default type.",
	})
	if err != nil {
		t.Fatalf("create default type comment: %v", err)
	}
	if defaultType.Type != entity.CommentGeneral {
		t.Fatalf("default type = %q", defaultType.Type)
	}
}

func TestCreateCommentRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "VALIDATE")
	issue, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Validate comments"})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	tests := []struct {
		name  string
		input entity.CreateCommentInput
	}{
		{
			name:  "empty author",
			input: entity.CreateCommentInput{IssueID: issue.ID, Body: "body"},
		},
		{
			name:  "empty body",
			input: entity.CreateCommentInput{IssueID: issue.ID, Author: "codex"},
		},
		{
			name:  "body too long",
			input: entity.CreateCommentInput{IssueID: issue.ID, Author: "codex", Body: strings.Repeat("x", 10001)},
		},
		{
			name:  "invalid type",
			input: entity.CreateCommentInput{IssueID: issue.ID, Author: "codex", Type: entity.CommentType("note"), Body: "body"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := store.CreateComment(ctx, tt.input); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestCreateCommentRequiresExistingIssue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	_, err = store.CreateComment(ctx, entity.CreateCommentInput{
		IssueID: 999999,
		Author:  "codex",
		Type:    entity.CommentGeneral,
		Body:    "Missing issue.",
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}
}

func TestCommentsByIssueID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "LIST")
	issue, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "List comments"})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	first := createComment(t, store, issue.ID, "first")
	second := createComment(t, store, issue.ID, "second")
	third := createComment(t, store, issue.ID, "third")

	comments, err := store.CommentsByIssueID(ctx, issue.ID, 0, 2)
	if err != nil {
		t.Fatalf("list comments: %v", err)
	}
	assertCommentIDs(t, comments, []int64{first.ID, second.ID})

	comments, err = store.CommentsByIssueID(ctx, issue.ID, second.ID, 50)
	if err != nil {
		t.Fatalf("list comments after cursor: %v", err)
	}
	assertCommentIDs(t, comments, []int64{third.ID})

	comments, err = store.CommentsByIssueID(ctx, issue.ID, third.ID, 50)
	if err != nil {
		t.Fatalf("list comments after final cursor: %v", err)
	}
	if comments == nil || len(comments) != 0 {
		t.Fatalf("comments = %+v, want empty slice", comments)
	}
}

func TestCommentsByIssueIDClampsLimit(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "CLAMP")
	issue, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Clamp comments"})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	for i := 0; i < 105; i++ {
		createComment(t, store, issue.ID, "comment")
	}
	comments, err := store.CommentsByIssueID(ctx, issue.ID, 0, 101)
	if err != nil {
		t.Fatalf("list comments: %v", err)
	}
	if len(comments) != 100 {
		t.Fatalf("comment count = %d, want 100", len(comments))
	}
}

func TestProjectCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	created, err := store.CreateProject(ctx, entity.CreateProjectInput{
		Key:         "PRODUCT",
		Name:        "Product Website",
		Description: "Public marketing and product site",
		Location:    t.TempDir(),
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	if created.ID == 0 {
		t.Fatal("created project id is zero")
	}
	if created.Key != "PRODUCT" || created.Name != "Product Website" || created.Description != "Public marketing and product site" || created.Location == "" {
		t.Fatalf("created project = %+v", created)
	}

	renamed := "Product Experience"
	updated, err := store.UpdateProject(ctx, created.ID, entity.UpdateProjectInput{Name: &renamed})
	if err != nil {
		t.Fatalf("update project: %v", err)
	}
	if updated.Name != "Product Experience" {
		t.Fatalf("updated project name = %q", updated.Name)
	}

	projects, err := store.Projects(ctx)
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("project count = %d", len(projects))
	}

	if err := store.DeleteProject(ctx, created.ID); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	projects, err = store.Projects(ctx)
	if err != nil {
		t.Fatalf("list projects after delete: %v", err)
	}
	if len(projects) != 0 {
		t.Fatalf("project count after delete = %d", len(projects))
	}
}

func TestDeleteProjectRejectsLinkedIssues(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "LINKED")
	if _, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Linked issue"}); err != nil {
		t.Fatalf("create issue: %v", err)
	}
	if err := store.DeleteProject(ctx, project.ID); err == nil || !strings.Contains(err.Error(), "linked issues") {
		t.Fatalf("err = %v, want linked issues error", err)
	}
	if _, err := store.Project(ctx, project.ID); err != nil {
		t.Fatalf("project should remain: %v", err)
	}
}

func TestProjectsReturnsEmptyArray(t *testing.T) {
	t.Parallel()

	store, err := Open(context.Background(), filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	projects, err := store.Projects(context.Background())
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}
	if projects == nil {
		t.Fatal("projects is nil")
	}

	payload, err := json.Marshal(projects)
	if err != nil {
		t.Fatalf("marshal projects: %v", err)
	}
	if string(payload) != "[]" {
		t.Fatalf("projects json = %s", payload)
	}
}

func TestWorkspaceCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project, err := store.CreateProject(ctx, entity.CreateProjectInput{Key: "API", Name: "API Backend", Location: t.TempDir()})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	workspacePath := t.TempDir()
	created, err := store.CreateWorkspace(ctx, entity.CreateWorkspaceInput{
		ProjectID: project.ID,
		Name:      "API Main",
		Path:      workspacePath,
	})
	if err != nil {
		t.Fatalf("create workspace: %v", err)
	}
	if created.Status != entity.WorkspaceActive {
		t.Fatalf("created workspace status = %q", created.Status)
	}

	status := entity.WorkspaceInactive
	updated, err := store.UpdateWorkspace(ctx, created.ID, entity.UpdateWorkspaceInput{Status: &status})
	if err != nil {
		t.Fatalf("update workspace: %v", err)
	}
	if updated.Status != entity.WorkspaceInactive {
		t.Fatalf("updated workspace status = %q", updated.Status)
	}

	workspaces, err := store.Workspaces(ctx)
	if err != nil {
		t.Fatalf("list workspaces: %v", err)
	}
	if len(workspaces) != 1 {
		t.Fatalf("workspace count = %d", len(workspaces))
	}

	if err := store.DeleteWorkspace(ctx, created.ID); err != nil {
		t.Fatalf("delete workspace: %v", err)
	}
	workspaces, err = store.Workspaces(ctx)
	if err != nil {
		t.Fatalf("list workspaces after delete: %v", err)
	}
	if len(workspaces) != 0 {
		t.Fatalf("workspace count after delete = %d", len(workspaces))
	}
}

func TestIssuesByStates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "ISSUES")
	backlog, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Plan tracker", Status: entity.StatusBacklog})
	if err != nil {
		t.Fatalf("create backlog issue: %v", err)
	}
	ready, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Build worker", Status: entity.StatusReady})
	if err != nil {
		t.Fatalf("create ready issue: %v", err)
	}
	review, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Review API", Status: entity.StatusReview})
	if err != nil {
		t.Fatalf("create review issue: %v", err)
	}

	filtered, err := store.IssuesByStates(ctx, []entity.Status{entity.StatusReady, entity.StatusReview})
	if err != nil {
		t.Fatalf("list issues by states: %v", err)
	}
	assertIssueIDs(t, filtered, []int64{ready.ID, review.ID})
	assertIssueStatuses(t, filtered, []entity.Status{entity.StatusReady, entity.StatusReview})

	allByStates, err := store.IssuesByStates(ctx, nil)
	if err != nil {
		t.Fatalf("list issues with empty states: %v", err)
	}
	all, err := store.Issues(ctx)
	if err != nil {
		t.Fatalf("list all issues: %v", err)
	}
	assertIssueIDs(t, allByStates, []int64{backlog.ID, ready.ID, review.ID})
	if len(allByStates) != len(all) {
		t.Fatalf("empty states issue count = %d, all issue count = %d", len(allByStates), len(all))
	}
}

func TestIssuesByFilterFiltersByProject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	primary := createTestProject(t, store, "PRIMARY")
	secondary := createTestProject(t, store, "SECONDARY")
	ready, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: primary.ID, Title: "Primary ready", Status: entity.StatusReady})
	if err != nil {
		t.Fatalf("create primary ready issue: %v", err)
	}
	if _, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: primary.ID, Title: "Primary backlog", Status: entity.StatusBacklog}); err != nil {
		t.Fatalf("create primary backlog issue: %v", err)
	}
	other, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: secondary.ID, Title: "Secondary ready", Status: entity.StatusReady})
	if err != nil {
		t.Fatalf("create secondary issue: %v", err)
	}

	filtered, err := store.IssuesByFilter(ctx, IssueFilter{States: []entity.Status{entity.StatusReady}, ProjectID: &primary.ID})
	if err != nil {
		t.Fatalf("list issues by filter: %v", err)
	}
	assertIssueIDs(t, filtered, []int64{ready.ID})
	if filtered[0].ProjectID != primary.ID || filtered[0].ProjectKey != primary.Key {
		t.Fatalf("filtered issue project = %+v", filtered[0])
	}

	secondaryIssues, err := store.IssuesByFilter(ctx, IssueFilter{ProjectID: &secondary.ID})
	if err != nil {
		t.Fatalf("list secondary issues: %v", err)
	}
	assertIssueIDs(t, secondaryIssues, []int64{other.ID})
}

func TestIssueStatesByIDs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "STATES")
	backlog, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Plan tracker", Status: entity.StatusBacklog})
	if err != nil {
		t.Fatalf("create backlog issue: %v", err)
	}
	ready, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Build worker", Status: entity.StatusReady})
	if err != nil {
		t.Fatalf("create ready issue: %v", err)
	}
	review, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Review API", Status: entity.StatusReview})
	if err != nil {
		t.Fatalf("create review issue: %v", err)
	}

	states, err := store.IssueStatesByIDs(ctx, []int64{review.ID, 999999, backlog.ID, ready.ID})
	if err != nil {
		t.Fatalf("list issue states by ids: %v", err)
	}
	if len(states) != 3 {
		t.Fatalf("issue state count = %d, states = %+v", len(states), states)
	}
	got := map[int64]entity.Status{}
	for _, state := range states {
		got[state.ID] = state.Status
	}
	for id, want := range map[int64]entity.Status{
		backlog.ID: entity.StatusBacklog,
		ready.ID:   entity.StatusReady,
		review.ID:  entity.StatusReview,
	} {
		if got[id] != want {
			t.Fatalf("issue state for id %d = %q, want %q", id, got[id], want)
		}
	}
	if _, ok := got[999999]; ok {
		t.Fatal("missing issue id was returned")
	}
}

func TestIssueStatesByIDsWithEmptyIDsDoesNotQuery(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := Open(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := store.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}

	states, err := store.IssueStatesByIDs(ctx, nil)
	if err != nil {
		t.Fatalf("list issue states by empty ids: %v", err)
	}
	if states == nil {
		t.Fatal("issue states is nil")
	}
	if len(states) != 0 {
		t.Fatalf("issue state count = %d", len(states))
	}
}

func TestSummaryReturnsColumnsWithoutRuns(t *testing.T) {
	t.Parallel()

	store, err := Open(context.Background(), filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	summary, err := store.Summary(context.Background())
	if err != nil {
		t.Fatalf("read summary: %v", err)
	}
	payload, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("decode summary: %v", err)
	}
	if _, ok := decoded["runs"]; ok {
		t.Fatalf("summary json contains runs: %s", payload)
	}
}

func schemaObjectExists(t *testing.T, store *Store, name string) bool {
	t.Helper()

	var exists bool
	err := store.db.QueryRow(`SELECT EXISTS (
		SELECT 1 FROM sqlite_master WHERE name = ?
	)`, name).Scan(&exists)
	if err != nil {
		t.Fatalf("query schema object %q: %v", name, err)
	}
	return exists
}

func assertIssueIDs(t *testing.T, issues []entity.Issue, want []int64) {
	t.Helper()

	got := map[int64]bool{}
	for _, issue := range issues {
		got[issue.ID] = true
	}
	for _, id := range want {
		if !got[id] {
			t.Fatalf("issue id %d is missing from %+v", id, issues)
		}
	}
	if len(got) != len(want) {
		t.Fatalf("issue ids count = %d, want %d, issues = %+v", len(got), len(want), issues)
	}
}

func assertIssueStatuses(t *testing.T, issues []entity.Issue, want []entity.Status) {
	t.Helper()

	allowed := map[entity.Status]bool{}
	for _, status := range want {
		allowed[status] = true
	}
	for _, issue := range issues {
		if !allowed[issue.Status] {
			t.Fatalf("issue status %q is not expected in %+v", issue.Status, issues)
		}
	}
}

func createTestProject(t *testing.T, store *Store, key string) entity.Project {
	t.Helper()

	project, err := store.CreateProject(context.Background(), entity.CreateProjectInput{
		Key:      key,
		Name:     key,
		Location: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	return project
}

func createComment(t *testing.T, store *Store, issueID int64, body string) entity.Comment {
	t.Helper()

	comment, err := store.CreateComment(context.Background(), entity.CreateCommentInput{
		IssueID: issueID,
		Author:  "codex",
		Type:    entity.CommentGeneral,
		Body:    body,
	})
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	return comment
}

func assertCommentIDs(t *testing.T, comments []entity.Comment, want []int64) {
	t.Helper()

	if len(comments) != len(want) {
		t.Fatalf("comment length = %d, want %d; comments = %+v", len(comments), len(want), comments)
	}
	for i, comment := range comments {
		if comment.ID != want[i] {
			t.Fatalf("comments[%d].id = %d, want %d; comments = %+v", i, comment.ID, want[i], comments)
		}
	}
}
