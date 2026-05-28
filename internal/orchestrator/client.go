package orchestrator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	issue "github.com/version-1/tasq/internal/issue/domain/entity"
	orchestratorentity "github.com/version-1/tasq/internal/orchestrator/domain/entity"
)

type IssueTrackerClient struct {
	baseURL string
	client  *http.Client
}

func NewIssueTrackerClient(baseURL string) *IssueTrackerClient {
	return &IssueTrackerClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  http.DefaultClient,
	}
}

func (c *IssueTrackerClient) ClaimWorkItem(ctx context.Context, orchestratorID string, leaseSeconds int) (*issue.WorkItem, error) {
	var output issue.ClaimWorkItemOutput
	err := c.request(ctx, http.MethodPost, "/api/v1/work-items/claim", issue.ClaimWorkItemInput{
		OrchestratorID: orchestratorID,
		LeaseSeconds:   leaseSeconds,
	}, &output)
	if err != nil {
		return nil, err
	}
	return output.WorkItem, nil
}

func (c *IssueTrackerClient) SendRunEvent(ctx context.Context, event orchestratorentity.OutboxEvent) error {
	return c.request(ctx, http.MethodPost, "/api/v1/orchestrator-events", issue.RunEventInput{
		EventID:        event.EventID,
		WorkItemID:     event.WorkItemID,
		IssueID:        event.IssueID,
		RunID:          event.RunID,
		ClaimToken:     event.ClaimToken,
		Status:         issue.RunStatus(event.Status),
		Workspace:      event.Workspace,
		Attempt:        event.Attempt,
		Error:          event.Error,
		OccurredAt:     event.OccurredAt,
		OrchestratorID: event.OrchestratorID,
	}, nil)
}

func (c *IssueTrackerClient) request(ctx context.Context, method string, path string, input any, output any) error {
	var body *bytes.Reader
	if input == nil {
		body = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var payload struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&payload)
		if payload.Error != "" {
			return fmt.Errorf("%s %s returned %s: %s", method, path, resp.Status, payload.Error)
		}
		return fmt.Errorf("%s %s returned %s", method, path, resp.Status)
	}
	if output != nil {
		if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
			return err
		}
	}
	return nil
}
