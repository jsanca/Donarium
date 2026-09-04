package http

import (
	"context"

	"donarium/server/internal/identity/pgx"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IdentityRuntime struct {
	setupHandler *SetupHandler
	loginHandler *LoginHandler
	meHandler    *MeHandler
	authMW       *AuthenticationMiddleware
}

func NewIdentityRuntime(
	pool *pgxpool.Pool,
	setup SetupPerformer,
	login LoginPerformer,
	cookieWriter SessionCookieWriter,
	authMW *AuthenticationMiddleware,
) *IdentityRuntime {
	orgRepo := pgx.NewOrganizationRepo()
	statusReader := &poolStatusReader{pool: pool, orgRepo: orgRepo}

	return &IdentityRuntime{
		setupHandler: NewSetupHandler(setup, statusReader),
		loginHandler: NewLoginHandler(login, pool, cookieWriter),
		meHandler:    NewMeHandler(),
		authMW:       authMW,
	}
}

func (r *IdentityRuntime) RegisterRoutes(router chi.Router) {
	router.HandleFunc("/api/setup", r.setupHandler.Setup)
	router.HandleFunc("/api/setup/status", r.setupHandler.Status)
	router.HandleFunc("/api/auth/login", r.loginHandler.Login)

	router.Route("/api/auth/me", func(protected chi.Router) {
		protected.Use(r.authMW.RequireAuthentication)
		protected.Use(RequireAuthenticated())
		protected.HandleFunc("/", r.meHandler.Me)
	})
}

type poolStatusReader struct {
	pool    *pgxpool.Pool
	orgRepo *pgx.OrganizationRepo
}

func (s *poolStatusReader) IsInitialized(ctx context.Context) (bool, error) {
	return s.orgRepo.ExistsAny(ctx, pgx.NewExecutorFromPool(s.pool))
}
