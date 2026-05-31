package runstore

import (
	"strings"
	"testing"
	"time"

	"github.com/version-1/tasq/internal/orchestrator/run"
)

func TestValidateCreateRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   CreateRunInput
		wantErr bool
	}{
		{name: "valid", input: CreateRunInput{IssueID: 1, Workspace: "/tmp/workspace", Attempt: 0, OrchestratorID: "orchestrator"}},
		{name: "missing issue", input: CreateRunInput{Workspace: "/tmp/workspace", OrchestratorID: "orchestrator"}, wantErr: true},
		{name: "workspace too long", input: CreateRunInput{IssueID: 1, Workspace: strings.Repeat("x", 1001), OrchestratorID: "orchestrator"}, wantErr: true},
		{name: "negative attempt", input: CreateRunInput{IssueID: 1, Attempt: -1, OrchestratorID: "orchestrator"}, wantErr: true},
		{name: "empty orchestrator", input: CreateRunInput{IssueID: 1}, wantErr: true},
		{name: "orchestrator too long", input: CreateRunInput{IssueID: 1, OrchestratorID: strings.Repeat("x", 201)}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCreateRun(tt.input)
			assertValidationErr(t, err, tt.wantErr)
		})
	}
}

func TestValidateRunStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		status  run.Status
		wantErr bool
	}{
		{name: "valid", status: run.StatusQueued},
		{name: "invalid", status: run.Status("unknown"), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRunStatus(tt.status)
			assertValidationErr(t, err, tt.wantErr)
		})
	}
}

func TestValidateRunnerEvent(t *testing.T) {
	t.Parallel()

	occurredAt := time.Now().UTC()
	tests := []struct {
		name        string
		runID       string
		eventType   string
		message     string
		payloadJSON string
		occurredAt  time.Time
		wantErr     bool
	}{
		{name: "valid", runID: "run-1", eventType: "event", message: "done", payloadJSON: `{"ok":true}`, occurredAt: occurredAt},
		{name: "empty run id", eventType: "event", occurredAt: occurredAt, wantErr: true},
		{name: "run id too long", runID: strings.Repeat("x", 201), eventType: "event", occurredAt: occurredAt, wantErr: true},
		{name: "event type too long", runID: "run-1", eventType: strings.Repeat("x", 201), occurredAt: occurredAt, wantErr: true},
		{name: "message too long", runID: "run-1", eventType: "event", message: strings.Repeat("x", 10001), occurredAt: occurredAt, wantErr: true},
		{name: "payload too long", runID: "run-1", eventType: "event", payloadJSON: strings.Repeat("x", 50001), occurredAt: occurredAt, wantErr: true},
		{name: "invalid json", runID: "run-1", eventType: "event", payloadJSON: `{`, occurredAt: occurredAt, wantErr: true},
		{name: "zero occurred at", runID: "run-1", eventType: "event", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRunnerEvent(tt.runID, tt.eventType, tt.message, tt.payloadJSON, tt.occurredAt)
			assertValidationErr(t, err, tt.wantErr)
		})
	}
}

func TestValidateWorkspaceMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   WorkspaceMetadataInput
		wantErr bool
	}{
		{name: "valid", input: WorkspaceMetadataInput{WorkspaceKey: "ISSUE-1", IssueID: 1, Path: "/tmp/workspace", SourcePath: "/tmp/repo"}},
		{name: "empty key", input: WorkspaceMetadataInput{IssueID: 1, Path: "/tmp/workspace"}, wantErr: true},
		{name: "key too long", input: WorkspaceMetadataInput{WorkspaceKey: strings.Repeat("x", 201), IssueID: 1, Path: "/tmp/workspace"}, wantErr: true},
		{name: "missing issue", input: WorkspaceMetadataInput{WorkspaceKey: "ISSUE-1", Path: "/tmp/workspace"}, wantErr: true},
		{name: "relative path", input: WorkspaceMetadataInput{WorkspaceKey: "ISSUE-1", IssueID: 1, Path: "workspace"}, wantErr: true},
		{name: "path too long", input: WorkspaceMetadataInput{WorkspaceKey: "ISSUE-1", IssueID: 1, Path: "/" + strings.Repeat("x", 1000)}, wantErr: true},
		{name: "relative source path", input: WorkspaceMetadataInput{WorkspaceKey: "ISSUE-1", IssueID: 1, Path: "/tmp/workspace", SourcePath: "repo"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWorkspaceMetadata(tt.input)
			assertValidationErr(t, err, tt.wantErr)
		})
	}
}

func TestValidateWorkspaceSetupFailure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		issueID      int64
		workspaceKey string
		path         string
		errText      string
		wantErr      bool
	}{
		{name: "valid", issueID: 1, workspaceKey: "ISSUE-1", path: "/tmp/workspace", errText: "failed"},
		{name: "missing issue", workspaceKey: "ISSUE-1", path: "/tmp/workspace", errText: "failed", wantErr: true},
		{name: "key too long", issueID: 1, workspaceKey: strings.Repeat("x", 201), path: "/tmp/workspace", errText: "failed", wantErr: true},
		{name: "path too long", issueID: 1, workspaceKey: "ISSUE-1", path: strings.Repeat("x", 1001), errText: "failed", wantErr: true},
		{name: "relative path", issueID: 1, workspaceKey: "ISSUE-1", path: "workspace", errText: "failed", wantErr: true},
		{name: "empty error", issueID: 1, workspaceKey: "ISSUE-1", path: "/tmp/workspace", wantErr: true},
		{name: "error too long", issueID: 1, workspaceKey: "ISSUE-1", path: "/tmp/workspace", errText: strings.Repeat("x", 10001), wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWorkspaceSetupFailure(tt.issueID, tt.workspaceKey, tt.path, tt.errText)
			assertValidationErr(t, err, tt.wantErr)
		})
	}
}

func assertValidationErr(t *testing.T, err error, wantErr bool) {
	t.Helper()
	if wantErr && err == nil {
		t.Fatal("expected error")
	}
	if !wantErr && err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
