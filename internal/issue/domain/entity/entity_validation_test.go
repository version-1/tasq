package entity

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   CreateIssueInput
		wantErr bool
	}{
		{name: "valid minimal", input: CreateIssueInput{Title: "Issue"}},
		{name: "empty title", input: CreateIssueInput{}, wantErr: true},
		{name: "title too long", input: CreateIssueInput{Title: strings.Repeat("x", 501)}, wantErr: true},
		{name: "description too long", input: CreateIssueInput{Title: "Issue", Description: strings.Repeat("x", 10001)}, wantErr: true},
		{name: "assignee too long", input: CreateIssueInput{Title: "Issue", Assignee: strings.Repeat("x", 201)}, wantErr: true},
		{name: "invalid status", input: CreateIssueInput{Title: "Issue", Status: Status("unknown")}, wantErr: true},
		{name: "invalid priority", input: CreateIssueInput{Title: "Issue", Priority: Priority("unknown")}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeCreate(tt.input)
			assertErr(t, err, tt.wantErr)
		})
	}
}

func TestNormalizeUpdateIssue(t *testing.T) {
	t.Parallel()

	validTitle := "Issue"
	emptyTitle := ""
	longTitle := strings.Repeat("x", 501)
	longDescription := strings.Repeat("x", 10001)
	invalidStatus := Status("unknown")
	invalidPriority := Priority("unknown")
	longAssignee := strings.Repeat("x", 201)
	tests := []struct {
		name    string
		input   UpdateIssueInput
		wantErr bool
	}{
		{name: "valid", input: UpdateIssueInput{Title: &validTitle}},
		{name: "empty title", input: UpdateIssueInput{Title: &emptyTitle}, wantErr: true},
		{name: "title too long", input: UpdateIssueInput{Title: &longTitle}, wantErr: true},
		{name: "description too long", input: UpdateIssueInput{Description: &longDescription}, wantErr: true},
		{name: "invalid status", input: UpdateIssueInput{Status: &invalidStatus}, wantErr: true},
		{name: "invalid priority", input: UpdateIssueInput{Priority: &invalidPriority}, wantErr: true},
		{name: "assignee too long", input: UpdateIssueInput{Assignee: &longAssignee}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeUpdateIssue(tt.input)
			assertErr(t, err, tt.wantErr)
		})
	}
}

func TestNormalizeCreateProject(t *testing.T) {
	t.Parallel()

	location := t.TempDir()
	tests := []struct {
		name    string
		input   CreateProjectInput
		wantErr bool
	}{
		{name: "valid", input: CreateProjectInput{Key: "PROJ_1", Name: "Project", Location: location}},
		{name: "valid lowercase kebab", input: CreateProjectInput{Key: "project-api", Name: "Project", Location: location}},
		{name: "empty key", input: CreateProjectInput{Name: "Project", Location: location}, wantErr: true},
		{name: "invalid key uppercase hyphen", input: CreateProjectInput{Key: "PROJ-API", Name: "Project", Location: location}, wantErr: true},
		{name: "invalid key starts with number", input: CreateProjectInput{Key: "1PROJ", Name: "Project", Location: location}, wantErr: true},
		{name: "key too long", input: CreateProjectInput{Key: "PROJECT_KEY_TOO_LONG_1", Name: "Project", Location: location}, wantErr: true},
		{name: "empty name", input: CreateProjectInput{Key: "PROJ", Location: location}, wantErr: true},
		{name: "name too long", input: CreateProjectInput{Key: "PROJ", Name: strings.Repeat("x", 201), Location: location}, wantErr: true},
		{name: "description too long", input: CreateProjectInput{Key: "PROJ", Name: "Project", Description: strings.Repeat("x", 10001), Location: location}, wantErr: true},
		{name: "relative location", input: CreateProjectInput{Key: "PROJ", Name: "Project", Location: "relative"}, wantErr: true},
		{name: "missing location", input: CreateProjectInput{Key: "PROJ", Name: "Project", Location: filepath.Join(location, "missing")}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeCreateProject(tt.input)
			assertErr(t, err, tt.wantErr)
		})
	}
}

func TestNormalizeUpdateProject(t *testing.T) {
	t.Parallel()

	location := t.TempDir()
	key := "project-api"
	invalidKey := "Project-API"
	name := "Project"
	emptyName := ""
	longName := strings.Repeat("x", 201)
	longDescription := strings.Repeat("x", 10001)
	relativeLocation := "relative"
	missingLocation := filepath.Join(location, "missing")
	tests := []struct {
		name    string
		input   UpdateProjectInput
		wantErr bool
	}{
		{name: "valid", input: UpdateProjectInput{Key: &key, Name: &name, Location: &location}},
		{name: "invalid key", input: UpdateProjectInput{Key: &invalidKey}, wantErr: true},
		{name: "empty name", input: UpdateProjectInput{Name: &emptyName}, wantErr: true},
		{name: "name too long", input: UpdateProjectInput{Name: &longName}, wantErr: true},
		{name: "description too long", input: UpdateProjectInput{Description: &longDescription}, wantErr: true},
		{name: "relative location", input: UpdateProjectInput{Location: &relativeLocation}, wantErr: true},
		{name: "missing location", input: UpdateProjectInput{Location: &missingLocation}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeUpdateProject(tt.input)
			assertErr(t, err, tt.wantErr)
		})
	}
}

func TestNormalizeCreateWorkspace(t *testing.T) {
	t.Parallel()

	path := t.TempDir()
	tests := []struct {
		name    string
		input   CreateWorkspaceInput
		wantErr bool
	}{
		{name: "valid", input: CreateWorkspaceInput{ProjectID: 1, Name: "Workspace", Path: path}},
		{name: "missing project", input: CreateWorkspaceInput{Name: "Workspace", Path: path}, wantErr: true},
		{name: "empty name", input: CreateWorkspaceInput{ProjectID: 1, Path: path}, wantErr: true},
		{name: "name too long", input: CreateWorkspaceInput{ProjectID: 1, Name: strings.Repeat("x", 201), Path: path}, wantErr: true},
		{name: "relative path", input: CreateWorkspaceInput{ProjectID: 1, Name: "Workspace", Path: "relative"}, wantErr: true},
		{name: "missing path", input: CreateWorkspaceInput{ProjectID: 1, Name: "Workspace", Path: filepath.Join(path, "missing")}, wantErr: true},
		{name: "invalid status", input: CreateWorkspaceInput{ProjectID: 1, Name: "Workspace", Path: path, Status: WorkspaceStatus("unknown")}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeCreateWorkspace(tt.input)
			assertErr(t, err, tt.wantErr)
		})
	}
}

func TestNormalizeUpdateWorkspace(t *testing.T) {
	t.Parallel()

	path := t.TempDir()
	projectID := int64(1)
	invalidProjectID := int64(0)
	name := "Workspace"
	emptyName := ""
	longName := strings.Repeat("x", 201)
	relativePath := "relative"
	missingPath := filepath.Join(path, "missing")
	status := WorkspaceInactive
	invalidStatus := WorkspaceStatus("unknown")
	tests := []struct {
		name    string
		input   UpdateWorkspaceInput
		wantErr bool
	}{
		{name: "valid", input: UpdateWorkspaceInput{ProjectID: &projectID, Name: &name, Path: &path, Status: &status}},
		{name: "invalid project", input: UpdateWorkspaceInput{ProjectID: &invalidProjectID}, wantErr: true},
		{name: "empty name", input: UpdateWorkspaceInput{Name: &emptyName}, wantErr: true},
		{name: "name too long", input: UpdateWorkspaceInput{Name: &longName}, wantErr: true},
		{name: "relative path", input: UpdateWorkspaceInput{Path: &relativePath}, wantErr: true},
		{name: "missing path", input: UpdateWorkspaceInput{Path: &missingPath}, wantErr: true},
		{name: "invalid status", input: UpdateWorkspaceInput{Status: &invalidStatus}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeUpdateWorkspace(tt.input)
			assertErr(t, err, tt.wantErr)
		})
	}
}

func assertErr(t *testing.T, err error, wantErr bool) {
	t.Helper()
	if wantErr && err == nil {
		t.Fatal("expected error")
	}
	if !wantErr && err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
