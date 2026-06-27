package main_test

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
	"github.com/pressly/goose/v3"
)

// migrate applies all goose migrations from backend/migrations.
func migrate(t *testing.T, url string) {
	t.Helper()
	sqlDB, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer sqlDB.Close()
	if err := goose.SetDialect("postgres"); err != nil {
		t.Fatalf("dialect: %v", err)
	}
	migDir, err := filepath.Abs("../../migrations")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	if err := goose.UpContext(context.Background(), sqlDB, migDir); err != nil {
		t.Fatalf("up: %v", err)
	}
}

func TestServerBootsAndHealthzReturnsOK(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	migrate(t, url)

	wd, _ := os.Getwd()
	root := filepath.Join(wd, "..", "..")
	bin := filepath.Join(t.TempDir(), "api")
	build := exec.Command("go", "build", "-o", bin, "./cmd/api")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		"PORT=18080",
		"ENV=development",
		"DATABASE_URL="+url,
		"CORS_ORIGINS=http://localhost:5173",
		"LOG_LEVEL=debug",
	)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	if err := cmd.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	deadline := time.Now().Add(10 * time.Second)
	var resp *http.Response
	var err error
	for time.Now().Before(deadline) {
		resp, err = http.Get("http://127.0.0.1:18080/healthz")
		if err == nil && resp.StatusCode == 200 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if err != nil || resp == nil || resp.StatusCode != 200 {
		if resp != nil {
			t.Fatalf("healthz code = %d", resp.StatusCode)
		}
		t.Fatalf("healthz never returned 200: %v", err)
	}

	// Verify /api/v1/categories is reachable (empty array — no seed data).
	resp, err = http.Get("http://127.0.0.1:18080/api/v1/categories")
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			t.Fatalf("/api/v1/categories code = %d", resp.StatusCode)
		}
		t.Fatalf("/api/v1/categories failed: %v", err)
	}
}
