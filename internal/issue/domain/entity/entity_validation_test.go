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
		{name: "valid minimal", input: CreateIssueInput{ProjectID: 1, Title: "Issue"}},
		{name: "missing project", input: CreateIssueInput{Title: "Issue"}, wantErr: true},
		{name: "empty title", input: CreateIssueInput{}, wantErr: true},
		{name: "title too long", input: CreateIssueInput{ProjectID: 1, Title: strings.Repeat("x", 501)}, wantErr: true},
		{name: "description too long", input: CreateIssueInput{ProjectID: 1, Title: "Issue", Description: strings.Repeat("x", 10001)}, wantErr: true},
		{name: "assignee too long", input: CreateIssueInput{ProjectID: 1, Title: "Issue", Assignee: strings.Repeat("x", 201)}, wantErr: true},
		{name: "invalid status", input: CreateIssueInput{ProjectID: 1, Title: "Issue", Status: Status("unknown")}, wantErr: true},
		{name: "invalid priority", input: CreateIssueInput{ProjectID: 1, Title: "Issue", Priority: Priority("unknown")}, wantErr: true},
		{name: "valid dependency ids", input: CreateIssueInput{ProjectID: 1, Title: "Issue", DependencyIDs: []int64{1, 2}}},
		{name: "invalid dependency id", input: CreateIssueInput{ProjectID: 1, Title: "Issue", DependencyIDs: []int64{1, 0}}, wantErr: true},
		{name: "duplicate dependency id", input: CreateIssueInput{ProjectID: 1, Title: "Issue", DependencyIDs: []int64{1, 1}}, wantErr: true},
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
	validDependencyIDs := []int64{1, 2}
	invalidDependencyIDs := []int64{1, 0}
	duplicateDependencyIDs := []int64{1, 1}
	tests := []struct {
		name    string
		input   UpdateIssueInput
		wantErr bool
	}{
		{name: "valid", input: UpdateIssueInput{Title: &validTitle}},
		{name: "valid dependency ids", input: UpdateIssueInput{DependencyIDs: &validDependencyIDs}},
		{name: "empty title", input: UpdateIssueInput{Title: &emptyTitle}, wantErr: true},
		{name: "title too long", input: UpdateIssueInput{Title: &longTitle}, wantErr: true},
		{name: "description too long", input: UpdateIssueInput{Description: &longDescription}, wantErr: true},
		{name: "invalid status", input: UpdateIssueInput{Status: &invalidStatus}, wantErr: true},
		{name: "invalid priority", input: UpdateIssueInput{Priority: &invalidPriority}, wantErr: true},
		{name: "assignee too long", input: UpdateIssueInput{Assignee: &longAssignee}, wantErr: true},
		{name: "invalid dependency id", input: UpdateIssueInput{DependencyIDs: &invalidDependencyIDs}, wantErr: true},
		{name: "duplicate dependency id", input: UpdateIssueInput{DependencyIDs: &duplicateDependencyIDs}, wantErr: true},
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
		{name: "nonexistent absolute location", input: CreateProjectInput{Key: "PROJ", Name: "Project", Location: filepath.Join(location, "missing")}},
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
		{name: "nonexistent absolute location", input: UpdateProjectInput{Location: &missingLocation}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeUpdateProject(tt.input)
			assertErr(t, err, tt.wantErr)
		})
	}
}

func TestNormalizeCreateChangeRequest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   CreateChangeRequestInput
		wantErr bool
	}{
		{name: "valid", input: CreateChangeRequestInput{IssueID: 1, Author: "reviewer", Body: "Update docs."}},
		{name: "missing issue", input: CreateChangeRequestInput{Author: "reviewer", Body: "Update docs."}, wantErr: true},
		{name: "missing author", input: CreateChangeRequestInput{IssueID: 1, Body: "Update docs."}, wantErr: true},
		{name: "missing body", input: CreateChangeRequestInput{IssueID: 1, Author: "reviewer"}, wantErr: true},
		{name: "body too long", input: CreateChangeRequestInput{IssueID: 1, Author: "reviewer", Body: strings.Repeat("x", 10001)}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeCreateChangeRequest(tt.input)
			assertErr(t, err, tt.wantErr)
		})
	}
}

func TestNormalizeUpdateChangeRequest(t *testing.T) {
	t.Parallel()

	body := "Updated body"
	emptyBody := ""
	inProgress := ChangeRequestInProgress
	resolved := ChangeRequestResolved
	canceled := ChangeRequestCanceled
	invalidStatus := ChangeRequestStatus("unknown")
	runID := "run-1"
	commentID := int64(10)
	tests := []struct {
		name          string
		currentStatus ChangeRequestStatus
		input         UpdateChangeRequestInput
		wantErr       bool
	}{
		{name: "edit open body", currentStatus: ChangeRequestOpen, input: UpdateChangeRequestInput{Body: &body}},
		{name: "claim open", currentStatus: ChangeRequestOpen, input: UpdateChangeRequestInput{Status: &inProgress}},
		{name: "cancel open", currentStatus: ChangeRequestOpen, input: UpdateChangeRequestInput{Status: &canceled}},
		{name: "resolve in progress", currentStatus: ChangeRequestInProgress, input: UpdateChangeRequestInput{Status: &resolved, ResolvedByRunID: &runID, ResultCommentID: &commentID}},
		{name: "empty body", currentStatus: ChangeRequestOpen, input: UpdateChangeRequestInput{Body: &emptyBody}, wantErr: true},
		{name: "edit in progress", currentStatus: ChangeRequestInProgress, input: UpdateChangeRequestInput{Body: &body}, wantErr: true},
		{name: "invalid status", currentStatus: ChangeRequestOpen, input: UpdateChangeRequestInput{Status: &invalidStatus}, wantErr: true},
		{name: "open directly resolved", currentStatus: ChangeRequestOpen, input: UpdateChangeRequestInput{Status: &resolved}, wantErr: true},
		{name: "resolved immutable", currentStatus: ChangeRequestResolved, input: UpdateChangeRequestInput{Body: &body}, wantErr: true},
		{name: "canceled immutable", currentStatus: ChangeRequestCanceled, input: UpdateChangeRequestInput{Body: &body}, wantErr: true},
		{name: "run id without resolved", currentStatus: ChangeRequestInProgress, input: UpdateChangeRequestInput{ResolvedByRunID: &runID}, wantErr: true},
		{name: "comment id without resolved", currentStatus: ChangeRequestInProgress, input: UpdateChangeRequestInput{ResultCommentID: &commentID}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NormalizeUpdateChangeRequest(tt.currentStatus, tt.input)
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
