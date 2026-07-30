package tq

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAllowedAPIRoutes(t *testing.T) {
	tests := []struct{ method, path string }{
		{"GET", "/api/v1/health"}, {"GET", "/api/v1/summary"},
		{"GET", "/api/v1/projects"}, {"POST", "/api/v1/projects"}, {"GET", "/api/v1/projects/1"}, {"PATCH", "/api/v1/projects/1"}, {"DELETE", "/api/v1/projects/1"}, {"POST", "/api/v1/projects/1/check"},
		{"GET", "/api/v1/projects/1/workflow"}, {"PUT", "/api/v1/projects/1/workflow"}, {"DELETE", "/api/v1/projects/1/workflow"},
		{"GET", "/api/v1/issues"}, {"POST", "/api/v1/issues"}, {"POST", "/api/v1/issues/states"}, {"GET", "/api/v1/queue"}, {"GET", "/api/v1/issues/1"}, {"PATCH", "/api/v1/issues/1"},
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
	client, err := newAPIClient(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	client.httpClient.Timeout = 20 * time.Millisecond
	_, err = client.doRaw(context.Background(), apiRequest{method: http.MethodGet, path: "/api/v1/summary", headers: make(http.Header)})
	if err == nil {
		t.Fatal("expected timeout")
	}
}

func runAPI(t *testing.T, apiURL string, stdin io.Reader, args []string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := run(context.Background(), append([]string{"--api-url", apiURL, "api"}, args...), stdin, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}
