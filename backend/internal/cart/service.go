// Package cart owns guest + authenticated carts. The service wraps the
// repository with at-add-time price snapshots, IDOR-safe item operations,
// and shipping-aware View totals.
package cart

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/catalog"
	sqlcq "github.com/oti-adjei/ruecosmetics/internal/db/sqlc"
	"github.com/oti-adjei/ruecosmetics/internal/logging"
	"github.com/oti-adjei/ruecosmetics/internal/shipping"
)

// Service-level sentinels surfaced to handlers.
var (
	ErrInvalidQty     = errors.New("cart: qty must be >= 1")
	ErrUnknownProduct = errors.New("cart: unknown product")
	ErrItemNotFound   = errors.New("cart: item not found in cart")
)

// Service orchestrates cart reads/writes. Construct via NewService.
type Service struct {
	Repo     *Repository
	Catalog  *catalog.Repository
	Shipping *shipping.Service
	Log      *zap.Logger
	Now      func() time.Time
}

// NewService returns a Service wired with sane defaults (time.Now clock).
func NewService(repo *Repository, cat *catalog.Repository, ship *shipping.Service, log *zap.Logger) *Service {
	return &Service{
		Repo:     repo,
		Catalog:  cat,
		Shipping: ship,
		Log:      log,
		Now:      time.Now,
	}
}

// CartIdentity is what the handler hands to the service after resolving the
// caller. Exactly one field is non-zero in a well-formed call; if both are
// zero the service mints a guest cart.
type CartIdentity struct {
	UserID     uuid.UUID
	GuestToken string
}

func (c CartIdentity) hasUser() bool  { return c.UserID != uuid.Nil }
func (c CartIdentity) hasGuest() bool { return c.GuestToken != "" }

// View is the wire-shape returned to handlers (which adapt it to JSON).
type View struct {
	CartID                        uuid.UUID
	GuestToken                    string
	Items                         []ItemView
	SubtotalGhsMinor              int64
	ShippingCostGhsMinor          int64
	FreeShippingRemainderGhsMinor int64
	TotalGhsMinor                 int64
}

// ItemView denormalises an item with the product fields the cart UI needs.
type ItemView struct {
	ID                uuid.UUID
	ProductID         uuid.UUID
	ProductSlug       string
	ProductName       string
	ProductImagePath  string
	Qty               int32
	UnitPriceGhsMinor int64
	LineTotalGhsMinor int64
}

// GetOrCreate resolves the identity to a cart (creating a guest cart, and
// minting a new token, if the identity is empty), then returns the View and
// the newly-minted guest token (empty string when nothing was minted).
func (s *Service) GetOrCreate(ctx context.Context, id CartIdentity) (View, string, error) {
	log := logging.From(ctx, s.Log)
	cart, minted, err := s.resolveOrCreate(ctx, id)
	if err != nil {
		log.Error("cart: resolve", zap.Error(err))
		return View{}, "", err
	}
	view, err := s.buildView(ctx, cart)
	if err != nil {
		log.Error("cart: build view", zap.Error(err))
		return View{}, "", err
	}
	return view, minted, nil
}

// AddItem snapshots the product's current price into unit_price_ghs_minor on
// first insert; subsequent adds for the same product sum qty and KEEP the
// original snapshot (the SQL upsert never touches unit_price on conflict).
func (s *Service) AddItem(ctx context.Context, id CartIdentity, productID uuid.UUID, qty int32) (View, error) {
	log := logging.From(ctx, s.Log)
	if qty < 1 {
		return View{}, ErrInvalidQty
	}
	product, err := s.Catalog.GetProductByID(ctx, productID)
	if errors.Is(err, catalog.ErrNotFound) {
		return View{}, ErrUnknownProduct
	}
	if err != nil {
		log.Error("cart: lookup product", zap.Error(err))
		return View{}, err
	}
	cart, _, err := s.resolveOrCreate(ctx, id)
	if err != nil {
		return View{}, err
	}
	if _, err := s.Repo.UpsertCartItemAddQty(ctx, cart.ID, productID, qty, product.PriceGhsMinor); err != nil {
		log.Error("cart: upsert item", zap.Error(err))
		return View{}, err
	}
	return s.buildView(ctx, cart)
}

// UpdateQty rewrites a specific item's qty. The item is row-scoped to the
// resolved cart — a cross-cart attempt surfaces as ErrItemNotFound so the
// handler returns 404 (anti-IDOR-enumeration).
func (s *Service) UpdateQty(ctx context.Context, id CartIdentity, itemID uuid.UUID, qty int32) (View, error) {
	log := logging.From(ctx, s.Log)
	if qty < 1 {
		return View{}, ErrInvalidQty
	}
	cart, err := s.resolveExisting(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return View{}, ErrItemNotFound
		}
		return View{}, err
	}
	if err := s.Repo.SetCartItemQty(ctx, itemID, cart.ID, qty); err != nil {
		if errors.Is(err, ErrNotFound) {
			return View{}, ErrItemNotFound
		}
		log.Error("cart: set qty", zap.Error(err))
		return View{}, err
	}
	return s.buildView(ctx, cart)
}

// RemoveItem deletes a specific item from the caller's cart. Row-scoped like
// UpdateQty.
func (s *Service) RemoveItem(ctx context.Context, id CartIdentity, itemID uuid.UUID) (View, error) {
	log := logging.From(ctx, s.Log)
	cart, err := s.resolveExisting(ctx, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return View{}, ErrItemNotFound
		}
		return View{}, err
	}
	if err := s.Repo.DeleteCartItem(ctx, itemID, cart.ID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return View{}, ErrItemNotFound
		}
		log.Error("cart: delete item", zap.Error(err))
		return View{}, err
	}
	return s.buildView(ctx, cart)
}

// ── internal helpers ────────────────────────────────────────────────────────

// resolveOrCreate returns the caller's cart, minting one when necessary.
// The string return is the guest token freshly minted by this call (empty
// when no minting happened).
func (s *Service) resolveOrCreate(ctx context.Context, id CartIdentity) (cartRow, string, error) {
	if id.hasUser() {
		c, err := s.Repo.GetCartByUserID(ctx, id.UserID)
		if err == nil {
			return toCartRow(c), "", nil
		}
		if !errors.Is(err, ErrNotFound) {
			return cartRow{}, "", err
		}
		c, err = s.Repo.CreateCartForUser(ctx, id.UserID)
		if err != nil {
			return cartRow{}, "", err
		}
		return toCartRow(c), "", nil
	}
	if id.hasGuest() {
		c, err := s.Repo.GetCartByGuestToken(ctx, id.GuestToken)
		if err == nil {
			return toCartRow(c), "", nil
		}
		if !errors.Is(err, ErrNotFound) {
			return cartRow{}, "", err
		}
		// Token-in-cookie that doesn't exist on the server — mint a new
		// cart with the SAME token so the cookie keeps working.
		c, err = s.Repo.CreateCartForGuest(ctx, id.GuestToken)
		if err != nil {
			return cartRow{}, "", err
		}
		return toCartRow(c), "", nil
	}
	// No identity → mint a fresh guest token + cart.
	token := uuid.NewString()
	c, err := s.Repo.CreateCartForGuest(ctx, token)
	if err != nil {
		return cartRow{}, "", err
	}
	return toCartRow(c), token, nil
}

// resolveExisting looks up a cart but never creates one. Used by mutating
// endpoints that operate on a specific item_id: if no cart exists, the item
// can't exist either.
func (s *Service) resolveExisting(ctx context.Context, id CartIdentity) (cartRow, error) {
	if id.hasUser() {
		c, err := s.Repo.GetCartByUserID(ctx, id.UserID)
		if err != nil {
			return cartRow{}, err
		}
		return toCartRow(c), nil
	}
	if id.hasGuest() {
		c, err := s.Repo.GetCartByGuestToken(ctx, id.GuestToken)
		if err != nil {
			return cartRow{}, err
		}
		return toCartRow(c), nil
	}
	return cartRow{}, ErrNotFound
}

// cartRow is a lightweight view of the sqlc Cart row stripped of pgtype
// gymnastics so callers don't need to know about pgtype.UUID.
type cartRow struct {
	ID         uuid.UUID
	GuestToken string
}

func toCartRow(c sqlcq.Cart) cartRow {
	r := cartRow{ID: c.ID}
	if c.GuestToken != nil {
		r.GuestToken = *c.GuestToken
	}
	return r
}

// buildView reads all items in the cart, looks up each product for slug/name/
// image, computes line/sub/shipping totals via shipping.Service, and returns
// the assembled View.
func (s *Service) buildView(ctx context.Context, c cartRow) (View, error) {
	items, err := s.Repo.ListCartItems(ctx, c.ID)
	if err != nil {
		return View{}, err
	}
	out := View{
		CartID:     c.ID,
		GuestToken: c.GuestToken,
		Items:      make([]ItemView, 0, len(items)),
	}
	var subtotal int64
	for _, it := range items {
		product, err := s.Catalog.GetProductByID(ctx, it.ProductID)
		if err != nil {
			// Product was deleted while in cart — skip from totals but
			// surface the failure so we don't quietly hide rows.
			if errors.Is(err, catalog.ErrNotFound) {
				continue
			}
			return View{}, err
		}
		line := int64(it.Qty) * it.UnitPriceGhsMinor
		subtotal += line
		out.Items = append(out.Items, ItemView{
			ID:                it.ID,
			ProductID:         it.ProductID,
			ProductSlug:       product.Slug,
			ProductName:       product.Name,
			ProductImagePath:  product.ImagePath,
			Qty:               it.Qty,
			UnitPriceGhsMinor: it.UnitPriceGhsMinor,
			LineTotalGhsMinor: line,
		})
	}
	q := s.Shipping.Quote(subtotal)
	out.SubtotalGhsMinor = subtotal
	if len(out.Items) == 0 {
		// Plan: an empty cart has shipping_cost_ghs_minor: 0 and
		// total_ghs_minor: 0; only free_shipping_remainder reflects
		// the configured free-over threshold.
		out.ShippingCostGhsMinor = 0
		out.FreeShippingRemainderGhsMinor = q.FreeOverGhsMinor
		out.TotalGhsMinor = 0
		return out, nil
	}
	out.ShippingCostGhsMinor = q.AppliedCostGhsMinor
	out.FreeShippingRemainderGhsMinor = q.FreeShippingRemainderGhsMinor
	out.TotalGhsMinor = subtotal + q.AppliedCostGhsMinor
	return out, nil
}
