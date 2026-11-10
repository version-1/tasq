package tq

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAllowedAPIRoutes(t *testing.T) {
	expected := []apiRoute{
		{"GET", "/api/v1/health"}, {"GET", "/api/v1/summary"},
		{"GET", "/api/v1/projects"}, {"POST", "/api/v1/projects"}, {"GET", "/api/v1/projects/{id}"}, {"PATCH", "/api/v1/projects/{id}"}, {"DELETE", "/api/v1/projects/{id}"}, {"POST", "/api/v1/projects/{id}/check"},
		{"GET", "/api/v1/projects/{id}/workflow"}, {"PUT", "/api/v1/projects/{id}/workflow"}, {"DELETE", "/api/v1/projects/{id}/workflow"},
		{"GET", "/api/v1/issues"}, {"POST", "/api/v1/issues"}, {"POST", "/api/v1/issues/states"}, {"GET", "/api/v1/queue"}, {"GET", "/api/v1/issues/{id}"}, {"PATCH", "/api/v1/issues/{id}"},
		{"PUT", "/api/v1/issues/{issueId}/artifacts/{type}"}, {"DELETE", "/api/v1/issues/{issueId}/artifacts/{type}"},
		{"GET", "/api/v1/issues/{issueId}/comments"}, {"POST", "/api/v1/issues/{issueId}/comments"}, {"PATCH", "/api/v1/comments/{id}"},
		{"GET", "/api/v1/issues/{issueId}/change-requests"}, {"POST", "/api/v1/issues/{issueId}/change-requests"}, {"GET", "/api/v1/change-requests/{id}"}, {"PATCH", "/api/v1/change-requests/{id}"}, {"POST", "/api/v1/change-requests/{id}/cancel"},
		{"GET", "/api/v1/attachments"}, {"GET", "/api/v1/attachments/{attachmentId}/content"}, {"DELETE", "/api/v1/attachments/{attachmentId}"},
	}
	if len(apiRoutes) != len(expected) {
		t.Fatalf("allowlist has %d routes, want %d", len(apiRoutes), len(expected))
	}
	expectedSet := make(map[string]struct{}, len(expected))
	for _, route := range expected {
		expectedSet[apiRouteKey(route)] = struct{}{}
	}
	if len(expectedSet) != len(expected) {
		t.Fatal("expected allowlist contains duplicate routes")
	}
	actualSet := make(map[string]struct{}, len(apiRoutes))
	for _, route := range apiRoutes {
		actualSet[apiRouteKey(route)] = struct{}{}
	}
	if len(actualSet) != len(apiRoutes) {
		t.Fatal("allowlist contains duplicate routes")
	}
	for key := range expectedSet {
		if _, ok := actualSet[key]; !ok {
			t.Fatalf("allowlist is missing %s", key)
		}
	}
	tests := []struct{ method, path string }{
		{"GET", "/api/v1/health"}, {"GET", "/api/v1/summary"},
		{"GET", "/api/v1/projects"}, {"POST", "/api/v1/projects"}, {"GET", "/api/v1/projects/1"}, {"PATCH", "/api/v1/projects/1"}, {"DELETE", "/api/v1/projects/1"}, {"POST", "/api/v1/projects/1/check"},
		{"GET", "/api/v1/projects/1/workflow"}, {"PUT", "/api/v1/projects/1/workflow"}, {"DELETE", "/api/v1/projects/1/workflow"},
		{"GET", "/api/v1/issues"}, {"POST", "/api/v1/issues"}, {"POST", "/api/v1/issues/states"}, {"GET", "/api/v1/queue"}, {"GET", "/api/v1/issues/1"}, {"PATCH", "/api/v1/issues/1"},
		{"PUT", "/api/v1/issues/1/artifacts/pull_request"}, {"DELETE", "/api/v1/issues/1/artifacts/pull_request"},
		{"GET", "/api/v1/issues/1/comments"}, {"POST", "/api/v1/issues/1/comments"}, {"PATCH", "/api/v1/comments/1"},
		{"GET", "/api/v1/issues/1/change-requests"}, {"POST", "/api/v1/issues/1/change-requests"}, {"GET", "/api/v1/change-requests/1"}, {"PATCH", "/api/v1/change-requests/1"}, {"POST", "/api/v1/change-requests/1/cancel"},
		{"GET", "/api/v1/attachments"}, {"GET", "/api/v1/attachments/att_1/content"}, {"DELETE", "/api/v1/attachments/att_1"},
	}
	for _, test := range tests {
		t.Run(test.method+" "+test.path, func(t *testing.T) {
			if !allowedAPIRoute(test.method, test.path) {
				t.Fatal("route is not allowed")
			}
		})
	}
}

func apiRouteKey(route apiRoute) string { return route.method + " " + route.template }

func TestAPIRejectsInvalidRoutesAndPaths(t *testing.T) {
	for _, args := range [][]string{
		{"POST", "/api/v1/attachments"}, {"PATCH", "/api/v1/attachments/att_1"}, {"HEAD", "/api/v1/issues"}, {"GET", "/api/v1/unknown"}, {"GET", "/api/v1/issues/0"}, {"GET", "/api/v1/issues/-1"}, {"GET", "/api/v1/issues/9223372036854775808"}, {"GET", "/api/v1/issues/abc"},
		{"GET", "https://example.test/api/v1/issues"}, {"GET", "/api/v1/issues/"}, {"GET", "/api/v1/../issues"}, {"GET", "/api//v1/issues"}, {"GET", "/api/v1/%69ssues"}, {"GET", "/api/v1/issues#part"},
	} {
		_, err := (app{}).parseAPIRequest(args)
		if err == nil {
			t.Fatalf("args %v: expected error", args)
		}
	}
	if _, err := (app{}).parseAPIRequest([]string{"GET", "/api/v1/issues?search=100%25"}); err != nil {
		t.Fatalf("encoded query must remain permitted: %v", err)
	}
	for _, args := range [][]string{
		{"GET", "/api/v1/issues", "--query", "missing-value"},
		{"GET", "/api/v1/issues", "--header", "missing-colon"},
		{"GET", "/api/v1/issues", "--header", "Bad Name: value"},
		{"GET", "/api/v1/issues", "--header", "X-Test: first\nsecond"},
	} {
		if _, err := (app{}).parseAPIRequest(args); err == nil {
			t.Fatalf("args %v: expected error", args)
		}
	}
}

func TestAPIForwardsRequestAndResponseBytes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/attachments" {
			_, _ = w.Write([]byte{0x00, 0xff, '\n'})
			return
		}
		if r.Method != http.MethodPatch || r.URL.Path != "/api/v1/issues/42" {
			t.Fatalf("request=%s %s", r.Method, r.URL.Path)
		}
		if got, want := r.URL.Query()["state"], []string{"ready", "again"}; strings.Join(got, ",") != strings.Join(want, ",") {
			t.Fatalf("query=%v", got)
		}
		if r.Header.Get("X-Test") != "second" || r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("headers=%v", r.Header)
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"status":"done"}` {
			t.Fatalf("body=%q", body)
		}
		_, _ = w.Write([]byte{0x00, 0xff, '\n'})
	}))
	defer server.Close()
	stdout, stderr, code := runAPI(t, server.URL, strings.NewReader(""), []string{"patch", "/api/v1/issues/42?state=ready", "--query", "state=again", "--header", "X-Test: first", "--header", "x-test: second", "--data", `{"status":"done"}`})
	if code != 0 || stderr != "" || stdout != string([]byte{0x00, 0xff, '\n'}) {
		t.Fatalf("code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	stdout, stderr, code = runAPI(t, server.URL, strings.NewReader(""), []string{"--output", "json", "GET", "/api/v1/attachments"})
	if code != 0 || stderr != "" || stdout != string([]byte{0x00, 0xff, '\n'}) {
		t.Fatalf("json output must stay raw: code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestAPIReadsDataFileAndDoesNotValidateLiteralJSON(t *testing.T) {
	file := filepath.Join(t.TempDir(), "request.txt")
	if err := os.WriteFile(file, []byte("from-file"), 0o600); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if string(body) != "from-file" && string(body) != "not-json" {
			t.Fatalf("body=%q", body)
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()
	for _, data := range []string{"@" + file, "not-json"} {
		stdout, stderr, code := runAPI(t, server.URL, strings.NewReader(""), []string{"POST", "/api/v1/issues", "--data", data})
		if code != 0 || stderr != "" || stdout != "ok" {
			t.Fatalf("data=%q code=%d stdout=%q stderr=%q", data, code, stdout, stderr)
		}
	}
}

func TestAPIStatusRedirectAndInputErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/health" {
			w.Header().Set("Location", "/api/v1/issues")
			w.WriteHeader(http.StatusFound)
			_, _ = w.Write([]byte("redirect"))
			return
		}
		if r.URL.Path == "/api/v1/summary" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("server-error"))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad"))
	}))
	defer server.Close()
	for _, test := range []struct {
		args []string
		want string
		code int
	}{
		{[]string{"GET", "/api/v1/health"}, "redirect", 1}, {[]string{"GET", "/api/v1/issues"}, "bad", 1}, {[]string{"GET", "/api/v1/summary"}, "server-error", 1},
		{[]string{"GET", "/api/v1/issues", "--data", "x"}, "", 2}, {[]string{"GET", "/api/v1/issues", "--header", "Host: example.test"}, "", 2},
	} {
		stdout, stderr, code := runAPI(t, server.URL, strings.NewReader(""), test.args)
		if code != test.code || stdout != test.want || stderr == "" && code == 2 {
			t.Fatalf("args=%v code=%d stdout=%q stderr=%q", test.args, code, stdout, stderr)
		}
		if code == 1 && stderr != "" {
			t.Fatalf("HTTP status must not write stderr: %q", stderr)
		}
	}
}

func TestAPIReadsDataFromStdinAndHonorsClientTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/issues" {
			body, _ := io.ReadAll(r.Body)
			if string(body) != "stdin" {
				t.Fatalf("body=%q", body)
			}
			return
		}
		time.Sleep(time.Second)
	}))
	defer server.Close()
	_, stderr, code := runAPI(t, server.URL, strings.NewReader("stdin"), []string{"POST", "/api/v1/issues", "--data", "-"})
	if code != 0 || stderr != "" {
		t.Fatalf("stdin code=%d stderr=%q", code, stderr)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, stderr, code = runAPIContext(t, ctx, server.URL, strings.NewReader(""), []string{"GET", "/api/v1/summary"})
	if code != 1 || stderr == "" {
		t.Fatalf("timeout code=%d stderr=%q", code, stderr)
	}
}

func TestAPIEmptyResponseAndConnectionFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	stdout, stderr, code := runAPI(t, server.URL, strings.NewReader(""), []string{"GET", "/api/v1/health"})
	if code != 0 || stdout != "" || stderr != "" {
		t.Fatalf("204 code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	stdout, stderr, code = runAPI(t, "http://"+address, strings.NewReader(""), []string{"GET", "/api/v1/health"})
	if code != 1 || stdout != "" || stderr == "" {
		t.Fatalf("connection failure code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func runAPI(t *testing.T, apiURL string, stdin io.Reader, args []string) (string, string, int) {
	return runAPIContext(t, context.Background(), apiURL, stdin, args)
}

func runAPIContext(t *testing.T, ctx context.Context, apiURL string, stdin io.Reader, args []string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run(ctx, append([]string{"--api-url", apiURL, "api"}, args...), stdin, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}
