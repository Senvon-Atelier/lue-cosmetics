package auth_test

import (
	"strings"
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/auth"
)

func TestHashFormat(t *testing.T) {
	h, err := auth.Hash("hunter2", auth.TestParams)
	if err != nil {
		t.Fatalf("Hash: %v", err)
	}
	if !strings.HasPrefix(h, "$argon2id$") {
		t.Errorf("prefix wrong: %s", h)
	}
	parts := strings.Split(h, "$")
	if len(parts) != 6 {
		t.Errorf("expected 6 segments, got %d in %s", len(parts), h)
	}
}

func TestVerifyRoundTrip(t *testing.T) {
	h, _ := auth.Hash("hunter2", auth.TestParams)
	ok, err := auth.Verify("hunter2", h)
	if err != nil || !ok {
		t.Errorf("Verify(correct) = %v, %v", ok, err)
	}
	ok, err = auth.Verify("wrong", h)
	if err != nil || ok {
		t.Errorf("Verify(wrong) = %v, %v", ok, err)
	}
}

func TestVerifyRejectsMalformed(t *testing.T) {
	if _, err := auth.Verify("x", "not-a-phc-string"); err == nil {
		t.Fatal("expected error on malformed hash")
	}
	if _, err := auth.Verify("x", "$argon2id$v=19$bad"); err == nil {
		t.Fatal("expected error on truncated hash")
	}
}

func TestHashIsRandomised(t *testing.T) {
	h1, _ := auth.Hash("hunter2", auth.TestParams)
	h2, _ := auth.Hash("hunter2", auth.TestParams)
	if h1 == h2 {
		t.Errorf("expected different salts → different hashes")
	}
}
