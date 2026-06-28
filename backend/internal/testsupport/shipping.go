package testsupport

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// WriteShippingConfig writes a shipping_config.json file to a temp directory
// and returns the path. flat and freeOver are pesewas.
func WriteShippingConfig(t *testing.T, flat, freeOver int64) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "shipping_config.json")
	body := []byte(`{"flat_rate_ghs_minor":` + strconv.FormatInt(flat, 10) +
		`,"free_over_ghs_minor":` + strconv.FormatInt(freeOver, 10) + `}`)
	if err := os.WriteFile(p, body, 0o644); err != nil {
		t.Fatalf("write shipping config: %v", err)
	}
	return p
}
