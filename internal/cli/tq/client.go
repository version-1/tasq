package tq

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
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

type attachmentUploadInput struct {
	EntityType string
	EntityID   string
	Path       string
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

func (c *apiClient) listIssues(ctx context.Context, projectID *int64) ([]entity.Issue, error) {
	var issues []entity.Issue
	path := "/api/v1/issues"
	if projectID != nil {
		path += "?project_id=" + url.QueryEscape(strconv.FormatInt(*projectID, 10))
	}
	if err := c.do(ctx, http.MethodGet, path, nil, &issues); err != nil {
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

func (c *apiClient) projectByKey(ctx context.Context, key string) (entity.Project, error) {
	projects, err := c.listProjects(ctx)
	if err != nil {
		return entity.Project{}, err
	}
	for _, project := range projects {
		if project.Key == key {
			return project, nil
		}
	}
	return entity.Project{}, fmt.Errorf("project %q not found", key)
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

func (c *apiClient) updateComment(ctx context.Context, id int64, input entity.UpdateCommentInput) (entity.Comment, error) {
	var comment entity.Comment
	if err := c.do(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/comments/%d", id), input, &comment); err != nil {
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

func (c *apiClient) uploadAttachment(ctx context.Context, input attachmentUploadInput) (entity.Attachment, error) {
	file, err := os.Open(input.Path)
	if err != nil {
		return entity.Attachment{}, err
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for name, value := range map[string]string{
		"entity_type": input.EntityType,
		"entity_id":   input.EntityID,
	} {
		field, err := writer.CreateFormField(name)
		if err != nil {
			return entity.Attachment{}, err
		}
		if _, err := io.WriteString(field, value); err != nil {
			return entity.Attachment{}, err
		}
	}
	part, err := writer.CreateFormFile("file", filepath.Base(input.Path))
	if err != nil {
		return entity.Attachment{}, err
	}
	if _, err := io.Copy(part, file); err != nil {
		return entity.Attachment{}, err
	}
	if err := writer.Close(); err != nil {
		return entity.Attachment{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/attachments", &body)
	if err != nil {
		return entity.Attachment{}, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return entity.Attachment{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return entity.Attachment{}, readAPIError(resp)
	}
	var payload apiResponse[json.RawMessage]
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return entity.Attachment{}, fmt.Errorf("decode response: %w", err)
	}
	var attachment entity.Attachment
	if err := json.Unmarshal(payload.Data, &attachment); err != nil {
		return entity.Attachment{}, fmt.Errorf("decode response: %w", err)
	}
	return attachment, nil
}

func (c *apiClient) deleteAttachment(ctx context.Context, id string) error {
	return c.doNoContent(ctx, http.MethodDelete, fmt.Sprintf("/api/v1/attachments/%s", id))
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
