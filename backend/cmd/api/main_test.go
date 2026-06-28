package main_test

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

func TestServerBootsAndHealthzReturnsOK(t *testing.T) {
	url, stop := testsupport.StartPostgres(t)
	defer stop()

	testsupport.Migrate(t, url, "../../migrations")

	wd, _ := os.Getwd()
	root := filepath.Join(wd, "..", "..")
	bin := filepath.Join(t.TempDir(), "api")
	build := exec.Command("go", "build", "-o", bin, "./cmd/api")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %v\n%s", err, out)
	}

	// Build absolute path to the shipping config so the binary can find it
	// regardless of its working directory.
	shipConfigPath, err := filepath.Abs(filepath.Join(root, "seed", "config", "shipping_config.json"))
	if err != nil {
		t.Fatalf("shipping config abs: %v", err)
	}

	cmd := exec.Command(bin)
	cmd.Env = append(os.Environ(),
		"PORT=18080",
		"ENV=development",
		"DATABASE_URL="+url,
		"CORS_ORIGINS=http://localhost:5173",
		"LOG_LEVEL=debug",
		"SHIPPING_CONFIG_PATH="+shipConfigPath,
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
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) //nolint:errcheck

	// Verify /api/v1/categories is reachable (empty array — no seed data).
	resp, err = http.Get("http://127.0.0.1:18080/api/v1/categories")
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			t.Fatalf("/api/v1/categories code = %d", resp.StatusCode)
		}
		t.Fatalf("/api/v1/categories failed: %v", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) //nolint:errcheck

	// Verify /api/v1/shipping/quote returns a valid quote for a below-threshold subtotal.
	resp, err = http.Get("http://127.0.0.1:18080/api/v1/shipping/quote?subtotal=10000")
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			t.Fatalf("/shipping/quote code = %d", resp.StatusCode)
		}
		t.Fatalf("/shipping/quote failed: %v", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) //nolint:errcheck

	// POST /api/v1/auth/signup → 201, capture cookie
	signupBody := strings.NewReader(`{"email":"smoke@y.test","password":"hunter22","name":"Smoke"}`)
	signupReq, _ := http.NewRequest("POST", "http://127.0.0.1:18080/api/v1/auth/signup", signupBody)
	signupReq.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(signupReq)
	if err != nil {
		t.Fatalf("signup: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("signup code = %d", resp.StatusCode)
	}
	sessionCookie := testsupport.FindCookie(resp, "rue_session")
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if sessionCookie == nil {
		t.Fatal("no rue_session cookie set on signup")
	}

	// GET /api/v1/me with cookie → 200
	meReq, _ := http.NewRequest("GET", "http://127.0.0.1:18080/api/v1/me", nil)
	meReq.AddCookie(sessionCookie)
	resp, err = http.DefaultClient.Do(meReq)
	if err != nil {
		t.Fatalf("me: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("me code = %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
}
