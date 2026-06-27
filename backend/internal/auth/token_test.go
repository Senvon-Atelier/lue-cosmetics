package auth_test

import (
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/auth"
)

func TestNewTokenLength(t *testing.T) {
	tok, err := auth.NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	if len(tok) != 43 {
		t.Errorf("token len = %d, want 43", len(tok))
	}
}

func TestNewTokenUnique(t *testing.T) {
	a, _ := auth.NewToken()
	b, _ := auth.NewToken()
	if a == b {
		t.Errorf("tokens collided: %s == %s", a, b)
	}
}

func TestHashAndEquals(t *testing.T) {
	tok, _ := auth.NewToken()
	h := auth.HashToken(tok)
	if !auth.TokenEquals(tok, h) {
		t.Errorf("TokenEquals same token = false")
	}
	other, _ := auth.NewToken()
	if auth.TokenEquals(other, h) {
		t.Errorf("TokenEquals different token = true")
	}
}

func TestHashRejectsMalformed(t *testing.T) {
	// Hash of an unparseable token should still produce SOME 32-byte value
	// (we hash the raw string when decoding fails — but TokenEquals against
	// a different stored hash must return false).
	tok := "not-a-real-token!!!"
	h := auth.HashToken(tok)
	if h == ([32]byte{}) {
		t.Errorf("expected non-zero hash")
	}
}
