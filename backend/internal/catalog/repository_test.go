package catalog_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/oti-adjei/ruecosmetics/internal/catalog"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	"github.com/oti-adjei/ruecosmetics/internal/testsupport"
)

// seedSmall inserts one category, one brand, three products with different
// prices/ratings so sort/filter tests have something to bite on.
// Products are inserted with explicit created_at offsets so newest-first
// ordering is deterministic.
func seedSmall(t *testing.T, pool db.Pool) {
	t.Helper()
	ctx := context.Background()
	_, err := pool.Exec(ctx, `
		INSERT INTO categories (slug, label, sort_order) VALUES
			('skincare', 'Skincare', 1),
			('haircare', 'Haircare', 2);
		INSERT INTO brands (slug, name) VALUES
			('nuxe', 'Nuxe'),
			('cantu', 'Cantu');
	`)
	if err != nil {
		t.Fatalf("seed categories/brands: %v", err)
	}
	// Insert products with staggered created_at so ORDER BY created_at DESC
	// returns: gentle-cleanser (newest), curl-cream, rose-serum (oldest).
	_, err = pool.Exec(ctx, `
		INSERT INTO products (slug, name, brand_id, category_id,
		                      price_ghs_minor, tone, size, rating, review_count, tags, created_at)
		SELECT 'rose-serum', 'Rose Serum', b.id, c.id,
		       24500, 'rose', '30 ml', 4.8, 142, ARRAY['Bestseller'], now() - interval '2 minutes'
		FROM brands b, categories c WHERE b.slug='nuxe' AND c.slug='skincare';
	`)
	if err != nil {
		t.Fatalf("seed rose-serum: %v", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO products (slug, name, brand_id, category_id,
		                      price_ghs_minor, tone, size, rating, review_count, tags, created_at)
		SELECT 'curl-cream', 'Curl Cream', b.id, c.id,
		       8800, 'lavender', '340 g', 4.7, 512, ARRAY['Bestseller','New'], now() - interval '1 minute'
		FROM brands b, categories c WHERE b.slug='cantu' AND c.slug='haircare';
	`)
	if err != nil {
		t.Fatalf("seed curl-cream: %v", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO products (slug, name, brand_id, category_id,
		                      price_ghs_minor, tone, size, rating, review_count, tags, created_at)
		SELECT 'gentle-cleanser', 'Gentle Cleanser', b.id, c.id,
		       13500, 'lavender', '236 ml', 4.9, 201, ARRAY[]::text[], now()
		FROM brands b, categories c WHERE b.slug='nuxe' AND c.slug='skincare';
	`)
	if err != nil {
		t.Fatalf("seed gentle-cleanser: %v", err)
	}
}

func newRepo(t *testing.T) (*catalog.Repository, db.Pool, func()) {
	t.Helper()
	url, stop := testsupport.StartPostgres(t)
	testsupport.Migrate(t, url, "../../migrations")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := db.NewPool(ctx, url)
	if err != nil {
		stop()
		t.Fatalf("pool: %v", err)
	}
	seedSmall(t, pool)
	return catalog.NewRepository(pool), pool, func() { pool.Close(); stop() }
}

func TestListCategoriesAndBrands(t *testing.T) {
	repo, _, cleanup := newRepo(t)
	defer cleanup()
	cats, err := repo.ListCategories(context.Background())
	if err != nil {
		t.Fatalf("ListCategories: %v", err)
	}
	if len(cats) != 2 || cats[0].Slug != "skincare" {
		t.Errorf("cats = %+v", cats)
	}
	bs, err := repo.ListBrands(context.Background())
	if err != nil {
		t.Fatalf("ListBrands: %v", err)
	}
	if len(bs) != 2 || bs[0].Name != "Cantu" {
		t.Errorf("brands = %+v", bs)
	}
}

func TestGetProductBySlugFoundAndMissing(t *testing.T) {
	repo, _, cleanup := newRepo(t)
	defer cleanup()
	p, err := repo.GetProductBySlug(context.Background(), "rose-serum")
	if err != nil {
		t.Fatalf("found: %v", err)
	}
	if p.Name != "Rose Serum" || p.PriceGhsMinor != 24500 {
		t.Errorf("product = %+v", p)
	}
	_, err = repo.GetProductBySlug(context.Background(), "does-not-exist")
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("missing: want pgx.ErrNoRows, got %v", err)
	}
}

func TestListProductsFiltersAndSort(t *testing.T) {
	repo, _, cleanup := newRepo(t)
	defer cleanup()
	ctx := context.Background()

	// no filter, default sort (newest) — three products in reverse insert order.
	page, err := repo.ListProducts(ctx, catalog.ListProductsParams{
		Sort: catalog.SortNewest, Limit: 10, Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListProducts: %v", err)
	}
	if page.Total != 3 || len(page.Items) != 3 {
		t.Errorf("total/items = %d / %d", page.Total, len(page.Items))
	}
	if page.Items[0].Slug != "gentle-cleanser" {
		t.Errorf("newest first should be gentle-cleanser, got %s", page.Items[0].Slug)
	}

	// price_asc
	page, _ = repo.ListProducts(ctx, catalog.ListProductsParams{
		Sort: catalog.SortPriceAsc, Limit: 10,
	})
	if page.Items[0].Slug != "curl-cream" {
		t.Errorf("cheapest first should be curl-cream, got %s", page.Items[0].Slug)
	}

	// category filter
	page, _ = repo.ListProducts(ctx, catalog.ListProductsParams{
		CategorySlug: "skincare", Sort: catalog.SortPriceAsc, Limit: 10,
	})
	if page.Total != 2 {
		t.Errorf("skincare total = %d, want 2", page.Total)
	}

	// tag filter
	page, _ = repo.ListProducts(ctx, catalog.ListProductsParams{
		Tag: "Bestseller", Sort: catalog.SortPriceAsc, Limit: 10,
	})
	if page.Total != 2 {
		t.Errorf("bestseller total = %d, want 2", page.Total)
	}

	// search q
	page, _ = repo.ListProducts(ctx, catalog.ListProductsParams{
		Query: "serum", Sort: catalog.SortNewest, Limit: 10,
	})
	if page.Total != 1 || page.Items[0].Slug != "rose-serum" {
		t.Errorf("serum search = %+v", page)
	}
}
