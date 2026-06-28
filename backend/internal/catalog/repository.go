package catalog

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	sqlcq "github.com/oti-adjei/ruecosmetics/internal/db/sqlc"
)

type Repository struct {
	q *sqlcq.Queries
}

func NewRepository(pool db.Pool) *Repository {
	return &Repository{q: sqlcq.New(pool)}
}

// ErrNotFound is returned when a product lookup fails.
var ErrNotFound = errors.New("catalog: not found")

type ListProductsParams struct {
	CategorySlug string
	BrandSlug    string
	Tag          string
	Query        string
	Sort         SortKey
	Limit        int32
	Offset       int32
}

type ProductsPage struct {
	Items []sqlcq.Product
	Total int64
}

// nstr converts an empty string to nil so the sqlc.narg() guards in the SQL
// collapse to "no filter". sqlc generated *string for nullable text params.
func nstr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (r *Repository) ListCategories(ctx context.Context) ([]sqlcq.Category, error) {
	return r.q.ListCategories(ctx)
}

func (r *Repository) ListBrands(ctx context.Context) ([]sqlcq.Brand, error) {
	return r.q.ListBrands(ctx)
}

func (r *Repository) GetBrandByID(ctx context.Context, id uuid.UUID) (sqlcq.Brand, error) {
	b, err := r.q.GetBrandByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.Brand{}, ErrNotFound
	}
	return b, err
}

func (r *Repository) GetProductBySlug(ctx context.Context, slug string) (sqlcq.Product, error) {
	return r.q.GetProductBySlug(ctx, slug)
}

func (r *Repository) GetProductByID(ctx context.Context, id uuid.UUID) (sqlcq.Product, error) {
	p, err := r.q.GetProductByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return sqlcq.Product{}, ErrNotFound
	}
	return p, err
}

func (r *Repository) ListProducts(ctx context.Context, p ListProductsParams) (ProductsPage, error) {
	countArgs := sqlcq.CountProductsParams{
		CategorySlug: nstr(p.CategorySlug),
		BrandSlug:    nstr(p.BrandSlug),
		Tag:          nstr(p.Tag),
		Q:            nstr(p.Query),
	}
	total, err := r.q.CountProducts(ctx, countArgs)
	if err != nil {
		return ProductsPage{}, err
	}

	var items []sqlcq.Product
	switch p.Sort {
	case SortPriceAsc:
		items, err = r.q.ListProductsByPriceAsc(ctx, sqlcq.ListProductsByPriceAscParams{
			Limit: p.Limit, Offset: p.Offset,
			CategorySlug: countArgs.CategorySlug, BrandSlug: countArgs.BrandSlug, Tag: countArgs.Tag, Q: countArgs.Q,
		})
	case SortPriceDesc:
		items, err = r.q.ListProductsByPriceDesc(ctx, sqlcq.ListProductsByPriceDescParams{
			Limit: p.Limit, Offset: p.Offset,
			CategorySlug: countArgs.CategorySlug, BrandSlug: countArgs.BrandSlug, Tag: countArgs.Tag, Q: countArgs.Q,
		})
	case SortRatingDesc:
		items, err = r.q.ListProductsByRating(ctx, sqlcq.ListProductsByRatingParams{
			Limit: p.Limit, Offset: p.Offset,
			CategorySlug: countArgs.CategorySlug, BrandSlug: countArgs.BrandSlug, Tag: countArgs.Tag, Q: countArgs.Q,
		})
	case SortNameAsc:
		items, err = r.q.ListProductsByName(ctx, sqlcq.ListProductsByNameParams{
			Limit: p.Limit, Offset: p.Offset,
			CategorySlug: countArgs.CategorySlug, BrandSlug: countArgs.BrandSlug, Tag: countArgs.Tag, Q: countArgs.Q,
		})
	default: // SortNewest
		items, err = r.q.ListProductsByNewest(ctx, sqlcq.ListProductsByNewestParams{
			Limit: p.Limit, Offset: p.Offset,
			CategorySlug: countArgs.CategorySlug, BrandSlug: countArgs.BrandSlug, Tag: countArgs.Tag, Q: countArgs.Q,
		})
	}
	return ProductsPage{Items: items, Total: total}, err
}
