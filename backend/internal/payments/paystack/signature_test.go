package paystack

import "testing"

const (
	knownSecret    = "sk_test_secret"
	knownBody      = `{"event":"charge.success","data":{"reference":"RUE-DEADBEEF"}}`
	knownSignature = "c0c6217335b0589f72bbe4d40e0113a6dd209911dd008c2774e522ae1e03d2f12f088684f71fefe4ccb6271eee9dd2e83c18f04b002418c9dbac389f69310daa"
)

func TestVerifyWebhookSignature_KnownVector(t *testing.T) {
	if !VerifyWebhookSignature(knownSecret, []byte(knownBody), knownSignature) {
		t.Errorf("expected signature to verify for known vector")
	}
}

func TestVerifyWebhookSignature_TamperedBody(t *testing.T) {
	tampered := []byte(`{"event":"charge.success","data":{"reference":"RUE-CAFEBABE"}}`)
	if VerifyWebhookSignature(knownSecret, tampered, knownSignature) {
		t.Errorf("expected tampered body to fail signature check")
	}
}

func TestVerifyWebhookSignature_EmptyHeader(t *testing.T) {
	if VerifyWebhookSignature(knownSecret, []byte(knownBody), "") {
		t.Errorf("expected empty header to fail")
	}
}

func TestVerifyWebhookSignature_EmptyBody(t *testing.T) {
	if VerifyWebhookSignature(knownSecret, []byte{}, knownSignature) {
		t.Errorf("expected empty body to fail")
	}
}

func TestVerifyWebhookSignature_EmptySecret(t *testing.T) {
	if VerifyWebhookSignature("", []byte(knownBody), knownSignature) {
		t.Errorf("expected empty secret to fail")
	}
}
