package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/oti-adjei/ruecosmetics/internal/db"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

func TestMigrationsApply(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	testsupport.Migrate(t, url, "../../migrations")

	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	defer pool.Close()
	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM health_marker`).Scan(&n); err != nil {
		t.Fatalf("query: %v", err)
	}
	if n < 1 {
		t.Errorf("health_marker rows = %d, want >= 1", n)
	}
}
