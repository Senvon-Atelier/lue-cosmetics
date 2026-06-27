package db_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	sqlcq "github.com/oti-adjei/ruecosmetics/internal/db/sqlc"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
	"github.com/pressly/goose/v3"
)

func TestSqlcCountHealthMarkers(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sqlDB, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer sqlDB.Close()

	migDir, err := filepath.Abs("../../migrations")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("dialect: %v", err)
	}
	if err := goose.UpContext(ctx, sqlDB, migDir); err != nil {
		t.Fatalf("up: %v", err)
	}

	pool, err := db.NewPool(ctx, url)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	defer pool.Close()

	q := sqlcq.New(pool)
	n, err := q.CountHealthMarkers(ctx)
	if err != nil {
		t.Fatalf("count: %v", err)
	}
	if n < 1 {
		t.Errorf("n = %d, want >= 1", n)
	}
}
