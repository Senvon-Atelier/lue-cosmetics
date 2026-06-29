package orders

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/httpx"
	"github.com/oti-adjei/ruecosmetics/internal/logging"
	"github.com/oti-adjei/ruecosmetics/internal/payments/paystack"
)

// paystackWebhookEvent is the slice of Paystack's webhook payload we care
// about. Extra fields are ignored by the JSON decoder.
type paystackWebhookEvent struct {
	Event string `json:"event"` // e.g., "charge.success"
	Data  struct {
		Reference string `json:"reference"`
		ID        int64  `json:"id"`     // numeric transaction id
		Status    string `json:"status"` // "success", "failed", ...
		Amount    int64  `json:"amount"`
	} `json:"data"`
}

// paystackWebhook godoc
//
// @Summary  Paystack webhook receiver (public, HMAC-verified)
// @Tags     checkout
// @Accept   json
// @Produce  json
// @Success  200
// @Failure  401 {object} httpx.ErrorEnvelope
// @Failure  500 {object} httpx.ErrorEnvelope
// @Router   /webhooks/paystack [post]
func (h *Handlers) paystackWebhook(w http.ResponseWriter, r *http.Request) {
	log := logging.From(r.Context(), h.Log)

	// Read the RAW body BEFORE decoding — the HMAC is computed over the exact
	// bytes Paystack signed. Any JSON round-trip would mutate them.
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("webhook: read body", zap.Error(err))
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "webhook read failed", nil)
		return
	}

	signature := r.Header.Get("x-paystack-signature")
	if !paystack.VerifyWebhookSignature(h.PaystackSecret, rawBody, signature) {
		// Paystack will retry on 401; this is the right code per plan §29.
		httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthorized, "invalid signature", nil)
		return
	}

	var event paystackWebhookEvent
	if err := json.Unmarshal(rawBody, &event); err != nil {
		log.Warn("webhook: malformed body", zap.Error(err))
		// Acknowledge so Paystack stops retrying a permanently-bad payload.
		w.WriteHeader(http.StatusOK)
		return
	}

	// Filter: only process charge.success / data.status=success. Other events
	// are acknowledged but produce no state change.
	if event.Event != "charge.success" || event.Data.Status != "success" {
		log.Info("webhook: non-success event ignored",
			zap.String("event", event.Event),
			zap.String("status", event.Data.Status),
			zap.String("reference", event.Data.Reference))
		w.WriteHeader(http.StatusOK)
		return
	}

	txID := strconv.FormatInt(event.Data.ID, 10)
	if err := h.Svc.MarkPaid(r.Context(), event.Data.Reference, txID); err != nil {
		if errors.Is(err, ErrNotFound) {
			// Unknown reference — could be a clock-skewed test or a misrouted
			// webhook. Acknowledge so Paystack doesn't retry forever.
			log.Info("webhook: unknown reference, idempotent ack",
				zap.String("reference", event.Data.Reference))
			w.WriteHeader(http.StatusOK)
			return
		}
		log.Error("webhook: mark paid",
			zap.String("reference", event.Data.Reference),
			zap.Error(err))
		// 500 → Paystack will retry.
		httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternal, "webhook processing failed", nil)
		return
	}
	w.WriteHeader(http.StatusOK)
}
