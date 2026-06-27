package shipping_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/shipping"
)

func writeConfig(t *testing.T, flat, freeOver int64) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "shipping_config.json")
	body := []byte(
		`{"flat_rate_ghs_minor":` + itoa(flat) + `,"free_over_ghs_minor":` + itoa(freeOver) + `}`)
	if err := os.WriteFile(p, body, 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	return p
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}

func TestQuoteBelowThresholdChargesFlat(t *testing.T) {
	p := writeConfig(t, 2500, 50000)
	s, err := shipping.NewService(p)
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}
	q := s.Quote(10000)
	if q.AppliedCostGhsMinor != 2500 {
		t.Errorf("applied = %d, want 2500", q.AppliedCostGhsMinor)
	}
	if q.FreeShippingRemainderGhsMinor != 40000 {
		t.Errorf("remainder = %d, want 40000", q.FreeShippingRemainderGhsMinor)
	}
}

func TestQuoteAboveThresholdIsFree(t *testing.T) {
	p := writeConfig(t, 2500, 50000)
	s, _ := shipping.NewService(p)
	q := s.Quote(50000)
	if q.AppliedCostGhsMinor != 0 || q.FreeShippingRemainderGhsMinor != 0 {
		t.Errorf("quote at threshold = %+v", q)
	}
	q = s.Quote(75000)
	if q.AppliedCostGhsMinor != 0 || q.FreeShippingRemainderGhsMinor != 0 {
		t.Errorf("quote above threshold = %+v", q)
	}
}

func TestQuoteZeroSubtotal(t *testing.T) {
	p := writeConfig(t, 2500, 50000)
	s, _ := shipping.NewService(p)
	q := s.Quote(0)
	if q.AppliedCostGhsMinor != 2500 || q.FreeShippingRemainderGhsMinor != 50000 {
		t.Errorf("quote zero = %+v", q)
	}
}

func TestNewServiceRejectsMissingFile(t *testing.T) {
	if _, err := shipping.NewService("/nonexistent/config.json"); err == nil {
		t.Fatal("expected error for missing file")
	}
}
