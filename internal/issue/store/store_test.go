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

	store, err := OpenMigrated(context.Background(), filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	for _, name := range []string{
		"issues",
		"issues_project_id_idx",
		"issues_status_idx",
		"issue_dependencies",
		"issue_dependencies_dependency_issue_id_idx",
		"comments",
		"comments_issue_id_idx",
		"projects",
		"projects_key_idx",
		"project_workflows",
		"attachments",
		"attachments_entity_idx",
	} {
		if !schemaObjectExists(t, store, name) {
			t.Fatalf("schema object %q does not exist", name)
		}
	}
}

func TestOpenRejectsPendingIssueTrackerMigrations(t *testing.T) {
	t.Parallel()

	_, err := Open(context.Background(), filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err == nil {
		t.Fatal("open store succeeded, want pending migration error")
	}
	if !strings.Contains(err.Error(), "run `tq migrate`") {
		t.Fatalf("error = %v, want tq migrate guidance", err)
	}
}

func TestProjectWorkflowCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "WORKFLOW")
	created, err := store.UpsertProjectWorkflow(ctx, entity.UpsertProjectWorkflowInput{
		ProjectID:       project.ID,
		FrontmatterJSON: `{"agent":{"max_turns":3},"tracker":{"kind":"tasq"}}`,
		Body:            "Use tq to keep the issue tracker synchronized.",
		Checksum:        strings.Repeat("a", 64),
	})
	if err != nil {
		t.Fatalf("upsert project workflow: %v", err)
	}
	if created.ProjectID != project.ID || created.Body == "" || created.Checksum != strings.Repeat("a", 64) {
		t.Fatalf("created workflow = %+v", created)
	}
	agent, ok := created.Frontmatter["agent"].(map[string]any)
	if !ok || agent["max_turns"] != float64(3) {
		t.Fatalf("created workflow frontmatter = %#v", created.Frontmatter)
	}
	if created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("created workflow timestamps = %+v", created)
	}

	updated, err := store.UpsertProjectWorkflow(ctx, entity.UpsertProjectWorkflowInput{
		ProjectID:       project.ID,
		FrontmatterJSON: `{"tasq":{"task_work_prompt":false}}`,
		Body:            "Updated prompt template.",
		Checksum:        strings.Repeat("b", 64),
	})
	if err != nil {
		t.Fatalf("update project workflow: %v", err)
	}
	if updated.ProjectID != project.ID || updated.Body != "Updated prompt template." || updated.Checksum != strings.Repeat("b", 64) {
		t.Fatalf("updated workflow = %+v", updated)
	}
	tasq, ok := updated.Frontmatter["tasq"].(map[string]any)
	if !ok || tasq["task_work_prompt"] != false {
		t.Fatalf("updated workflow frontmatter = %#v", updated.Frontmatter)
	}
	if updated.CreatedAt.IsZero() || updated.UpdatedAt.IsZero() {
		t.Fatalf("updated workflow timestamps = %+v", updated)
	}

	skipped, err := store.UpsertProjectWorkflow(ctx, entity.UpsertProjectWorkflowInput{
		ProjectID:       project.ID,
		FrontmatterJSON: `{"tracker":{"kind":"changed"}}`,
		Body:            "This body should not be saved.",
		Checksum:        strings.Repeat("b", 64),
	})
	if err != nil {
		t.Fatalf("skip matching checksum workflow: %v", err)
	}
	if skipped.ProjectID != updated.ProjectID || skipped.Body != updated.Body || skipped.Checksum != updated.Checksum || !skipped.CreatedAt.Equal(updated.CreatedAt) || !skipped.UpdatedAt.Equal(updated.UpdatedAt) {
		t.Fatalf("skipped workflow = %+v, want existing %+v", skipped, updated)
	}
	tasq, ok = skipped.Frontmatter["tasq"].(map[string]any)
	if !ok || tasq["task_work_prompt"] != false {
		t.Fatalf("skipped workflow frontmatter = %#v", skipped.Frontmatter)
	}

	read, err := store.ProjectWorkflow(ctx, project.ID)
	if err != nil {
		t.Fatalf("read project workflow: %v", err)
	}
	if read.ProjectID != updated.ProjectID || read.Body != updated.Body || read.Checksum != updated.Checksum || !read.CreatedAt.Equal(updated.CreatedAt) || !read.UpdatedAt.Equal(updated.UpdatedAt) {
		t.Fatalf("read workflow = %+v, want %+v", read, updated)
	}
	tasq, ok = read.Frontmatter["tasq"].(map[string]any)
	if !ok || tasq["task_work_prompt"] != false {
		t.Fatalf("read workflow frontmatter = %#v", read.Frontmatter)
	}

	if err := store.DeleteProjectWorkflow(ctx, project.ID); err != nil {
		t.Fatalf("delete project workflow: %v", err)
	}
	if _, err := store.ProjectWorkflow(ctx, project.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}
	if err := store.DeleteProjectWorkflow(ctx, project.ID); err != nil {
		t.Fatalf("delete missing project workflow: %v", err)
	}
}

func TestDeleteProjectDeletesProjectWorkflow(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "WFDELETE")
	if _, err := store.UpsertProjectWorkflow(ctx, entity.UpsertProjectWorkflowInput{
		ProjectID:       project.ID,
		FrontmatterJSON: `{}`,
		Checksum:        strings.Repeat("e", 64),
	}); err != nil {
		t.Fatalf("upsert project workflow: %v", err)
	}

	if err := store.DeleteProject(ctx, project.ID); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	if _, err := store.ProjectWorkflow(ctx, project.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("workflow err = %v, want sql.ErrNoRows", err)
	}
}

func TestProjectWorkflowRequiresExistingProject(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	_, err = store.UpsertProjectWorkflow(ctx, entity.UpsertProjectWorkflowInput{
		ProjectID:       999999,
		FrontmatterJSON: `{}`,
		Checksum:        strings.Repeat("c", 64),
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err = %v, want sql.ErrNoRows", err)
	}
}

func TestProjectWorkflowRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "WFVALID")
	tests := []struct {
		name  string
		input entity.UpsertProjectWorkflowInput
	}{
		{
			name:  "missing project",
			input: entity.UpsertProjectWorkflowInput{FrontmatterJSON: `{}`, Checksum: strings.Repeat("d", 64)},
		},
		{
			name:  "invalid frontmatter json",
			input: entity.UpsertProjectWorkflowInput{ProjectID: project.ID, FrontmatterJSON: `{`, Checksum: strings.Repeat("d", 64)},
		},
		{
			name:  "invalid checksum",
			input: entity.UpsertProjectWorkflowInput{ProjectID: project.ID, FrontmatterJSON: `{}`, Checksum: "not-sha256"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := store.UpsertProjectWorkflow(ctx, tt.input); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestAttachmentStoreCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
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
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
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
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
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
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
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
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
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
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
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

func TestCommentCountsByIssueID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "COUNTS")
	withComments, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "With comments"})
	if err != nil {
		t.Fatalf("create issue with comments: %v", err)
	}
	withoutComments, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Without comments"})
	if err != nil {
		t.Fatalf("create issue without comments: %v", err)
	}
	for _, body := range []string{"first", "second"} {
		if _, err := store.CreateComment(ctx, entity.CreateCommentInput{IssueID: withComments.ID, Author: "tester", Body: body}); err != nil {
			t.Fatalf("create comment: %v", err)
		}
	}

	counts, err := store.commentCountsByIssueID(ctx)
	if err != nil {
		t.Fatalf("count comments: %v", err)
	}
	if counts[withComments.ID] != 2 {
		t.Fatalf("comment count for issue with comments = %d, want 2", counts[withComments.ID])
	}
	if _, ok := counts[withoutComments.ID]; ok {
		t.Fatalf("comment count includes issue without comments: %+v", counts)
	}
}

func TestProjectCRUD(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
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

	checksum := strings.Repeat("a", 64)
	workflow, err := store.UpsertProjectWorkflow(ctx, entity.UpsertProjectWorkflowInput{
		ProjectID:       created.ID,
		FrontmatterJSON: `{"agent":{"max_turns":3}}`,
		Body:            "Project prompt",
		Checksum:        checksum,
	})
	if err != nil {
		t.Fatalf("upsert project workflow: %v", err)
	}
	if workflow.Body != "Project prompt" || workflow.Checksum != checksum {
		t.Fatalf("workflow = %+v", workflow)
	}
	readWorkflow, err := store.ProjectWorkflow(ctx, created.ID)
	if err != nil {
		t.Fatalf("project workflow: %v", err)
	}
	if readWorkflow.Body != workflow.Body || readWorkflow.Checksum != workflow.Checksum {
		t.Fatalf("read workflow = %+v, want %+v", readWorkflow, workflow)
	}
	withWorkflowChecksum, err := store.Project(ctx, created.ID)
	if err != nil {
		t.Fatalf("project after workflow update: %v", err)
	}
	if withWorkflowChecksum.WorkflowChecksum != checksum {
		t.Fatalf("project workflow checksum = %q", withWorkflowChecksum.WorkflowChecksum)
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
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
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

func TestDeleteProjectWorkflowRemovesOverride(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "WORKFLOW")
	now := nowString()
	if _, err := store.db.ExecContext(ctx, `INSERT INTO project_workflows (
		project_id, frontmatter_json, body, checksum, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?)`, project.ID, `{}`, "custom workflow", strings.Repeat("a", 64), now, now); err != nil {
		t.Fatalf("insert project workflow: %v", err)
	}

	if err := store.DeleteProjectWorkflow(ctx, project.ID); err != nil {
		t.Fatalf("delete project workflow: %v", err)
	}
	var count int
	if err := store.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM project_workflows WHERE project_id = ?`, project.ID).Scan(&count); err != nil {
		t.Fatalf("count project workflows: %v", err)
	}
	if count != 0 {
		t.Fatalf("workflow count = %d, want 0", count)
	}
}

func TestDeleteProjectWorkflowAllowsMissingOverride(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "NOWORKFLOW")
	if err := store.DeleteProjectWorkflow(ctx, project.ID); err != nil {
		t.Fatalf("delete missing project workflow: %v", err)
	}
}

func TestDeleteProjectRemovesWorkflowOverride(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "PROJECTFLOW")
	now := nowString()
	if _, err := store.db.ExecContext(ctx, `INSERT INTO project_workflows (
		project_id, frontmatter_json, body, checksum, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?)`, project.ID, `{}`, "custom workflow", strings.Repeat("b", 64), now, now); err != nil {
		t.Fatalf("insert project workflow: %v", err)
	}

	if err := store.DeleteProject(ctx, project.ID); err != nil {
		t.Fatalf("delete project: %v", err)
	}
	var count int
	if err := store.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM project_workflows WHERE project_id = ?`, project.ID).Scan(&count); err != nil {
		t.Fatalf("count project workflows: %v", err)
	}
	if count != 0 {
		t.Fatalf("workflow count = %d, want 0", count)
	}
}

func TestProjectsReturnsEmptyArray(t *testing.T) {
	t.Parallel()

	store, err := OpenMigrated(context.Background(), filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
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

func TestIssuesByStates(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
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
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
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
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
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
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
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

func TestIssueDependenciesCRUDAndCascade(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "DEPSCRUD")
	parent, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Parent"})
	if err != nil {
		t.Fatalf("create parent issue: %v", err)
	}
	first, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "First dependency"})
	if err != nil {
		t.Fatalf("create first dependency issue: %v", err)
	}
	second, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Second dependency"})
	if err != nil {
		t.Fatalf("create second dependency issue: %v", err)
	}

	if err := store.SetDependencyIDs(ctx, parent.ID, []int64{second.ID, first.ID}); err != nil {
		t.Fatalf("set dependencies: %v", err)
	}
	dependencyIDs, err := store.DependencyIDs(ctx, parent.ID)
	if err != nil {
		t.Fatalf("read dependency ids: %v", err)
	}
	assertInt64s(t, dependencyIDs, []int64{first.ID, second.ID})

	if err := store.SetDependencyIDs(ctx, parent.ID, []int64{first.ID}); err != nil {
		t.Fatalf("replace dependencies: %v", err)
	}
	read, err := store.Issue(ctx, parent.ID)
	if err != nil {
		t.Fatalf("read issue: %v", err)
	}
	assertInt64s(t, read.DependencyIDs, []int64{first.ID})

	if _, err := store.db.ExecContext(ctx, `DELETE FROM issues WHERE id = ?`, first.ID); err != nil {
		t.Fatalf("delete dependency issue: %v", err)
	}
	dependencyIDs, err = store.DependencyIDs(ctx, parent.ID)
	if err != nil {
		t.Fatalf("read dependency ids after cascade: %v", err)
	}
	assertInt64s(t, dependencyIDs, []int64{})
}

func TestCreateIssueWithDependencyIDs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "CREATEDEPS")
	first := createStoreIssue(t, store, project.ID, "First dependency", entity.StatusBacklog, entity.PriorityNormal)
	second := createStoreIssue(t, store, project.ID, "Second dependency", entity.StatusBacklog, entity.PriorityNormal)

	created, err := store.CreateIssue(ctx, entity.CreateIssueInput{
		ProjectID:     project.ID,
		Title:         "Parent issue",
		DependencyIDs: []int64{second.ID, first.ID},
	})
	if err != nil {
		t.Fatalf("create issue with dependencies: %v", err)
	}
	assertInt64s(t, created.DependencyIDs, []int64{first.ID, second.ID})

	read, err := store.Issue(ctx, created.ID)
	if err != nil {
		t.Fatalf("read issue: %v", err)
	}
	assertInt64s(t, read.DependencyIDs, []int64{first.ID, second.ID})
}

func TestCreateIssueWithoutDependencyIDsReturnsEmptyDependencies(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "NODEPS")
	created, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Independent issue"})
	if err != nil {
		t.Fatalf("create issue: %v", err)
	}
	if created.DependencyIDs == nil || len(created.DependencyIDs) != 0 {
		t.Fatalf("dependency ids = %+v, want empty slice", created.DependencyIDs)
	}
}

func TestCreateIssueRejectsSelfDependency(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "SELFDEP")
	existing := createStoreIssue(t, store, project.ID, "Existing issue", entity.StatusBacklog, entity.PriorityNormal)

	if _, err := store.CreateIssue(ctx, entity.CreateIssueInput{
		ProjectID:     project.ID,
		Title:         "Self dependency",
		DependencyIDs: []int64{existing.ID + 1},
	}); err == nil {
		t.Fatal("create issue with self dependency succeeded, want error")
	}
}

func TestCreateIssueRejectsMissingDependencyAndRollsBackIssue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "MISSINGDEP")
	if _, err := store.CreateIssue(ctx, entity.CreateIssueInput{
		ProjectID:     project.ID,
		Title:         "Missing dependency",
		DependencyIDs: []int64{999999},
	}); err == nil {
		t.Fatal("create issue with missing dependency succeeded, want error")
	}

	issues, err := store.Issues(ctx)
	if err != nil {
		t.Fatalf("list issues: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("issues after failed create = %+v, want none", issues)
	}
}

func TestCreateIssueRejectsDuplicateDependencyIDs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "DUPDEP")
	dependency := createStoreIssue(t, store, project.ID, "Dependency", entity.StatusBacklog, entity.PriorityNormal)

	if _, err := store.CreateIssue(ctx, entity.CreateIssueInput{
		ProjectID:     project.ID,
		Title:         "Duplicate dependency",
		DependencyIDs: []int64{dependency.ID, dependency.ID},
	}); err == nil {
		t.Fatal("create issue with duplicate dependency succeeded, want error")
	}
}

func TestIssueDependenciesRejectCycles(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "DAG")
	a := createStoreIssue(t, store, project.ID, "A", entity.StatusBacklog, entity.PriorityNormal)
	b := createStoreIssue(t, store, project.ID, "B", entity.StatusBacklog, entity.PriorityNormal)
	c := createStoreIssue(t, store, project.ID, "C", entity.StatusBacklog, entity.PriorityNormal)
	d := createStoreIssue(t, store, project.ID, "D", entity.StatusBacklog, entity.PriorityNormal)

	if err := store.SetDependencyIDs(ctx, a.ID, []int64{a.ID}); err == nil {
		t.Fatal("self dependency succeeded, want error")
	}
	if err := store.SetDependencyIDs(ctx, a.ID, []int64{b.ID}); err != nil {
		t.Fatalf("set A -> B: %v", err)
	}
	if err := store.SetDependencyIDs(ctx, b.ID, []int64{a.ID}); err == nil {
		t.Fatal("two-node cycle succeeded, want error")
	}
	if err := store.SetDependencyIDs(ctx, b.ID, []int64{c.ID}); err != nil {
		t.Fatalf("set B -> C: %v", err)
	}
	if err := store.SetDependencyIDs(ctx, c.ID, []int64{a.ID}); err == nil {
		t.Fatal("three-node cycle succeeded, want error")
	}
	if err := store.SetDependencyIDs(ctx, d.ID, []int64{b.ID, c.ID}); err != nil {
		t.Fatalf("legal branching DAG rejected: %v", err)
	}
}

func TestQueueClassifiesAndSortsReadyIssues(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "QUEUE")
	doneDep := createStoreIssue(t, store, project.ID, "Done dependency", entity.StatusDone, entity.PriorityNormal)
	blockedDep := createStoreIssue(t, store, project.ID, "Blocked dependency", entity.StatusBlocked, entity.PriorityNormal)
	activeDep := createStoreIssue(t, store, project.ID, "Active dependency", entity.StatusInProgress, entity.PriorityNormal)
	low := createStoreIssue(t, store, project.ID, "Low ready", entity.StatusReady, entity.PriorityLow)
	high := createStoreIssue(t, store, project.ID, "High ready", entity.StatusReady, entity.PriorityHigh)
	urgent := createStoreIssue(t, store, project.ID, "Urgent ready", entity.StatusReady, entity.PriorityUrgent)
	pending := createStoreIssue(t, store, project.ID, "Pending ready", entity.StatusReady, entity.PriorityUrgent)
	normal := createStoreIssue(t, store, project.ID, "Normal ready", entity.StatusReady, entity.PriorityNormal)
	_ = createStoreIssue(t, store, project.ID, "Backlog issue", entity.StatusBacklog, entity.PriorityUrgent)
	if err := store.SetDependencyIDs(ctx, urgent.ID, []int64{doneDep.ID, blockedDep.ID}); err != nil {
		t.Fatalf("set resolved dependencies: %v", err)
	}
	if err := store.SetDependencyIDs(ctx, pending.ID, []int64{activeDep.ID, doneDep.ID}); err != nil {
		t.Fatalf("set pending dependencies: %v", err)
	}

	queue, err := store.Queue(ctx, IssueFilter{})
	if err != nil {
		t.Fatalf("queue: %v", err)
	}
	assertQueueIssueIDs(t, queue.Queued, []int64{urgent.ID, high.ID, normal.ID, low.ID})
	assertQueueIssueIDs(t, queue.Pending, []int64{pending.ID})
	assertInt64s(t, queue.Pending[0].BlockedDependencyIDs, []int64{activeDep.ID})
}

func TestSummaryReturnsColumnsWithIssueStats(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "STATS")
	withComments, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "With comments", Status: entity.StatusReady})
	if err != nil {
		t.Fatalf("create issue with comments: %v", err)
	}
	withoutComments, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Without comments", Status: entity.StatusReady})
	if err != nil {
		t.Fatalf("create issue without comments: %v", err)
	}
	for _, body := range []string{"first comment", "second comment"} {
		if _, err := store.CreateComment(ctx, entity.CreateCommentInput{IssueID: withComments.ID, Author: "tester", Body: body}); err != nil {
			t.Fatalf("create comment: %v", err)
		}
	}

	summary, err := store.Summary(ctx)
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
	if !strings.Contains(string(payload), `"stats":{"commentCount":2}`) {
		t.Fatalf("summary json does not contain comment stats: %s", payload)
	}

	readyColumn, ok := summaryColumn(summary, entity.StatusReady)
	if !ok {
		t.Fatalf("ready column was not returned: %+v", summary.Columns)
	}
	statsByIssueID := map[int64]entity.IssueStats{}
	for _, issue := range readyColumn.Issues {
		statsByIssueID[issue.ID] = issue.Stats
	}
	if statsByIssueID[withComments.ID].CommentCount != 2 {
		t.Fatalf("comment count for issue with comments = %d, want 2", statsByIssueID[withComments.ID].CommentCount)
	}
	if statsByIssueID[withoutComments.ID].CommentCount != 0 {
		t.Fatalf("comment count for issue without comments = %d, want 0", statsByIssueID[withoutComments.ID].CommentCount)
	}
}

func TestSummaryReturnsIssueQueueStatus(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store, err := OpenMigrated(ctx, filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer store.Close()

	project := createTestProject(t, store, "QUEUESTATUS")
	backlog, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Backlog", Status: entity.StatusBacklog})
	if err != nil {
		t.Fatalf("create backlog issue: %v", err)
	}
	activeDependency, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Active dependency", Status: entity.StatusReview})
	if err != nil {
		t.Fatalf("create active dependency issue: %v", err)
	}
	satisfiedDependency, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Satisfied dependency", Status: entity.StatusDone})
	if err != nil {
		t.Fatalf("create satisfied dependency issue: %v", err)
	}
	pending, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Pending", Status: entity.StatusReady})
	if err != nil {
		t.Fatalf("create pending issue: %v", err)
	}
	queued, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Queued", Status: entity.StatusReady})
	if err != nil {
		t.Fatalf("create queued issue: %v", err)
	}
	processing, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Processing", Status: entity.StatusInProgress})
	if err != nil {
		t.Fatalf("create processing issue: %v", err)
	}
	inactiveStatuses := []entity.Status{
		entity.StatusReview,
		entity.StatusBlocked,
		entity.StatusFailed,
		entity.StatusCancelled,
		entity.StatusDuplicate,
	}
	inactiveIssueIDs := []int64{}
	for _, status := range inactiveStatuses {
		issue, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Inactive status " + string(status), Status: status})
		if err != nil {
			t.Fatalf("create %s issue: %v", status, err)
		}
		inactiveIssueIDs = append(inactiveIssueIDs, issue.ID)
	}
	completed, err := store.CreateIssue(ctx, entity.CreateIssueInput{ProjectID: project.ID, Title: "Completed", Status: entity.StatusDone})
	if err != nil {
		t.Fatalf("create completed issue: %v", err)
	}
	if err := store.SetDependencyIDs(ctx, pending.ID, []int64{activeDependency.ID}); err != nil {
		t.Fatalf("set pending dependencies: %v", err)
	}
	if err := store.SetDependencyIDs(ctx, queued.ID, []int64{satisfiedDependency.ID}); err != nil {
		t.Fatalf("set queued dependencies: %v", err)
	}

	summary, err := store.Summary(ctx)
	if err != nil {
		t.Fatalf("read summary: %v", err)
	}
	queueStatusByIssueID := summaryQueueStatuses(summary)
	want := map[int64]entity.QueueStatus{
		backlog.ID:    entity.QueueStatusBacklog,
		pending.ID:    entity.QueueStatusPending,
		queued.ID:     entity.QueueStatusQueued,
		processing.ID: entity.QueueStatusProcessing,
		completed.ID:  entity.QueueStatusCompleted,
	}
	for _, id := range inactiveIssueIDs {
		want[id] = entity.QueueStatusInactive
	}
	for id, status := range want {
		if queueStatusByIssueID[id] != status {
			t.Fatalf("queue status for issue %d = %q, want %q", id, queueStatusByIssueID[id], status)
		}
	}
}

func summaryColumn(summary entity.Summary, status entity.Status) (entity.Column, bool) {
	for _, column := range summary.Columns {
		if column.Status == status {
			return column, true
		}
	}
	return entity.Column{}, false
}

func summaryQueueStatuses(summary entity.Summary) map[int64]entity.QueueStatus {
	statuses := map[int64]entity.QueueStatus{}
	for _, column := range summary.Columns {
		for _, issue := range column.Issues {
			statuses[issue.ID] = issue.QueueStatus
		}
	}
	return statuses
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

func createStoreIssue(t *testing.T, store *Store, projectID int64, title string, status entity.Status, priority entity.Priority) entity.Issue {
	t.Helper()

	issue, err := store.CreateIssue(context.Background(), entity.CreateIssueInput{
		ProjectID: projectID,
		Title:     title,
		Status:    status,
		Priority:  priority,
	})
	if err != nil {
		t.Fatalf("create issue %q: %v", title, err)
	}
	return issue
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

func assertInt64s(t *testing.T, got []int64, want []int64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("ids = %+v, want %+v", got, want)
	}
	for i, id := range want {
		if got[i] != id {
			t.Fatalf("ids = %+v, want %+v", got, want)
		}
	}
}

func assertQueueIssueIDs(t *testing.T, issues []entity.QueueIssue, want []int64) {
	t.Helper()
	if len(issues) != len(want) {
		t.Fatalf("queue ids length = %d, want %d; issues = %+v", len(issues), len(want), issues)
	}
	for i, id := range want {
		if issues[i].ID != id {
			t.Fatalf("queue issue ids = %+v, want %+v", queueIssueIDs(issues), want)
		}
	}
}

func queueIssueIDs(issues []entity.QueueIssue) []int64 {
	ids := make([]int64, 0, len(issues))
	for _, issue := range issues {
		ids = append(ids, issue.ID)
	}
	return ids
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
