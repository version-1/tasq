package tq

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

type apiClient struct {
	baseURL    string
	httpClient *http.Client
}

type apiResponse[T any] struct {
	Data T `json:"data"`
}

type apiErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func newAPIClient(baseURL string) (*apiClient, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, errors.New("api-url is required")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("api-url is invalid: %w", err)
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("api-url must include scheme and host")
	}
	return &apiClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (c *apiClient) listIssues(ctx context.Context) ([]entity.Issue, error) {
	var issues []entity.Issue
	if err := c.do(ctx, http.MethodGet, "/api/v1/issues", nil, &issues); err != nil {
		return nil, err
	}
	return issues, nil
}

func (c *apiClient) listProjects(ctx context.Context) ([]entity.Project, error) {
	var projects []entity.Project
	if err := c.do(ctx, http.MethodGet, "/api/v1/projects", nil, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (c *apiClient) createProject(ctx context.Context, input entity.CreateProjectInput) (entity.Project, error) {
	var project entity.Project
	if err := c.do(ctx, http.MethodPost, "/api/v1/projects", input, &project); err != nil {
		return entity.Project{}, err
	}
	return project, nil
}

func (c *apiClient) deleteProject(ctx context.Context, id int64) error {
	return c.doNoContent(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/projects/%d", id))
}

func (c *apiClient) createWorkspace(ctx context.Context, input entity.CreateWorkspaceInput) (entity.Workspace, error) {
	var workspace entity.Workspace
	if err := c.do(ctx, http.MethodPost, "/api/v1/workspaces", input, &workspace); err != nil {
		return entity.Workspace{}, err
	}
	return workspace, nil
}

func (c *apiClient) listWorkspaces(ctx context.Context) ([]entity.Workspace, error) {
	var workspaces []entity.Workspace
	if err := c.do(ctx, http.MethodGet, "/api/v1/workspaces", nil, &workspaces); err != nil {
		return nil, err
	}
	return workspaces, nil
}

func (c *apiClient) checkProject(ctx context.Context, id int64, workflow string) (projectCheckResult, error) {
	var result projectCheckResult
	if err := c.doText(ctx, http.MethodPost, fmt.Sprintf("/api/v1/projects/%d/check", id), workflow, &result); err != nil {
		return projectCheckResult{}, err
	}
	return result, nil
}

func (c *apiClient) getIssue(ctx context.Context, id int64) (entity.Issue, error) {
	var issue entity.Issue
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/issues/%d", id), nil, &issue); err != nil {
		return entity.Issue{}, err
	}
	return issue, nil
}

func (c *apiClient) createIssue(ctx context.Context, input entity.CreateIssueInput) (entity.Issue, error) {
	var issue entity.Issue
	if err := c.do(ctx, http.MethodPost, "/api/v1/issues", input, &issue); err != nil {
		return entity.Issue{}, err
	}
	return issue, nil
}

func (c *apiClient) updateIssue(ctx context.Context, id int64, patch map[string]string) (entity.Issue, error) {
	var issue entity.Issue
	if err := c.do(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/issues/%d", id), patch, &issue); err != nil {
		return entity.Issue{}, err
	}
	return issue, nil
}

func (c *apiClient) createComment(ctx context.Context, issueID int64, input entity.CreateCommentInput) (entity.Comment, error) {
	var comment entity.Comment
	if err := c.do(ctx, http.MethodPost, fmt.Sprintf("/api/v1/issues/%d/comments", issueID), input, &comment); err != nil {
		return entity.Comment{}, err
	}
	return comment, nil
}

func (c *apiClient) listComments(ctx context.Context, issueID int64) ([]entity.Comment, error) {
	var comments []entity.Comment
	if err := c.do(ctx, http.MethodGet, fmt.Sprintf("/api/v1/issues/%d/comments", issueID), nil, &comments); err != nil {
		return nil, err
	}
	return comments, nil
}

func (c *apiClient) doNoContent(ctx context.Context, method, path string) error {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, nil)
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
	return nil
}

func (c *apiClient) doText(ctx context.Context, method, path string, body string, output any) error {
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, strings.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "text/plain; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return readAPIError(resp)
	}
	var payload apiResponse[json.RawMessage]
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if err := json.Unmarshal(payload.Data, output); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func (c *apiClient) do(ctx context.Context, method, path string, body any, output any) error {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return readAPIError(resp)
	}
	var payload apiResponse[json.RawMessage]
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if err := json.Unmarshal(payload.Data, output); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func readAPIError(resp *http.Response) error {
	var payload apiErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return cliError{message: fmt.Sprintf("%s returned %s", resp.Request.URL.Path, resp.Status), code: 1}
	}
	message := payload.Error.Message
	if message == "" {
		message = resp.Status
	}
	return cliError{message: message, code: 1}
}
