package tq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

type issueListQuery struct {
	States     []entity.Status
	ProjectIDs []int64
	Search     string
	Offset     int
}

type issuePage struct {
	Issues     []entity.Issue
	NextOffset *int
}

type commentPage struct {
	Comments   []entity.Comment
	NextCursor *int64
}

type responseEnvelope[T any] struct {
	Data T              `json:"data"`
	Meta map[string]any `json:"meta"`
}

func (c *apiClient) listIssuesPage(ctx context.Context, query issueListQuery) (issuePage, error) {
	values := url.Values{}
	if len(query.States) > 0 {
		states := make([]string, len(query.States))
		for i, status := range query.States {
			states[i] = string(status)
		}
		values.Set("states", strings.Join(states, ","))
	}
	if len(query.ProjectIDs) > 0 {
		ids := make([]string, len(query.ProjectIDs))
		for i, id := range query.ProjectIDs {
			ids[i] = strconv.FormatInt(id, 10)
		}
		values.Set("project_ids", strings.Join(ids, ","))
	}
	if strings.TrimSpace(query.Search) != "" {
		values.Set("search", strings.TrimSpace(query.Search))
	}
	values.Set("limit", "50")
	values.Set("offset", strconv.Itoa(query.Offset))
	values.Set("sort_by", "updated_at")
	values.Set("sort_direction", "desc")

	var envelope responseEnvelope[[]entity.Issue]
	if err := c.getEnvelope(ctx, "/api/v1/issues?"+values.Encode(), &envelope); err != nil {
		return issuePage{}, err
	}
	return issuePage{Issues: envelope.Data, NextOffset: metaInt(envelope.Meta, "nextOffset")}, nil
}

func (c *apiClient) commentsPage(ctx context.Context, issueID int64, cursor int64) (commentPage, error) {
	values := url.Values{}
	values.Set("direction", "backward")
	values.Set("limit", "50")
	if cursor > 0 {
		values.Set("cursor", strconv.FormatInt(cursor, 10))
	}
	var envelope responseEnvelope[[]entity.Comment]
	if err := c.getEnvelope(ctx, fmt.Sprintf("/api/v1/issues/%d/comments?%s", issueID, values.Encode()), &envelope); err != nil {
		return commentPage{}, err
	}
	comments := envelope.Data
	for left, right := 0, len(comments)-1; left < right; left, right = left+1, right-1 {
		comments[left], comments[right] = comments[right], comments[left]
	}
	return commentPage{Comments: comments, NextCursor: metaInt64(envelope.Meta, "nextCursor")}, nil
}

func (c *apiClient) getEnvelope(ctx context.Context, path string, output any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return readAPIError(resp)
	}
	if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func metaInt(meta map[string]any, key string) *int {
	value, ok := meta[key].(float64)
	if !ok {
		return nil
	}
	result := int(value)
	return &result
}

func metaInt64(meta map[string]any, key string) *int64 {
	value, ok := meta[key].(float64)
	if !ok {
		return nil
	}
	result := int64(value)
	return &result
}

type orchestratorClient struct {
	baseURL    string
	httpClient *http.Client
}

type runtimeIssue struct {
	Status       string           `json:"status"`
	Workspace    runtimeWorkspace `json:"workspace"`
	Attempts     runtimeAttempts  `json:"attempts"`
	Runs         []runtimeRun     `json:"runs"`
	RecentEvents []runtimeEvent   `json:"recent_events"`
	LastError    any              `json:"last_error"`
}

type runtimeWorkspace struct {
	Path string `json:"path"`
}
type runtimeAttempts struct {
	RestartCount        int `json:"restart_count"`
	CurrentRetryAttempt int `json:"current_retry_attempt"`
}
type runtimeRun struct {
	RunID     string    `json:"run_id"`
	Status    string    `json:"status"`
	Attempt   int       `json:"attempt"`
	UpdatedAt time.Time `json:"updated_at"`
}
type runtimeEvent struct {
	At      time.Time `json:"at"`
	Event   string    `json:"event"`
	Message string    `json:"message"`
}

var errRuntimeNotFound = errors.New("run data not found")

func newOrchestratorClient(baseURL string) (*orchestratorClient, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, nil
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, errors.New("orchestrator-url must be an absolute http or https URL")
	}
	return &orchestratorClient{baseURL: baseURL, httpClient: &http.Client{Timeout: 5 * time.Second}}, nil
}

func (c *orchestratorClient) issue(ctx context.Context, issueID int64) (runtimeIssue, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/issue-%d", c.baseURL, issueID), nil)
	if err != nil {
		return runtimeIssue{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return runtimeIssue{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return runtimeIssue{}, errRuntimeNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return runtimeIssue{}, fmt.Errorf("orchestrator returned %s", resp.Status)
	}
	var result runtimeIssue
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return runtimeIssue{}, fmt.Errorf("decode orchestrator response: %w", err)
	}
	return result, nil
}
