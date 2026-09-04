package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"donarium/server/internal/identity"
	"donarium/server/internal/identity/application"
	"donarium/server/internal/identity/application/authentication"
	identityhttp "donarium/server/internal/identity/http"
	identitypgx "donarium/server/internal/identity/pgx"
	"donarium/server/internal/platform/config"
	"donarium/server/internal/platform/database"
	"donarium/server/internal/platform/runtime"
	propertiesApp "donarium/server/internal/properties/application"
	propertieshttp "donarium/server/internal/properties/http"
	propertiespgx "donarium/server/internal/properties/pgx"
)

func main() {
	os.Exit(run())
}

func run() int {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		return 1
	}

	pool, err := database.NewPool(context.Background(), *cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return 1
	}
	defer pool.Close()
	slog.Info("database connected")

	if err := database.RunMigrations(context.Background(), pool); err != nil {
		slog.Error("failed to run migrations", "error", err)
		return 1
	}
	slog.Info("migrations complete")

	userRepo := identitypgx.NewUserRepo()
	credRepo := identitypgx.NewCredentialRepo()
	orgRepo := identitypgx.NewOrganizationRepo()
	memRepo := identitypgx.NewMembershipRepo()
	grantRepo := identitypgx.NewPlatformGrantRepo()
	hasher := identitypgx.NewArgon2Hasher()
	normalizer := identity.NewDefaultEmailNormalizer()
	txManager := identitypgx.NewTransactionManager(pool)

	canonicalSetup := application.NewCanonicalSetupService(
		userRepo, credRepo, orgRepo, memRepo, grantRepo,
		hasher, normalizer,
	)
	txSetup := application.NewTransactionalSetupService(canonicalSetup, txManager)

	sessionHandler := identitypgx.NewHMACSessionIssuer(cfg.SessionSigningKey, cfg.SessionTTL)

	principalResolver := authentication.NewPrincipalResolverService(
		userRepo, grantRepo, memRepo, orgRepo,
	)

	authService := authentication.NewAuthenticateUserService(
		userRepo, credRepo,
		hasher, normalizer, sessionHandler, principalResolver,
	)

	cookieSecure := cfg.IsProduction()
	cookieWriter := identityhttp.NewCookieSessionHandler("donarium_session", "/", cookieSecure, cfg.SessionTTL)

	authMiddleware := identityhttp.NewAuthenticationMiddleware(
		sessionHandler, principalResolver, pool, cookieWriter.Read,
	)

	logger.Info("server configured",
		"environment", cfg.Environment,
		"cookie_secure", cookieSecure,
	)

	platformRuntime := runtime.NewPlatformRuntime(pool, logger)
	identityRuntime := identityhttp.NewIdentityRuntime(pool, txSetup, authService, cookieWriter, authMiddleware)

	propertyRepo := propertiespgx.NewRepository()
	stakeholderRepo := propertiespgx.NewStakeholderRepository()
	propertyTx := propertiespgx.NewTransactionManager(pool)
	propertyService := propertiesApp.NewServiceWithStakeholders(propertyRepo, stakeholderRepo, propertyTx)
	propertiesRuntime := propertieshttp.NewRuntime(pool, propertyService, authMiddleware)

	appRuntime := runtime.NewApplication(*cfg, platformRuntime, identityRuntime, propertiesRuntime)
	defer func() {
		if err := appRuntime.Close(); err != nil {
			slog.Error("runtime close error", "error", err)
		}
	}()

	errCh := make(chan error, 1)
	go func() {
		errCh <- appRuntime.Run()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		slog.Info("signal received", "signal", sig.String())
	case err := <-errCh:
		if err != nil {
			slog.Error("server error", "error", err)
			return 1
		}
		slog.Info("server stopped cleanly")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := appRuntime.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
		return 1
	}

	return 0
}
