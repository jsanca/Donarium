package health_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"donarium/server/internal/platform/http/health"
)

type fakeChecker struct {
	err error
}

func (f fakeChecker) Check(_ context.Context) error {
	return f.err
}

func TestLivenessHandler_Returns200(t *testing.T) {
	handler := health.LivenessHandler()
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp health.LivenessResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", resp.Status)
	}
}

func TestLivenessHandler_ReturnsJSONContentType(t *testing.T) {
	handler := health.LivenessHandler()
	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got '%s'", ct)
	}
}

func TestLivenessHandler_RejectsPOST(t *testing.T) {
	handler := health.LivenessHandler()
	req := httptest.NewRequest(http.MethodPost, "/health/live", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	allow := rec.Header().Get("Allow")
	if allow != http.MethodGet {
		t.Errorf("expected Allow: GET, got %q", allow)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type: application/json, got %q", ct)
	}
	var errResp health.ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error != "method not allowed" {
		t.Errorf("expected 'method not allowed', got %q", errResp.Error)
	}
}

func TestReadinessHandler_ReadyWhenHealthy(t *testing.T) {
	checker := fakeChecker{err: nil}
	handler := health.ReadinessHandler(checker)
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp health.ReadinessResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "ready" {
		t.Errorf("expected status 'ready', got '%s'", resp.Status)
	}
	if resp.Checks.Database != "up" {
		t.Errorf("expected database 'up', got '%s'", resp.Checks.Database)
	}
}

func TestReadinessHandler_NotReadyWhenUnhealthy(t *testing.T) {
	checker := fakeChecker{err: context.DeadlineExceeded}
	handler := health.ReadinessHandler(checker)
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}

	var resp health.ReadinessResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Status != "not_ready" {
		t.Errorf("expected status 'not_ready', got '%s'", resp.Status)
	}
	if resp.Checks.Database != "down" {
		t.Errorf("expected database 'down', got '%s'", resp.Checks.Database)
	}
}

func TestReadinessHandler_ReturnsJSONContentType(t *testing.T) {
	checker := fakeChecker{err: nil}
	handler := health.ReadinessHandler(checker)
	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got '%s'", ct)
	}
}

func TestReadinessHandler_RejectsPOST(t *testing.T) {
	checker := fakeChecker{err: nil}
	handler := health.ReadinessHandler(checker)
	req := httptest.NewRequest(http.MethodPost, "/health/ready", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	allow := rec.Header().Get("Allow")
	if allow != http.MethodGet {
		t.Errorf("expected Allow: GET, got %q", allow)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type: application/json, got %q", ct)
	}
	var errResp health.ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error != "method not allowed" {
		t.Errorf("expected 'method not allowed', got %q", errResp.Error)
	}
}
