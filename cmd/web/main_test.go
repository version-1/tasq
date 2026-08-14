package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
