package tracker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/version-1/tasq/internal/issue/domain/entity"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  http.DefaultClient,
	}
}

func (c *Client) Issue(ctx context.Context, id int64) (entity.Issue, error) {
	var output entity.Issue
	if err := c.request(ctx, http.MethodGet, fmt.Sprintf("/api/v1/issues/%d", id), nil, &output); err != nil {
		return entity.Issue{}, err
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

type IssueState = entity.IssueState

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
	if output != nil {
		var payload struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			return err
		}
		if err := json.Unmarshal(payload.Data, output); err != nil {
			return err
		}
	}
	return nil
}
