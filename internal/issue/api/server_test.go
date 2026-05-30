package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/version-1/tasq/internal/issue/domain/entity"
	"github.com/version-1/tasq/internal/issue/store"
)

func TestSuccessResponseEnvelope(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	var payload struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
		Meta map[string]any `json:"meta"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Data.Status != "ok" {
		t.Fatalf("data.status = %q", payload.Data.Status)
	}
	if payload.Meta == nil || len(payload.Meta) != 0 {
		t.Fatalf("meta = %#v", payload.Meta)
	}
}

func TestErrorResponseEnvelope(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/invalid", nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rec.Code)
	}
	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		Meta map[string]any `json:"meta"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Error.Code != "projects.get.invalid_id" {
		t.Fatalf("error.code = %q", payload.Error.Code)
	}
	if payload.Error.Message != "project id is invalid" {
		t.Fatalf("error.message = %q", payload.Error.Message)
	}
	if payload.Meta == nil || len(payload.Meta) != 0 {
		t.Fatalf("meta = %#v", payload.Meta)
	}
}

func TestNoContentResponseStaysEmpty(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	project, err := server.store.CreateProject(context.Background(), entity.CreateProjectInput{
		Key:  "docs",
		Name: "Docs",
	})
	if err != nil {
		t.Fatalf("create project: %v", err)
	}
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/"+stringID(project.ID), nil)
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestOptionsResponseStaysEmpty(t *testing.T) {
	t.Parallel()

	server := newTestServer(t)
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/projects", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec := httptest.NewRecorder()

	server.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d", rec.Code)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func newTestServer(t *testing.T) *Server {
	t.Helper()

	store, err := store.Open(context.Background(), filepath.Join(t.TempDir(), "issue-tracker.sqlite"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatalf("close store: %v", err)
		}
	})
	return NewServer(store)
}

func stringID(id int64) string {
	return strconv.FormatInt(id, 10)
}
