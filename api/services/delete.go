package services

import (
	"context"
	"log/slog"
	"net/http"
	"service-atlas/internal/customerrors"
	"time"

	"github.com/service-atlas/go-common/httphelpers"
	"github.com/service-atlas/go-common/httplog"
)

func (u *ServiceCallsHandler) DeleteServiceById(rw http.ResponseWriter, r *http.Request) {
	logger := httplog.LoggerFromContext(r.Context())
	id, ok := httphelpers.GetGuidFromRequestPath("id", r)
	logger.Debug("Request received - DeleteServiceById - " + id)
	if !ok {
		http.Error(rw, "Invalid Request", http.StatusBadRequest)
		logger.Debug("Invalid Request - " + id)
		return
	}
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	err := u.Repository.DeleteService(ctxWithTimeout, id)
	if err != nil {
		logger.Debug("Error deleting service:",
			slog.String("error", err.Error()))
		customerrors.HandleError(rw, err)
		return
	}
	rw.WriteHeader(http.StatusNoContent)

}
