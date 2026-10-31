package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAutoOKRejectsUnknownPhase(t *testing.T) {
	server := newTestServer(t)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/issues/1/auto-ok-result", strings.NewReader(`{"generation":1,"phase":"cliam"}`))
	rec := httptest.NewRecorder()
	server.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "issues.auto_ok.invalid_request") {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
