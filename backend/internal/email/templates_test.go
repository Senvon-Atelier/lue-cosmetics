package email

import (
	"strings"
	"testing"
)

func TestRenderer_OrderConfirmation_HappyData(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	data := map[string]any{
		"name":               "Ada",
		"paystack_reference": "RUE-DEADBEEF",
		"items": []map[string]any{
			{"name": "Lipstick", "brand": "LueBrand", "qty": 2, "line_total": "50.00"},
			{"name": "Mascara", "qty": 1, "line_total": "30.00"},
		},
		"subtotal": "80.00",
		"shipping": "5.00",
		"total":    "85.00",
	}
	html, text, err := r.Render("order_confirmation", data)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}

	for _, want := range []string{"Ada", "RUE-DEADBEEF", "Lipstick", "85.00"} {
		if !strings.Contains(html, want) {
			t.Errorf("html missing %q", want)
		}
		if !strings.Contains(text, want) {
			t.Errorf("text missing %q", want)
		}
	}
}

func TestRenderer_MissingKey_DoesNotPanic(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}
	data := map[string]any{
		"paystack_reference": "RUE-DEADBEEF",
		"items":              []map[string]any{},
		"subtotal":           "0.00",
		"shipping":           "0.00",
		"total":              "0.00",
	}
	html, text, err := r.Render("order_confirmation", data)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	// Per plan: missing key should not panic and should still produce output.
	if !strings.Contains(html, "RUE-DEADBEEF") {
		t.Errorf("expected html to render reference even with missing name; got: %s", html)
	}
	if !strings.Contains(text, "RUE-DEADBEEF") {
		t.Errorf("expected text to render reference even with missing name; got: %s", text)
	}
}
