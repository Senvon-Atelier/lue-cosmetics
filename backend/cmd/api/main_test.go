package main_test

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestServerBootsAndHealthzReturnsOK(t *testing.T) {
	ctx := context.Background()
	pg, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("ruetest"), postgres.WithUsername("rue"), postgres.WithPassword("rue_dev"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("pg: %v", err)
	}
	defer pg.Terminate(ctx)
	url, _ := pg.ConnectionString(ctx, "sslmode=disable")

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
}
