package reports

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"service-atlas/internal"
	"service-atlas/internal/customerrors"
	"strconv"
	"time"
)

func (c *CallsHandler) GetServicesByTier(rw http.ResponseWriter, r *http.Request) {
	logger := internal.LoggerFromContext(r.Context())
	tierStr := r.URL.Query().Get("tier")
	tier, err := strconv.Atoi(tierStr)
	if err != nil {
		http.Error(rw, "Invalid tier", http.StatusBadRequest)
		return
	}
	if tier < 1 || tier > 4 {
		http.Error(rw, "Invalid tier", http.StatusBadRequest)
		return
	}
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	services, err := c.repository.GetServicesByTier(ctxWithTimeout, tier)
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(services)
	if err != nil {
		logger.Debug("Error encoding services json",
			slog.String("error", err.Error()),
		)
	}
}
