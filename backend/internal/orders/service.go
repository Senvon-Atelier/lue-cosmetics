// Package orders provides the OrderService (InitCheckout / MarkPaid /
// VerifyCheckout) that converges the Paystack-hosted-payment redirect flow
// and the public webhook on a single idempotent MarkPaid.
package orders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/cart"
	"github.com/oti-adjei/ruecosmetics/internal/catalog"
	"github.com/oti-adjei/ruecosmetics/internal/db"
	sqlcq "github.com/oti-adjei/ruecosmetics/internal/db/sqlc"
	"github.com/oti-adjei/ruecosmetics/internal/email"
	"github.com/oti-adjei/ruecosmetics/internal/payments/paystack"
	"github.com/oti-adjei/ruecosmetics/internal/shipping"
)

type Service struct {
	Repo        *Repository
	Cart        *cart.Service
	Catalog     *catalog.Repository
	Shipping    *shipping.Service
	Paystack    *paystack.Client
	Email       email.Sender
	Pool        db.Pool
	Log         *zap.Logger
	Now         func() time.Time
	CallbackURL string
}

func NewService(repo *Repository, cartSvc *cart.Service, cat *catalog.Repository, ship *shipping.Service,
	pay *paystack.Client, mail email.Sender, pool db.Pool, log *zap.Logger,
	callbackURL string) *Service {
	return &Service{
		Repo:        repo,
		Cart:        cartSvc,
		Catalog:     cat,
		Shipping:    ship,
		Paystack:    pay,
		Email:       mail,
		Pool:        pool,
		Log:         log,
		Now:         time.Now,
		CallbackURL: callbackURL,
	}
}

type ShippingAddress struct {
	Line1  string `json:"line1"`
	Line2  string `json:"line2,omitempty"`
	City   string `json:"city"`
	Region string `json:"region"`
	Phone  string `json:"phone"`
	Label  string `json:"label,omitempty"`
}

type InitCheckoutInput struct {
	UserID          uuid.UUID
	UserEmail       string
	UserName        string
	ShippingAddress ShippingAddress
	ShippingMethod  string
}

type InitCheckoutOutput struct {
	OrderID          uuid.UUID
	Reference        string
	AuthorizationURL string
	TotalGhsMinor    int64
}

var (
	ErrEmptyCart        = errors.New("orders: cart is empty")
	ErrInvalidAddress   = errors.New("orders: invalid shipping address")
	ErrPaystackNotReady = errors.New("orders: paystack not configured")
)

// InitCheckout validates the shipping address, re-prices the user's cart from
// the catalog (the cart's unit_price snapshot is UI-only — checkout never
// trusts it), persists a pending order + items, and asks Paystack to
// initialize the transaction. The cart is NOT cleared here; cart_items are
// cleared only on successful MarkPaid.
func (s *Service) InitCheckout(ctx context.Context, in InitCheckoutInput) (InitCheckoutOutput, error) {
	if err := validateAddress(in.ShippingAddress); err != nil {
		return InitCheckoutOutput{}, err
	}
	if !s.Paystack.IsConfigured() {
		return InitCheckoutOutput{}, ErrPaystackNotReady
	}

	cartView, _, err := s.Cart.GetOrCreate(ctx, cart.CartIdentity{UserID: in.UserID})
	if err != nil {
		return InitCheckoutOutput{}, err
	}
	if len(cartView.Items) == 0 {
		return InitCheckoutOutput{}, ErrEmptyCart
	}

	type pricedItem struct {
		ProductID    uuid.UUID
		Qty          int32
		UnitPrice    int64
		ProductName  string
		ProductBrand string
		ProductImage string
	}
	priced := make([]pricedItem, 0, len(cartView.Items))
	var subtotal int64
	for _, it := range cartView.Items {
		prod, perr := s.Catalog.GetProductByID(ctx, it.ProductID)
		if errors.Is(perr, catalog.ErrNotFound) {
			return InitCheckoutOutput{}, fmt.Errorf("orders: product %s not found", it.ProductID)
		}
		if perr != nil {
			return InitCheckoutOutput{}, perr
		}
		brandName := ""
		if brand, berr := s.Catalog.GetBrandByID(ctx, prod.BrandID); berr == nil {
			brandName = brand.Name
		} else if !errors.Is(berr, catalog.ErrNotFound) {
			return InitCheckoutOutput{}, berr
		}
		priced = append(priced, pricedItem{
			ProductID:    prod.ID,
			Qty:          it.Qty,
			UnitPrice:    prod.PriceGhsMinor,
			ProductName:  prod.Name,
			ProductBrand: brandName,
			ProductImage: prod.ImagePath,
		})
		subtotal += int64(it.Qty) * prod.PriceGhsMinor
	}

	shipQuote := s.Shipping.Quote(subtotal)
	total := subtotal + shipQuote.AppliedCostGhsMinor

	reference, err := GenerateReference()
	if err != nil {
		return InitCheckoutOutput{}, err
	}
	addrJSON, err := json.Marshal(in.ShippingAddress)
	if err != nil {
		return InitCheckoutOutput{}, err
	}

	var orderID uuid.UUID
	err = db.WithTx(ctx, s.Pool, func(tx pgx.Tx) error {
		q := sqlcq.New(tx)
		order, qerr := q.CreateOrder(ctx, sqlcq.CreateOrderParams{
			UserID:            in.UserID,
			SubtotalGhsMinor:  subtotal,
			ShippingGhsMinor:  shipQuote.AppliedCostGhsMinor,
			TotalGhsMinor:     total,
			PaystackReference: reference,
			ShippingAddress:   addrJSON,
		})
		if qerr != nil {
			return qerr
		}
		for _, p := range priced {
			if _, ierr := q.CreateOrderItem(ctx, sqlcq.CreateOrderItemParams{
				OrderID:              order.ID,
				ProductID:            p.ProductID,
				Qty:                  p.Qty,
				UnitPriceGhsMinor:    p.UnitPrice,
				ProductNameSnapshot:  p.ProductName,
				ProductBrandSnapshot: p.ProductBrand,
				ProductImageSnapshot: p.ProductImage,
			}); ierr != nil {
				return ierr
			}
		}
		orderID = order.ID
		return nil
	})
	if err != nil {
		return InitCheckoutOutput{}, err
	}

	// Paystack init happens OUTSIDE the tx. If it fails, the order row remains
	// pending and the user can retry; we do not roll back the local write.
	psOut, perr := s.Paystack.InitializeTransaction(ctx, paystack.InitializeTransactionInput{
		Email:     in.UserEmail,
		Amount:    total,
		Reference: reference,
		Callback:  s.CallbackURL,
		Currency:  "GHS",
	})
	if perr != nil {
		return InitCheckoutOutput{}, perr
	}
	return InitCheckoutOutput{
		OrderID:          orderID,
		Reference:        reference,
		AuthorizationURL: psOut.AuthorizationURL,
		TotalGhsMinor:    total,
	}, nil
}

// MarkPaid is the single convergence point for the webhook and the verify-poll
// path. Inside one tx: lock the order row, ignore non-pending rows
// idempotently, flip status to paid, delete the user's cart_items, and read
// items+user for the post-commit confirmation email. The email is best-effort
// and never rolls back the order.
func (s *Service) MarkPaid(ctx context.Context, reference, paystackTransactionID string) error {
	var userIDForEmail uuid.UUID
	var sendEmailAfterCommit *emailPayload

	err := db.WithTx(ctx, s.Pool, func(tx pgx.Tx) error {
		q := sqlcq.New(tx)
		order, qerr := q.GetOrderByReferenceForUpdate(ctx, reference)
		if errors.Is(qerr, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if qerr != nil {
			return qerr
		}
		if order.Status != "pending" {
			return nil
		}
		if uerr := q.MarkOrderPaid(ctx, sqlcq.MarkOrderPaidParams{
			ID:                    order.ID,
			PaystackTransactionID: pgxNullable(paystackTransactionID),
		}); uerr != nil {
			return uerr
		}
		if derr := q.DeleteCartItemsByUserID(ctx, pgtype.UUID{Bytes: order.UserID, Valid: true}); derr != nil {
			return derr
		}
		items, ierr := q.ListOrderItems(ctx, order.ID)
		if ierr != nil {
			return ierr
		}
		user, uerr := q.GetUserByID(ctx, order.UserID)
		if uerr != nil {
			return uerr
		}
		// Re-read the order so the email payload reflects the post-update row
		// (status=paid, paystack_transaction_id populated). Cheap and avoids
		// drift between display and DB.
		updated, gerr := q.GetOrderByID(ctx, order.ID)
		if gerr != nil {
			return gerr
		}
		userIDForEmail = user.ID
		sendEmailAfterCommit = buildOrderEmailPayload(user, updated, items)
		return nil
	})
	if err != nil {
		return err
	}
	if sendEmailAfterCommit != nil {
		if sendErr := s.Email.Send(ctx, sendEmailAfterCommit.To,
			"order_confirmation", sendEmailAfterCommit.Data); sendErr != nil {
			s.Log.Error("order confirmation send failed",
				zap.String("user_id", userIDForEmail.String()),
				zap.String("reference", reference),
				zap.Error(sendErr))
			// Best-effort: do not surface the error.
		}
	}
	return nil
}

// VerifyCheckout is the helper called by the auth-gated verify-poll endpoint.
// If the local order is already in a terminal state, return it. Otherwise ask
// Paystack and, on success, call MarkPaid (which is idempotent — the webhook
// may have already arrived). Returns the order's resulting status.
func (s *Service) VerifyCheckout(ctx context.Context, reference string) (string, error) {
	order, err := s.Repo.GetOrderByReference(ctx, reference)
	if errors.Is(err, ErrNotFound) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if order.Status != "pending" {
		return order.Status, nil
	}
	if !s.Paystack.IsConfigured() {
		return "", ErrPaystackNotReady
	}
	res, err := s.Paystack.VerifyTransaction(ctx, reference)
	if err != nil {
		return "", err
	}
	if res.Status != "success" {
		return "pending", nil
	}
	txID := strconv.FormatInt(res.TransactionID, 10)
	if err := s.MarkPaid(ctx, reference, txID); err != nil {
		return "", err
	}
	return "paid", nil
}

func validateAddress(a ShippingAddress) error {
	if strings.TrimSpace(a.Line1) == "" ||
		strings.TrimSpace(a.City) == "" ||
		strings.TrimSpace(a.Region) == "" ||
		strings.TrimSpace(a.Phone) == "" {
		return ErrInvalidAddress
	}
	return nil
}

// pgxNullable converts a string to the sqlc-emitted nullable type for the
// orders.paystack_transaction_id column. Sqlc is configured with
// emit_pointers_for_null_types, so the type is *string.
func pgxNullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

type emailPayload struct {
	To   string
	Data map[string]any
}

func buildOrderEmailPayload(user sqlcq.User, order sqlcq.Order, items []sqlcq.OrderItem) *emailPayload {
	itemMaps := make([]map[string]any, 0, len(items))
	for _, it := range items {
		itemMaps = append(itemMaps, map[string]any{
			"name":       it.ProductNameSnapshot,
			"brand":      it.ProductBrandSnapshot,
			"qty":        it.Qty,
			"line_total": formatGHS(int64(it.Qty) * it.UnitPriceGhsMinor),
		})
	}
	return &emailPayload{
		To: user.Email,
		Data: map[string]any{
			"name":               user.Name,
			"paystack_reference": order.PaystackReference,
			"items":              itemMaps,
			"subtotal":           formatGHS(order.SubtotalGhsMinor),
			"shipping":           formatGHS(order.ShippingGhsMinor),
			"total":              formatGHS(order.TotalGhsMinor),
		},
	}
}

func formatGHS(pesewas int64) string {
	if pesewas < 0 {
		pesewas = -pesewas
		return "-" + fmt.Sprintf("%d.%02d", pesewas/100, pesewas%100)
	}
	return fmt.Sprintf("%d.%02d", pesewas/100, pesewas%100)
}
