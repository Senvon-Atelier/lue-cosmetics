// Package testsupport provides shared helpers for integration tests.
package testsupport

import (
	"context"
	"testing"

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
