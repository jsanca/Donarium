package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"donarium/server/internal/platform/config"
)

var (
	_ Runner     = (*ApplicationRuntime)(nil)
	_ Shutdowner = (*ApplicationRuntime)(nil)
	_ io.Closer  = (*ApplicationRuntime)(nil)
)

type ApplicationRuntime struct {
	cfg     config.Config
	router  chi.Router
	handler http.Handler
	server  *http.Server
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewApplication(cfg config.Config, modules ...ModuleRuntime) *ApplicationRuntime {
	r := chi.NewRouter()
	r.MethodNotAllowed(methodNotAllowedHandler)

	for _, m := range modules {
		m.RegisterRoutes(r)
	}

	return &ApplicationRuntime{
		cfg:     cfg,
		router:  r,
		handler: r,
	}
}

func (a *ApplicationRuntime) Handler() http.Handler {
	return a.handler
}

func (a *ApplicationRuntime) Run() error {
	a.server = &http.Server{
		Addr:              ":" + a.cfg.HTTPPort,
		Handler:           a.handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	slog.Info("server starting", "port", a.server.Addr)
	err := a.server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("server listen: %w", err)
	}
	return nil
}

func (a *ApplicationRuntime) Shutdown(ctx context.Context) error {
	slog.Info("server shutting down")
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}
	slog.Info("server stopped")
	return nil
}

func (a *ApplicationRuntime) Close() error {
	slog.Info("server closing")
	if a.server != nil {
		return a.server.Close()
	}
	return nil
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMethodNotAllowed)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "method not allowed"})
}
