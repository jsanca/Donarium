package runtime

import (
	"context"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"donarium/server/internal/platform/http/health"
)

type PlatformRuntime struct {
	checker health.ReadinessChecker
	logger  *slog.Logger
}

func NewPlatformRuntime(pool *pgxpool.Pool, logger *slog.Logger) *PlatformRuntime {
	return &PlatformRuntime{
		checker: &poolChecker{pool: pool},
		logger:  logger,
	}
}

func (p *PlatformRuntime) RegisterRoutes(r chi.Router) {
	r.Handle("/health/live", health.LivenessHandler())
	r.Handle("/health/ready", health.ReadinessHandler(p.checker))
}

type poolChecker struct {
	pool *pgxpool.Pool
}

func (c *poolChecker) Check(ctx context.Context) error {
	return c.pool.Ping(ctx)
}
