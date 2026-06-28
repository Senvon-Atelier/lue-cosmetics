package health

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/oti-adjei/ruecosmetics/internal/app"
	"github.com/oti-adjei/ruecosmetics/internal/httpx"
	"github.com/oti-adjei/ruecosmetics/internal/logging"
)

type response struct {
	Status string `json:"status"`
	DB     string `json:"db"`
}

// Handler godoc
//
// @Summary      Health check
// @Description  Verifies the service is up and the database is reachable.
// @Tags         meta
// @Produce      json
// @Success      200 {object} health.response
// @Failure      503 {object} httpx.ErrorEnvelope
// @Router       /healthz [get]
func Handler(a *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := a.Pool.Ping(ctx); err != nil {
			logging.From(ctx, a.Logger).Warn("healthz: db ping failed", zap.Error(err))
			httpx.WriteError(w, http.StatusServiceUnavailable, httpx.CodeInternal, "db unavailable", nil)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, response{Status: "ok", DB: "ok"})
	}
}
