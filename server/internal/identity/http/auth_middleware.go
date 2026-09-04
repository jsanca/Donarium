package http

import (
	"context"
	"log/slog"
	"net/http"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/application/authentication"
	"donarium/server/internal/identity/pgx"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey struct{ name string }

var principalCtxKey = contextKey{name: "donarium-principal"}

func PrincipalFromContext(ctx context.Context) (authentication.AuthenticatedPrincipal, bool) {
	p, ok := ctx.Value(principalCtxKey).(authentication.AuthenticatedPrincipal)
	return p, ok
}

func WithPrincipal(ctx context.Context, p authentication.AuthenticatedPrincipal) context.Context {
	return context.WithValue(ctx, principalCtxKey, p)
}

func setPrincipal(ctx context.Context, p authentication.AuthenticatedPrincipal) context.Context {
	return context.WithValue(ctx, principalCtxKey, p)
}

type SessionCookieReadFn func(r *http.Request) (string, error)

type AuthenticationMiddleware struct {
	verifier authentication.SessionVerifier
	resolver authentication.PrincipalResolver
	pool     *pgxpool.Pool
	readCookie SessionCookieReadFn
}

func NewAuthenticationMiddleware(
	verifier authentication.SessionVerifier,
	resolver authentication.PrincipalResolver,
	pool *pgxpool.Pool,
	readCookie SessionCookieReadFn,
) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{
		verifier:    verifier,
		resolver:    resolver,
		pool:        pool,
		readCookie:  readCookie,
	}
}

func (m *AuthenticationMiddleware) RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := m.readCookie(r)
		if err != nil || token == "" {
			slog.Debug("auth middleware: no session cookie")
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		claims, err := m.verifier.Verify(token)
		if err != nil {
			slog.Debug("auth middleware: token verification failed", "error", err)
			mapAuthMiddlewareError(w, err)
			return
		}

		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			slog.Debug("auth middleware: invalid subject", "sub", claims.Subject)
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}

		principal, err := m.resolver.Resolve(r.Context(), pgx.NewExecutorFromPool(m.pool), identity.UserID(userID))
		if err != nil {
			slog.Debug("auth middleware: principal resolution failed", "error", err)
			mapAuthMiddlewareError(w, err)
			return
		}

		ctx := setPrincipal(r.Context(), principal)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireAuthenticated() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := PrincipalFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func mapAuthMiddlewareError(w http.ResponseWriter, err error) {
	switch {
	case err == identity.ErrInvalidSession,
		err == identity.ErrExpiredSession,
		err == identity.ErrInvalidCredentials,
		err == identity.ErrInvalidSession,
		err == identity.ErrExpiredSession:
		writeError(w, http.StatusUnauthorized, "authentication required")
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
