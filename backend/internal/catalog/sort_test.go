package catalog_test

import (
	"testing"

	"github.com/oti-adjei/ruecosmetics/internal/catalog"
)

func TestParseSortAllowlist(t *testing.T) {
	cases := []struct {
		in   string
		want catalog.SortKey
	}{
		{"", catalog.SortNewest},
		{"newest", catalog.SortNewest},
		{"price_asc", catalog.SortPriceAsc},
		{"price_desc", catalog.SortPriceDesc},
		{"rating_desc", catalog.SortRatingDesc},
		{"name_asc", catalog.SortNameAsc},
	}
	for _, c := range cases {
		got, err := catalog.ParseSort(c.in)
		if err != nil {
			t.Errorf("ParseSort(%q): unexpected error %v", c.in, err)
		}
		if got != c.want {
			t.Errorf("ParseSort(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestParseSortRejectsUnknown(t *testing.T) {
	if _, err := catalog.ParseSort("DROP TABLE products"); err == nil {
		t.Fatal("expected error for unknown sort")
	}
	if _, err := catalog.ParseSort("price"); err == nil {
		t.Fatal("expected error for partial match")
	}
}
