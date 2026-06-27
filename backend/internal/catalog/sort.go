// Package catalog serves the public catalog API: products, categories, brands.
package catalog

import "fmt"

// SortKey is an allowlisted sort column. Unknown values are rejected at
// parse time — the column name never reaches SQL via string concatenation.
type SortKey int

const (
	SortNewest SortKey = iota
	SortPriceAsc
	SortPriceDesc
	SortRatingDesc
	SortNameAsc
)

// ParseSort accepts the public query-string values for ?sort=. The empty
// string maps to SortNewest (the default).
func ParseSort(s string) (SortKey, error) {
	switch s {
	case "", "newest":
		return SortNewest, nil
	case "price_asc":
		return SortPriceAsc, nil
	case "price_desc":
		return SortPriceDesc, nil
	case "rating_desc":
		return SortRatingDesc, nil
	case "name_asc":
		return SortNameAsc, nil
	}
	return 0, fmt.Errorf("unknown sort %q", s)
}
