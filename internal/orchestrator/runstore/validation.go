package runstore

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"time"
	"unicode/utf8"

	"github.com/version-1/tasq/internal/orchestrator/run"
)

const (
	maxRunIDLength            = 200
	maxRunThreadIDLength      = 200
	maxRunWorkspaceLength     = 1000
	maxRunErrorLength         = 10000
	maxOrchestratorIDLength   = 200
	maxEventTypeLength        = 200
	maxEventMessageLength     = 10000
	maxEventPayloadJSONLength = 50000
	maxWorkspaceKeyLength     = 200
	maxWorkspacePathLength    = 1000
)

func ValidateCreateRun(input CreateRunInput) error {
	if input.IssueID <= 0 {
		return errors.New("issueId is required")
	}
	if runeCount(input.Workspace) > maxRunWorkspaceLength {
		return errors.New("workspace must be 1000 characters or fewer")
	}
	if runeCount(input.ThreadID) > maxRunThreadIDLength {
		return errors.New("threadId must be 200 characters or fewer")
	}
	if input.Attempt < 0 {
		return errors.New("attempt must be greater than or equal to 0")
	}
	if input.OrchestratorID == "" {
		return errors.New("orchestratorId is required")
	}
	if runeCount(input.OrchestratorID) > maxOrchestratorIDLength {
		return errors.New("orchestratorId must be 200 characters or fewer")
	}
	return nil
}

func ValidateRunStatus(status run.Status) error {
	switch status {
	case run.StatusQueued, run.StatusRunning, run.StatusSucceeded, run.StatusFailed, run.StatusCancelled:
		return nil
	default:
		return errors.New("status is invalid")
	}
}

func ValidateRunnerEvent(runID, eventType, message, payloadJSON string, occurredAt time.Time) error {
	if runID == "" {
		return errors.New("runId is required")
	}
	if runeCount(runID) > maxRunIDLength {
		return errors.New("runId must be 200 characters or fewer")
	}
	if eventType == "" {
		eventType = "event"
	}
	if runeCount(eventType) > maxEventTypeLength {
		return errors.New("eventType must be 200 characters or fewer")
	}
	if runeCount(message) > maxEventMessageLength {
		return errors.New("message must be 10000 characters or fewer")
	}
	if runeCount(payloadJSON) > maxEventPayloadJSONLength {
		return errors.New("payloadJSON must be 50000 characters or fewer")
	}
	if payloadJSON != "" && !json.Valid([]byte(payloadJSON)) {
		return errors.New("payloadJSON is invalid")
	}
	if occurredAt.IsZero() {
		return errors.New("occurredAt is required")
	}
	return nil
}

func ValidateWorkspaceMetadata(input WorkspaceMetadataInput) error {
	if input.WorkspaceKey == "" {
		return errors.New("workspaceKey is required")
	}
	if runeCount(input.WorkspaceKey) > maxWorkspaceKeyLength {
		return errors.New("workspaceKey must be 200 characters or fewer")
	}
	if input.IssueID <= 0 {
		return errors.New("issueId is required")
	}
	if err := validateAbsolutePath("path", input.Path); err != nil {
		return err
	}
	if input.SourcePath != "" {
		if err := validateAbsolutePath("sourcePath", input.SourcePath); err != nil {
			return err
		}
	}
	return nil
}

func ValidateWorkspaceSetupFailure(issueID int64, workspaceKey, path, errText string) error {
	if issueID <= 0 {
		return errors.New("issueId is required")
	}
	if runeCount(workspaceKey) > maxWorkspaceKeyLength {
		return errors.New("workspaceKey must be 200 characters or fewer")
	}
	if runeCount(path) > maxWorkspacePathLength {
		return errors.New("path must be 1000 characters or fewer")
	}
	if path != "" && !filepath.IsAbs(path) {
		return errors.New("path must be an absolute path")
	}
	if errText == "" {
		return errors.New("error is required")
	}
	if runeCount(errText) > maxRunErrorLength {
		return errors.New("error must be 10000 characters or fewer")
	}
	return nil
}

func validateAbsolutePath(field string, value string) error {
	if value == "" {
		return errors.New(field + " is required")
	}
	if runeCount(value) > maxWorkspacePathLength {
		return errors.New(field + " must be 1000 characters or fewer")
	}
	if !filepath.IsAbs(value) {
		return errors.New(field + " must be an absolute path")
	}
	return nil
}

func runeCount(value string) int {
	return utf8.RuneCountInString(value)
}
