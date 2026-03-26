package reports

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"service-atlas/internal"
	"service-atlas/internal/customerrors"
	"time"
)

func (c *CallsHandler) GetServiceTypes(rw http.ResponseWriter, r *http.Request) {
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	logger := internal.LoggerFromContext(r.Context())
	defer cancel()
	serviceTypes, err := c.repository.GetServiceTypes(ctxWithTimeout)
	if err != nil {
		logger.Error("Error getting service types:", slog.String("error", err.Error()))
		customerrors.HandleError(rw, err)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(serviceTypes)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
	}
}
