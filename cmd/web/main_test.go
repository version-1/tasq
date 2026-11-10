package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/version-1/tasq/internal/config"
)

func TestHealth(t *testing.T) {
	handler, err := newMux("http://127.0.0.1:1", "http://127.0.0.1:2")
	if err != nil {
		t.Fatalf("new mux: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status=%d", response.Code)
	}
}

func TestOrchestratorProxyRejectsStateForAnotherTracker(t *testing.T) {
	t.Setenv(config.EnvHome, t.TempDir())
	requests := 0
	orchestrator := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		_, _ = io.WriteString(w, `{"status":"ok"}`)
	}))
	defer orchestrator.Close()
	if err := config.UpdateState(func(state *config.State) error {
		state.IssueTracker = &config.ServiceState{Addr: "127.0.0.1:47651"}
		state.Orchestrator = &config.ServiceState{Addr: strings.TrimPrefix(orchestrator.URL, "http://")}
		return nil
	}); err != nil {
		t.Fatalf("update state: %v", err)
	}
	handler, err := newMux("http://127.0.0.1:37651", "")
	if err != nil {
		t.Fatalf("new mux: %v", err)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/orchestrator/health", nil))
	if response.Code != http.StatusServiceUnavailable || requests != 0 {
		t.Fatalf("status=%d requests=%d, want 503 and 0", response.Code, requests)
	}
}

func TestIndexIncludesBuildMetadata(t *testing.T) {
	handler, err := newMux("http://127.0.0.1:1", "http://127.0.0.1:2")
	if err != nil {
		t.Fatalf("new mux: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	body := response.Body.String()
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d", response.Code)
	}
	if strings.Contains(body, "__TASQ_") {
		t.Fatalf("body contains unresolved build metadata placeholder")
	}
	if !strings.Contains(body, `meta name="tasq-version" content="dev"`) {
		t.Fatalf("body does not contain development version metadata")
	}
	if !strings.Contains(body, `meta name="tasq-commit" content="`) {
		t.Fatalf("body does not contain commit metadata")
	}
}
