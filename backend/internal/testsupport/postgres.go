// Package testsupport provides shared helpers for integration tests.
package testsupport

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

// StartPostgres launches an ephemeral Postgres 16 container scoped to the test t.
// Returns the connection URL and a stop function the caller must defer.
func StartPostgres(t *testing.T) (string, func()) {
	t.Helper()
	ctx := context.Background()
	pg, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("ruetest"),
		postgres.WithUsername("rue"),
		postgres.WithPassword("rue_dev"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("start postgres: %v", err)
	}
	url, err := pg.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = pg.Terminate(ctx)
		t.Fatalf("conn string: %v", err)
	}
	return url, func() { _ = pg.Terminate(ctx) }
}

// Migrate applies all goose .sql migrations under migrationsRelPath (relative
// to the test's working directory) against the given Postgres URL.
func Migrate(t *testing.T, url, migrationsRelPath string) {
	t.Helper()
	sqlDB, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer sqlDB.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("dialect: %v", err)
	}
	migDir, err := filepath.Abs(migrationsRelPath)
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	if err := goose.UpContext(context.Background(), sqlDB, migDir); err != nil {
		t.Fatalf("up: %v", err)
	}
}
