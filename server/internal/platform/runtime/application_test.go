package runtime_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"donarium/server/internal/platform/config"
	"donarium/server/internal/platform/runtime"
)

type fakeModule struct {
	routeRegistered string
}

func (f *fakeModule) RegisterRoutes(r chi.Router) {
	f.routeRegistered = "/test"
	r.Get("/test", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestApplicationRuntime_RegistersModuleRoutes(t *testing.T) {
	fake := &fakeModule{}

	cfg := config.Config{HTTPPort: "8080"}
	app := runtime.NewApplication(cfg, fake)

	if fake.routeRegistered != "/test" {
		t.Error("expected module to receive route registration")
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestApplicationRuntime_MultipleModules(t *testing.T) {
	mod1 := &fakeModule{}
	mod2 := &fakeModule{}

	cfg := config.Config{HTTPPort: "8080"}
	runtime.NewApplication(cfg, mod1, mod2)

	if mod1.routeRegistered != "/test" {
		t.Error("expected module 1 to be registered")
	}
	if mod2.routeRegistered != "/test" {
		t.Error("expected module 2 to be registered")
	}
}

func TestApplicationRuntime_NoModules(t *testing.T) {
	cfg := config.Config{HTTPPort: "8080"}
	app := runtime.NewApplication(cfg)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unregistered route, got %d", rec.Code)
	}
}

type setupModule struct{}

func (m *setupModule) RegisterRoutes(r chi.Router) {
	r.HandleFunc("/api/setup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(runtime.ErrorResponse{Error: "method not allowed"})
			return
		}
		w.WriteHeader(http.StatusCreated)
	})
	r.HandleFunc("/api/setup/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			_ = json.NewEncoder(w).Encode(runtime.ErrorResponse{Error: "method not allowed"})
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

func TestApplicationRuntime_405OnGetSetup(t *testing.T) {
	cfg := config.Config{HTTPPort: "8080"}
	app := runtime.NewApplication(cfg, &setupModule{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/setup", nil)
	app.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}

	allow := rec.Header().Get("Allow")
	if allow != http.MethodPost {
		t.Errorf("expected Allow: POST, got %q", allow)
	}
	ct := rec.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type: application/json, got %q", ct)
	}

	var errResp runtime.ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error != "method not allowed" {
		t.Errorf("expected 'method not allowed', got %q", errResp.Error)
	}
}

func TestApplicationRuntime_405OnPostStatus(t *testing.T) {
	cfg := config.Config{HTTPPort: "8080"}
	app := runtime.NewApplication(cfg, &setupModule{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/setup/status", nil)
	app.Handler().ServeHTTP(rec, req)

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

	var errResp runtime.ErrorResponse
	_ = json.NewDecoder(rec.Body).Decode(&errResp)
	if errResp.Error != "method not allowed" {
		t.Errorf("expected 'method not allowed', got %q", errResp.Error)
	}
}

func TestApplicationRuntime_RunReturnsNilOnShutdown(t *testing.T) {
	cfg := config.Config{HTTPPort: "0"}
	app := runtime.NewApplication(cfg)

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Run()
	}()

	time.Sleep(50 * time.Millisecond)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	err := <-errCh
	if err != nil {
		t.Errorf("Run should return nil after graceful shutdown, got: %v", err)
	}
}

func TestApplicationRuntime_CloseAfterShutdown(t *testing.T) {
	cfg := config.Config{HTTPPort: "0"}
	app := runtime.NewApplication(cfg)

	errCh := make(chan error, 1)
	go func() {
		errCh <- app.Run()
	}()

	time.Sleep(50 * time.Millisecond)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if err := app.Close(); err != nil {
		t.Errorf("Close after Shutdown should be safe, got: %v", err)
	}
}

func TestApplicationRuntime_ImplementsLifecycle(t *testing.T) {
	cfg := config.Config{HTTPPort: "8080"}
	app := runtime.NewApplication(cfg)

	var _ runtime.Runner = app
	var _ runtime.Shutdowner = app
	var _ io.Closer = app
}
