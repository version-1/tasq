package tracker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

type Client struct {
	baseURL string
	client  *http.Client
}

type upsertWorkflowRequest struct {
	Frontmatter json.RawMessage `json:"frontmatter"`
	Body        string          `json:"body"`
	Checksum    string          `json:"checksum"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  http.DefaultClient,
	}
}

func (c *Client) Projects(ctx context.Context) ([]entity.Project, error) {
	var output []entity.Project
	if err := c.request(ctx, http.MethodGet, "/api/v1/projects", nil, &output); err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Client) Issue(ctx context.Context, id int64) (entity.Issue, error) {
	var output entity.Issue
	if err := c.request(ctx, http.MethodGet, fmt.Sprintf("/api/v1/issues/%d", id), nil, &output); err != nil {
		return entity.Issue{}, err
	}
	return output, nil
}

func (c *Client) Project(ctx context.Context, id int64) (entity.Project, error) {
	var output entity.Project
	if err := c.request(ctx, http.MethodGet, fmt.Sprintf("/api/v1/projects/%d", id), nil, &output); err != nil {
		return entity.Project{}, err
	}
	return output, nil
}

func (c *Client) Workflow(ctx context.Context, projectID int64) (entity.ProjectWorkflow, bool, error) {
	var output entity.ProjectWorkflow
	found, err := c.requestOptional(ctx, http.MethodGet, fmt.Sprintf("/api/v1/projects/%d/workflow", projectID), nil, &output)
	if err != nil {
		return entity.ProjectWorkflow{}, false, err
	}
	if !found {
		return entity.ProjectWorkflow{}, false, nil
	}
	return output, true, nil
}

func (c *Client) UpsertWorkflow(ctx context.Context, projectID int64, input entity.UpsertProjectWorkflowInput) (entity.ProjectWorkflow, error) {
	var output entity.ProjectWorkflow
	request := upsertWorkflowRequest{
		Frontmatter: json.RawMessage(input.FrontmatterJSON),
		Body:        input.Body,
		Checksum:    input.Checksum,
	}
	if err := c.request(ctx, http.MethodPut, fmt.Sprintf("/api/v1/projects/%d/workflow", projectID), request, &output); err != nil {
		return entity.ProjectWorkflow{}, err
	}
	return output, nil
}

func (c *Client) UpdateIssue(ctx context.Context, id int64, input entity.UpdateIssueInput) (entity.Issue, error) {
	var output entity.Issue
	if err := c.request(ctx, http.MethodPatch, fmt.Sprintf("/api/v1/issues/%d", id), input, &output); err != nil {
		return entity.Issue{}, err
	}
	return output, nil
}

func (c *Client) CreateComment(ctx context.Context, issueID int64, input entity.CreateCommentInput) (entity.Comment, error) {
	var output entity.Comment
	if err := c.request(ctx, http.MethodPost, fmt.Sprintf("/api/v1/issues/%d/comments", issueID), input, &output); err != nil {
		return entity.Comment{}, err
	}
	return output, nil
}

func (c *Client) Issues(ctx context.Context) ([]entity.Issue, error) {
	var output []entity.Issue
	if err := c.request(ctx, http.MethodGet, "/api/v1/issues", nil, &output); err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Client) IssuesByStates(ctx context.Context, states []string) ([]entity.Issue, error) {
	if len(states) == 0 {
		return []entity.Issue{}, nil
	}
	query := url.Values{}
	query.Set("states", strings.Join(states, ","))

	var output []entity.Issue
	if err := c.request(ctx, http.MethodGet, "/api/v1/issues?"+query.Encode(), nil, &output); err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Client) Queue(ctx context.Context) (entity.Queue, error) {
	var output entity.Queue
	if err := c.request(ctx, http.MethodGet, "/api/v1/queue", nil, &output); err != nil {
		return entity.Queue{}, err
	}
	return output, nil
}

type IssueState struct {
	ID     int64         `json:"id"`
	Status entity.Status `json:"status"`
}

func (c *Client) IssueStatesByIDs(ctx context.Context, ids []int64) ([]IssueState, error) {
	if len(ids) == 0 {
		return []IssueState{}, nil
	}
	var output []IssueState
	if err := c.request(ctx, http.MethodPost, "/api/v1/issues/states", map[string][]int64{
		"ids": ids,
	}, &output); err != nil {
		return nil, err
	}
	return output, nil
}

func (c *Client) request(ctx context.Context, method string, path string, input any, output any) error {
	resp, err := c.do(ctx, method, path, input)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return responseError(method, path, resp)
	}
	return decodeData(resp.Body, output)
}

func (c *Client) requestOptional(ctx context.Context, method string, path string, input any, output any) (bool, error) {
	resp, err := c.do(ctx, method, path, input)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, responseError(method, path, resp)
	}
	if err := decodeData(resp.Body, output); err != nil {
		return false, err
	}
	return true, nil
}

func (c *Client) do(ctx context.Context, method string, path string, input any) (*http.Response, error) {
	var body *bytes.Reader
	if input == nil {
		body = bytes.NewReader(nil)
	} else {
		raw, err := json.Marshal(input)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.client.Do(req)
}

func responseError(method string, path string, resp *http.Response) error {
	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&payload)
	if payload.Error.Message != "" {
		return fmt.Errorf("%s %s returned %s: %s: %s", method, path, resp.Status, payload.Error.Code, payload.Error.Message)
	}
	return fmt.Errorf("%s %s returned %s", method, path, resp.Status)
}

func decodeData(body io.Reader, output any) error {
	if output != nil {
		var payload struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.NewDecoder(body).Decode(&payload); err != nil {
			return err
		}
		if err := json.Unmarshal(payload.Data, output); err != nil {
			return err
		}
	}
	return nil
}
