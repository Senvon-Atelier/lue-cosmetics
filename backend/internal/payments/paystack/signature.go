package paystack

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
)

// VerifyWebhookSignature checks the x-paystack-signature header against
// HMAC-SHA512(secret, body). Returns true iff the signatures match in
// constant time.
func VerifyWebhookSignature(secret string, body []byte, headerValue string) bool {
	if secret == "" || headerValue == "" || len(body) == 0 {
		return false
	}
	mac := hmac.New(sha512.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(headerValue))
}
