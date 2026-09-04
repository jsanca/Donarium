package http

import (
	"donarium/server/internal/identity/http"
	"donarium/server/internal/properties/application"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Runtime struct {
	handler *Handler
	authMW  *http.AuthenticationMiddleware
}

func NewRuntime(pool *pgxpool.Pool, service *application.Service, authMW *http.AuthenticationMiddleware) *Runtime {
	return &Runtime{
		handler: NewHandler(service, pool),
		authMW:  authMW,
	}
}

func (r *Runtime) RegisterRoutes(router chi.Router) {
	router.Route("/api/properties", func(protected chi.Router) {
		protected.Use(r.authMW.RequireAuthentication)
		protected.Use(http.RequireAuthenticated())
		protected.HandleFunc("/", r.handler.Collection)
		protected.HandleFunc("", r.handler.Collection)
		protected.HandleFunc("/{id}", r.handler.GetProperty)
	})
}
