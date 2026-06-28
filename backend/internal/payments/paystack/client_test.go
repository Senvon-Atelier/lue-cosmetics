package paystack

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newServer(t *testing.T, handler http.HandlerFunc) (*Client, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := NewClient(srv.URL, "sk_test_secret")
	return c, srv.Close
}

func TestClient_InitializeTransaction_Happy(t *testing.T) {
	c, stop := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/transaction/initialize" {
			t.Errorf("expected /transaction/initialize, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk_test_secret" {
			t.Errorf("expected bearer token, got %q", got)
		}
		body, _ := io.ReadAll(r.Body)
		var in InitializeTransactionInput
		if err := json.Unmarshal(body, &in); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if in.Email != "buyer@example.com" || in.Amount != 12500 || in.Reference != "RUE-CAFEBABE" {
			t.Errorf("unexpected body: %+v", in)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": true,
			"message": "Authorization URL created",
			"data": {
				"authorization_url": "https://checkout.paystack.com/abc",
				"access_code": "ac_123",
				"reference": "RUE-CAFEBABE"
			}
		}`))
	})
	defer stop()

	out, err := c.InitializeTransaction(context.Background(), InitializeTransactionInput{
		Email:     "buyer@example.com",
		Amount:    12500,
		Reference: "RUE-CAFEBABE",
		Callback:  "https://app.example.com/checkout/return",
		Currency:  "GHS",
	})
	if err != nil {
		t.Fatalf("initialize: %v", err)
	}
	if out.AuthorizationURL != "https://checkout.paystack.com/abc" {
		t.Errorf("authorization_url = %q", out.AuthorizationURL)
	}
	if out.Reference != "RUE-CAFEBABE" {
		t.Errorf("reference = %q", out.Reference)
	}
}

func TestClient_VerifyTransaction_Happy(t *testing.T) {
	c, stop := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/transaction/verify/RUE-CAFEBABE" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk_test_secret" {
			t.Errorf("expected bearer token, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status": true,
			"message": "Verification successful",
			"data": {
				"reference": "RUE-CAFEBABE",
				"status": "success",
				"amount": 12500,
				"currency": "GHS",
				"gateway_response": "Successful",
				"id": 9999
			}
		}`))
	})
	defer stop()

	out, err := c.VerifyTransaction(context.Background(), "RUE-CAFEBABE")
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if out.Reference != "RUE-CAFEBABE" {
		t.Errorf("reference = %q", out.Reference)
	}
	if out.Status != "success" {
		t.Errorf("status = %q", out.Status)
	}
	if out.TransactionID != 9999 {
		t.Errorf("transaction id = %d", out.TransactionID)
	}
}

func TestClient_InitializeTransaction_Non2xx(t *testing.T) {
	c, stop := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"status":false,"message":"Invalid key"}`, http.StatusUnauthorized)
	})
	defer stop()

	_, err := c.InitializeTransaction(context.Background(), InitializeTransactionInput{
		Email: "buyer@example.com", Amount: 12500, Reference: "RUE-FFFFFFFF",
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrPaystackUpstream) {
		t.Errorf("expected ErrPaystackUpstream, got %v", err)
	}
}

func TestClient_VerifyTransaction_StatusFalseEnvelope(t *testing.T) {
	c, stop := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":false,"message":"Transaction not found","data":{}}`))
	})
	defer stop()

	_, err := c.VerifyTransaction(context.Background(), "RUE-AAAAAAAA")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrPaystackUpstream) {
		t.Errorf("expected ErrPaystackUpstream, got %v", err)
	}
	if !strings.Contains(err.Error(), "Transaction not found") {
		t.Errorf("expected wrapped message in error, got %v", err)
	}
}

func TestClient_InitializeTransaction_MalformedJSON(t *testing.T) {
	c, stop := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{not-json`))
	})
	defer stop()

	_, err := c.InitializeTransaction(context.Background(), InitializeTransactionInput{
		Email: "buyer@example.com", Amount: 12500, Reference: "RUE-BBBBBBBB",
	})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, ErrPaystackUpstream) {
		t.Errorf("expected ErrPaystackUpstream, got %v", err)
	}
}

func TestClient_IsConfigured(t *testing.T) {
	if (&Client{}).IsConfigured() {
		t.Error("expected zero client to be unconfigured")
	}
	if (&Client{BaseURL: "x"}).IsConfigured() {
		t.Error("expected client missing secret to be unconfigured")
	}
	if !(&Client{BaseURL: "x", SecretKey: "y"}).IsConfigured() {
		t.Error("expected fully populated client to be configured")
	}
}
