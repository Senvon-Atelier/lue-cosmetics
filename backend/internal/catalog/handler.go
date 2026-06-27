package catalog

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	sqlcq "github.com/oti-adjei/ruecosmetics/internal/db/sqlc"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
)

const (
	defaultLimit = 24
	maxLimit     = 100
)

// Handlers holds catalog HTTP handlers.
type Handlers struct {
	repo *Repository
}

// NewHandlers creates a new Handlers wired to the given Repository.
func NewHandlers(repo *Repository) *Handlers {
	return &Handlers{repo: repo}
}

// Mount registers all catalog routes on the given router.
func (h *Handlers) Mount(r chi.Router) {
	r.Get("/categories", h.listCategories)
	r.Get("/brands", h.listBrands)
	r.Get("/products", h.listProducts)
	r.Get("/products/{slug}", h.getProductBySlug)
}

// listCategories godoc
//
// @Summary  List categories
// @Tags     catalog
// @Produce  json
// @Success  200 {array} sqlc.Category
// @Failure  500 {object} httpx.ErrorEnvelope
// @Router   /categories [get]
func (h *Handlers) listCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.repo.ListCategories(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to list categories", nil)
		return
	}
	if cats == nil {
		cats = []sqlcq.Category{}
	}
	httpx.WriteJSON(w, http.StatusOK, cats)
}

// listBrands godoc
//
// @Summary  List brands
// @Tags     catalog
// @Produce  json
// @Success  200 {array} sqlc.Brand
// @Failure  500 {object} httpx.ErrorEnvelope
// @Router   /brands [get]
func (h *Handlers) listBrands(w http.ResponseWriter, r *http.Request) {
	bs, err := h.repo.ListBrands(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to list brands", nil)
		return
	}
	if bs == nil {
		bs = []sqlcq.Brand{}
	}
	httpx.WriteJSON(w, http.StatusOK, bs)
}

type productsResponse struct {
	Items []sqlcProductView `json:"items"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
	Total int64             `json:"total"`
}

// listProducts godoc
//
// @Summary  List products
// @Tags     catalog
// @Produce  json
// @Param    category query string false "Category slug"
// @Param    brand    query string false "Brand slug"
// @Param    tag      query string false "Tag"
// @Param    q        query string false "Search query against name"
// @Param    sort     query string false "newest|price_asc|price_desc|rating_desc|name_asc"
// @Param    page     query int    false "Page (1-based)"
// @Param    limit    query int    false "Page size (default 24, max 100)"
// @Success  200 {object} productsResponse
// @Failure  400 {object} httpx.ErrorEnvelope
// @Failure  500 {object} httpx.ErrorEnvelope
// @Router   /products [get]
func (h *Handlers) listProducts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	sort, err := ParseSort(q.Get("sort"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid sort", map[string]string{"sort": err.Error()})
		return
	}
	page := 1
	if v := q.Get("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid page", map[string]string{"page": "must be a positive integer"})
			return
		}
		page = n
	}
	limit := defaultLimit
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			httpx.WriteError(w, http.StatusBadRequest, httpx.CodeValidation, "invalid limit", map[string]string{"limit": "must be a positive integer"})
			return
		}
		if n > maxLimit {
			n = maxLimit
		}
		limit = n
	}
	offset := (page - 1) * limit

	pageOut, err := h.repo.ListProducts(r.Context(), ListProductsParams{
		CategorySlug: q.Get("category"),
		BrandSlug:    q.Get("brand"),
		Tag:          q.Get("tag"),
		Query:        q.Get("q"),
		Sort:         sort,
		Limit:        int32(limit),
		Offset:       int32(offset),
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to list products", nil)
		return
	}
	views := make([]sqlcProductView, 0, len(pageOut.Items))
	for _, p := range pageOut.Items {
		views = append(views, productViewFromSqlc(p))
	}
	httpx.WriteJSON(w, http.StatusOK, productsResponse{
		Items: views, Page: page, Limit: limit, Total: pageOut.Total,
	})
}

// getProductBySlug godoc
//
// @Summary  Get product by slug
// @Tags     catalog
// @Produce  json
// @Param    slug path string true "Slug"
// @Success  200 {object} sqlcProductView
// @Failure  404 {object} httpx.ErrorEnvelope
// @Failure  500 {object} httpx.ErrorEnvelope
// @Router   /products/{slug} [get]
func (h *Handlers) getProductBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	p, err := h.repo.GetProductBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpx.WriteError(w, http.StatusNotFound, httpx.CodeNotFound, "product not found", nil)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "failed to get product", nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, productViewFromSqlc(p))
}

type sqlcProductView struct {
	ID               string   `json:"id"`
	Slug             string   `json:"slug"`
	Name             string   `json:"name"`
	BrandID          string   `json:"brand_id"`
	CategoryID       string   `json:"category_id"`
	PriceGhsMinor    int64    `json:"price_ghs_minor"`
	WasPriceGhsMinor *int64   `json:"was_price_ghs_minor,omitempty"`
	Tone             string   `json:"tone"`
	Size             string   `json:"size"`
	Rating           *float64 `json:"rating,omitempty"`
	ReviewCount      int32    `json:"review_count"`
	Tags             []string `json:"tags"`
	ImagePath        string   `json:"image_path"`
	CreatedAt        string   `json:"created_at"`
}

func productViewFromSqlc(p sqlcq.Product) sqlcProductView {
	v := sqlcProductView{
		ID:            p.ID.String(),
		Slug:          p.Slug,
		Name:          p.Name,
		BrandID:       p.BrandID.String(),
		CategoryID:    p.CategoryID.String(),
		PriceGhsMinor: p.PriceGhsMinor,
		Tone:          p.Tone,
		Size:          p.Size,
		ReviewCount:   p.ReviewCount,
		Tags:          p.Tags,
		ImagePath:     p.ImagePath,
		CreatedAt:     p.CreatedAt.Time.Format(time.RFC3339),
	}
	if p.WasPriceGhsMinor != nil {
		v.WasPriceGhsMinor = p.WasPriceGhsMinor
	}
	if p.Rating.Valid {
		f, err := p.Rating.Float64Value()
		if err == nil {
			ff := f.Float64
			v.Rating = &ff
		}
	}
	return v
}
