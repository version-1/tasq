package tq

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type apiStatusError struct{ code int }

func (e apiStatusError) Error() string { return "HTTP request failed" }

type repeatedValue []string

func (v *repeatedValue) String() string { return strings.Join(*v, ",") }
func (v *repeatedValue) Set(value string) error {
	*v = append(*v, value)
	return nil
}

type apiRequest struct {
	method  string
	path    string
	headers http.Header
	body    io.Reader
}

func (a app) api(ctx context.Context, args []string) error {
	req, err := a.parseAPIRequest(args)
	if err != nil {
		return err
	}
	response, err := a.client.doRaw(ctx, req)
	if file, ok := req.body.(*os.File); ok {
		_ = file.Close()
	}
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if _, err := io.Copy(a.stdout, response.Body); err != nil {
		return fmt.Errorf("write response: %w", err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return apiStatusError{code: 1}
	}
	return nil
}

func (a app) parseAPIRequest(args []string) (apiRequest, error) {
	if len(args) < 2 {
		return apiRequest{}, usageError("usage: tq api <method> <path> [--query key=value] [--header 'Name: value'] [--data value|@file|-]")
	}
	method := strings.ToUpper(args[0])
	path, err := validateAPIPath(args[1])
	if err != nil {
		return apiRequest{}, err
	}
	var queries, headers repeatedValue
	fs := newFlagSet("api")
	fs.Var(&queries, "query", "append query parameter")
	fs.Var(&headers, "header", "set HTTP header")
	data := fs.String("data", "", "request body")
	if err := fs.Parse(args[2:]); err != nil {
		return apiRequest{}, usageError("%s", err)
	}
	if len(fs.Args()) != 0 {
		return apiRequest{}, usageError("usage: tq api <method> <path> [--query key=value] [--header 'Name: value'] [--data value|@file|-]")
	}
	if !allowedAPIRoute(method, path) {
		return apiRequest{}, usageError("method and path are not allowed: %s %s", method, path)
	}
	path, err = appendAPIQuery(path, queries)
	if err != nil {
		return apiRequest{}, err
	}

	requestHeaders, err := parseAPIHeaders(headers)
	if err != nil {
		return apiRequest{}, err
	}
	var body io.Reader
	dataProvided := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == "data" {
			dataProvided = true
		}
	})
	if dataProvided {
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch {
			return apiRequest{}, usageError("--data is only allowed with POST, PUT, or PATCH")
		}
		body, err = apiBody(*data, a.stdin)
		if err != nil {
			return apiRequest{}, err
		}
		if requestHeaders.Get("Content-Type") == "" {
			requestHeaders.Set("Content-Type", "application/json")
		}
	}
	return apiRequest{method: method, path: path, headers: requestHeaders, body: body}, nil
}

func apiBody(value string, stdin io.Reader) (io.Reader, error) {
	if value == "-" {
		return stdin, nil
	}
	if strings.HasPrefix(value, "@") {
		file, err := os.Open(strings.TrimPrefix(value, "@"))
		if err != nil {
			return nil, usageError("read data file: %v", err)
		}
		return file, nil
	}
	return strings.NewReader(value), nil
}

func appendAPIQuery(path string, queries []string) (string, error) {
	if len(queries) == 0 {
		return path, nil
	}
	parts := make([]string, 0, len(queries))
	for _, query := range queries {
		key, value, ok := strings.Cut(query, "=")
		if !ok {
			return "", usageError("query must be key=value")
		}
		parts = append(parts, url.QueryEscape(key)+"="+url.QueryEscape(value))
	}
	separator := "?"
	if strings.Contains(path, "?") {
		separator = "&"
	}
	return path + separator + strings.Join(parts, "&"), nil
}

func parseAPIHeaders(values []string) (http.Header, error) {
	headers := make(http.Header)
	for _, value := range values {
		name, headerValue, ok := strings.Cut(value, ":")
		name = strings.TrimSpace(name)
		if !ok || name == "" {
			return nil, usageError("header must be 'Name: value'")
		}
		if !isValidHeaderName(name) {
			return nil, usageError("header name %q is invalid", name)
		}
		if strings.ContainsAny(headerValue, "\r\n") {
			return nil, usageError("header %q must not contain a newline", name)
		}
		if isTransportHeader(name) {
			return nil, usageError("header %q is managed by the HTTP transport", name)
		}
		headers.Set(name, strings.TrimSpace(headerValue))
	}
	return headers, nil
}

func isValidHeaderName(value string) bool {
	for _, character := range value {
		if !((character >= 'a' && character <= 'z') ||
			(character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') ||
			strings.ContainsRune("!#$%&'*+-.^_`|~", character)) {
			return false
		}
	}
	return value != ""
}

func isTransportHeader(name string) bool {
	switch strings.ToLower(name) {
	case "host", "content-length", "transfer-encoding", "connection", "trailer", "upgrade", "proxy-connection", "keep-alive", "te":
		return true
	default:
		return false
	}
}

func validateAPIPath(value string) (string, error) {
	path, _, _ := strings.Cut(value, "?")
	if !strings.HasPrefix(value, "/") || strings.Contains(value, "#") || strings.Contains(path, "%") {
		return "", usageError("path must be an unencoded absolute API path without a fragment")
	}
	if path == "/" || strings.HasSuffix(path, "/") || strings.Contains(path, "//") {
		return "", usageError("path must not contain empty segments or a trailing slash")
	}
	for _, segment := range strings.Split(strings.TrimPrefix(path, "/"), "/") {
		if segment == "." || segment == ".." {
			return "", usageError("path must not contain dot segments")
		}
	}
	return value, nil
}

func allowedAPIRoute(method, requestPath string) bool {
	path, _, _ := strings.Cut(requestPath, "?")
	for _, route := range apiRoutes {
		if route.method == method && route.matches(path) {
			return true
		}
	}
	return false
}

type apiRoute struct{ method, template string }

func (r apiRoute) matches(path string) bool {
	want, got := strings.Split(strings.TrimPrefix(r.template, "/"), "/"), strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(want) != len(got) {
		return false
	}
	for i := range want {
		if want[i] == "{id}" || want[i] == "{issueId}" {
			if !isPositiveInt64(got[i]) {
				return false
			}
			continue
		}
		if want[i] == "{attachmentId}" {
			if got[i] == "" {
				return false
			}
			continue
		}
		if want[i] == "{type}" {
			if got[i] != "pull_request" {
				return false
			}
			continue
		}
		if want[i] != got[i] {
			return false
		}
	}
	return true
}

func isPositiveInt64(value string) bool {
	if value == "" {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	return err == nil && parsed > 0
}

// apiRoutes is intentionally explicit: new server endpoints remain unavailable until reviewed here.
var apiRoutes = []apiRoute{
	{"GET", "/api/v1/health"}, {"GET", "/api/v1/summary"},
	{"GET", "/api/v1/projects"}, {"POST", "/api/v1/projects"}, {"GET", "/api/v1/projects/{id}"}, {"PATCH", "/api/v1/projects/{id}"}, {"DELETE", "/api/v1/projects/{id}"}, {"POST", "/api/v1/projects/{id}/check"},
	{"GET", "/api/v1/projects/{id}/workflow"}, {"PUT", "/api/v1/projects/{id}/workflow"}, {"DELETE", "/api/v1/projects/{id}/workflow"},
	{"GET", "/api/v1/issues"}, {"POST", "/api/v1/issues"}, {"POST", "/api/v1/issues/states"}, {"GET", "/api/v1/queue"}, {"GET", "/api/v1/issues/{id}"}, {"PATCH", "/api/v1/issues/{id}"},
	{"PUT", "/api/v1/issues/{issueId}/artifacts/{type}"}, {"DELETE", "/api/v1/issues/{issueId}/artifacts/{type}"},
	{"GET", "/api/v1/issues/{issueId}/comments"}, {"POST", "/api/v1/issues/{issueId}/comments"}, {"PATCH", "/api/v1/comments/{id}"},
	{"GET", "/api/v1/issues/{issueId}/change-requests"}, {"POST", "/api/v1/issues/{issueId}/change-requests"}, {"GET", "/api/v1/change-requests/{id}"}, {"PATCH", "/api/v1/change-requests/{id}"}, {"POST", "/api/v1/change-requests/{id}/cancel"},
	{"GET", "/api/v1/attachments"}, {"GET", "/api/v1/attachments/{attachmentId}/content"}, {"DELETE", "/api/v1/attachments/{attachmentId}"},
}

func (c *apiClient) doRaw(ctx context.Context, input apiRequest) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, input.method, c.baseURL+input.path, input.body)
	if err != nil {
		return nil, err
	}
	req.Header = input.headers
	client := c.rawHTTPClient()
	return client.Do(req)
}

func (c *apiClient) rawHTTPClient() *http.Client {
	client := http.Client{Timeout: 10 * time.Second}
	if c.httpClient != nil {
		client = *c.httpClient
	}
	client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &client
}
