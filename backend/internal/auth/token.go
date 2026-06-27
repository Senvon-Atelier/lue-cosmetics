package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
)

const tokenBytes = 32

// NewToken returns a fresh URL-safe session/verification token: 32 random
// bytes encoded with base64.URLEncoding without padding (43 chars).
func NewToken() (string, error) {
	b := make([]byte, tokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashToken returns sha256 of the decoded bytes if the input is valid
// base64-url; otherwise sha256 of the raw input. Either way it is suitable
// for use as a DB lookup key — but the "raw-input" fallback path is unlikely
// to match anything we ourselves issued.
func HashToken(raw string) [32]byte {
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return sha256.Sum256([]byte(raw))
	}
	return sha256.Sum256(b)
}

// TokenEquals compares a presented raw token against a stored hash in
// constant time.
func TokenEquals(raw string, stored [32]byte) bool {
	h := HashToken(raw)
	return subtle.ConstantTimeCompare(h[:], stored[:]) == 1
}
