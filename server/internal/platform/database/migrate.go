package database

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	if err := ensureMigrationTable(ctx, pool); err != nil {
		return fmt.Errorf("ensure migration table: %w", err)
	}

	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migration files: %w", err)
	}

	upFiles := make(map[string]string)
	for _, f := range entries {
		name := f.Name()
		if strings.HasSuffix(name, ".up.sql") {
			base := strings.TrimSuffix(name, ".up.sql")
			upFiles[base] = name
		}
	}

	var migrations []string
	for base := range upFiles {
		migrations = append(migrations, base)
	}
	sort.Strings(migrations)

	for _, base := range migrations {
		applied, err := isMigrationApplied(ctx, pool, base)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", base, err)
		}
		if applied {
			continue
		}

		content, err := migrationFS.ReadFile("migrations/" + upFiles[base])
		if err != nil {
			return fmt.Errorf("read migration %s: %w", base, err)
		}

		slog.Info("applying migration", "name", base)
		if _, err := pool.Exec(ctx, string(content)); err != nil {
			return fmt.Errorf("apply migration %s: %w", base, err)
		}

		if err := recordMigration(ctx, pool, base); err != nil {
			return fmt.Errorf("record migration %s: %w", base, err)
		}
	}

	return nil
}

func ensureMigrationTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func isMigrationApplied(ctx context.Context, pool *pgxpool.Pool, name string) (bool, error) {
	var count int
	err := pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM schema_migrations WHERE name = $1", name,
	).Scan(&count)
	return count > 0, err
}

func recordMigration(ctx context.Context, pool *pgxpool.Pool, name string) error {
	_, err := pool.Exec(ctx,
		"INSERT INTO schema_migrations (name) VALUES ($1)", name,
	)
	return err
}
