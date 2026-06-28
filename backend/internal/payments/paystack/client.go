// Package paystack is a thin REST client for the two Paystack endpoints we
// need: transaction/initialize and transaction/verify. It also exposes
// VerifyWebhookSignature for the webhook handler.
//
// The client makes NO ASSUMPTION about the surrounding service — it's
// reusable from anywhere, returns typed responses, and never panics on
// missing fields (Paystack occasionally omits optional fields).
package paystack

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	BaseURL    string // e.g., "https://api.paystack.co"
	SecretKey  string // sk_test_xxxxx or sk_live_xxxxx
	HTTPClient *http.Client
}

func NewClient(baseURL, secretKey string) *Client {
	return &Client{
		BaseURL:    baseURL,
		SecretKey:  secretKey,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// IsConfigured returns true when both BaseURL and SecretKey are non-empty.
// Handlers should 503 with "not_configured" otherwise.
func (c *Client) IsConfigured() bool {
	return c.BaseURL != "" && c.SecretKey != ""
}

// InitializeTransactionInput models the subset of fields we send.
type InitializeTransactionInput struct {
	Email     string   `json:"email"`
	Amount    int64    `json:"amount"`                // pesewas
	Reference string   `json:"reference"`             // "RUE-XXXXXXXX"
	Callback  string   `json:"callback_url,omitempty"`
	Currency  string   `json:"currency,omitempty"`    // "GHS"
	Channels  []string `json:"channels,omitempty"`    // e.g. ["card","mobile_money"]
}

// InitializeTransactionOutput is the relevant subset of Paystack's response.
type InitializeTransactionOutput struct {
	AuthorizationURL string `json:"authorization_url"`
	AccessCode       string `json:"access_code"`
	Reference        string `json:"reference"`
}

var ErrPaystackUpstream = errors.New("paystack: upstream error")

// InitializeTransaction calls POST /transaction/initialize. On Paystack
// returning non-2xx, the error wraps ErrPaystackUpstream.
func (c *Client) InitializeTransaction(ctx context.Context, in InitializeTransactionInput) (InitializeTransactionOutput, error) {
	var out InitializeTransactionOutput
	body, _ := json.Marshal(in)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/transaction/initialize", bytes.NewReader(body))
	if err != nil {
		return out, err
	}
	req.Header.Set("Authorization", "Bearer "+c.SecretKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return out, fmt.Errorf("%w: %v", ErrPaystackUpstream, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return out, fmt.Errorf("%w: status %d body=%s", ErrPaystackUpstream, resp.StatusCode, string(raw))
	}
	var envelope struct {
		Status  bool                        `json:"status"`
		Message string                      `json:"message"`
		Data    InitializeTransactionOutput `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return out, fmt.Errorf("%w: parse: %v", ErrPaystackUpstream, err)
	}
	if !envelope.Status {
		return out, fmt.Errorf("%w: %s", ErrPaystackUpstream, envelope.Message)
	}
	return envelope.Data, nil
}

// VerifyTransactionOutput is the subset we read from the verify endpoint.
type VerifyTransactionOutput struct {
	Reference       string `json:"reference"`
	Status          string `json:"status"` // "success", "failed", "abandoned", ...
	Amount          int64  `json:"amount"`
	Currency        string `json:"currency"`
	GatewayResponse string `json:"gateway_response"`
	TransactionID   int64  `json:"id"` // Paystack's numeric transaction id; convert to string at the boundary
}

// VerifyTransaction calls GET /transaction/verify/{reference}.
func (c *Client) VerifyTransaction(ctx context.Context, reference string) (VerifyTransactionOutput, error) {
	var out VerifyTransactionOutput
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/transaction/verify/"+reference, nil)
	if err != nil {
		return out, err
	}
	req.Header.Set("Authorization", "Bearer "+c.SecretKey)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return out, fmt.Errorf("%w: %v", ErrPaystackUpstream, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode/100 != 2 {
		return out, fmt.Errorf("%w: status %d body=%s", ErrPaystackUpstream, resp.StatusCode, string(raw))
	}
	var envelope struct {
		Status  bool                    `json:"status"`
		Message string                  `json:"message"`
		Data    VerifyTransactionOutput `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return out, fmt.Errorf("%w: parse: %v", ErrPaystackUpstream, err)
	}
	if !envelope.Status {
		return out, fmt.Errorf("%w: %s", ErrPaystackUpstream, envelope.Message)
	}
	return envelope.Data, nil
}
