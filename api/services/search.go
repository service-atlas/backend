package services

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"service-atlas/internal/customerrors"

	"github.com/service-atlas/go-common/httplog"
)

func (u *ServiceCallsHandler) Search(rw http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(rw, "query parameter is required", http.StatusBadRequest)
		return
	}
	logger := httplog.LoggerFromContext(r.Context())
	services, err := u.Repository.Search(r.Context(), query)
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(services)
	if err != nil {
		logger.Debug("Error encoding services json",
			slog.String("error", err.Error()))
	}
}
