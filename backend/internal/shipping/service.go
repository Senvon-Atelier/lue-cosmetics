// Package shipping owns the shipping-quote calculation. Config is loaded
// from a JSON file at process startup and cached in memory; changes require
// a server restart.
package shipping

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds the shipping rate parameters loaded from JSON.
type Config struct {
	FlatRateGhsMinor int64 `json:"flat_rate_ghs_minor"`
	FreeOverGhsMinor int64 `json:"free_over_ghs_minor"`
}

// Quote is the wire-format response for a shipping cost calculation.
type Quote struct {
	FlatRateGhsMinor              int64 `json:"flat_rate_ghs_minor"`
	FreeOverGhsMinor              int64 `json:"free_over_ghs_minor"`
	AppliedCostGhsMinor           int64 `json:"applied_cost_ghs_minor"`
	FreeShippingRemainderGhsMinor int64 `json:"free_shipping_remainder_ghs_minor"`
}

// Service holds an immutable shipping Config and is safe for concurrent use.
type Service struct {
	cfg Config
}

// LoadConfig reads and parses a shipping config JSON file.
func LoadConfig(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("shipping: read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("shipping: parse config: %w", err)
	}
	return cfg, nil
}

// New returns a Service from an already-parsed Config (no I/O).
func New(cfg Config) *Service {
	return &Service{cfg: cfg}
}

// NewService is a convenience wrapper: LoadConfig + New. Tests and tooling
// that already hold a file path can call this directly.
func NewService(configPath string) (*Service, error) {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	return New(cfg), nil
}

// Quote calculates the shipping cost for the given cart subtotal (in pesewas).
// If subtotal >= FreeOverGhsMinor, shipping is free. Otherwise the flat rate applies.
func (s *Service) Quote(subtotal int64) Quote {
	q := Quote{
		FlatRateGhsMinor: s.cfg.FlatRateGhsMinor,
		FreeOverGhsMinor: s.cfg.FreeOverGhsMinor,
	}
	if subtotal >= s.cfg.FreeOverGhsMinor {
		q.AppliedCostGhsMinor = 0
		q.FreeShippingRemainderGhsMinor = 0
		return q
	}
	q.AppliedCostGhsMinor = s.cfg.FlatRateGhsMinor
	q.FreeShippingRemainderGhsMinor = s.cfg.FreeOverGhsMinor - subtotal
	return q
}
