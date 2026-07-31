package services

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"service-atlas/internal"
	"service-atlas/internal/customerrors"
	"service-atlas/repositories"
	"time"
)

func (u *ServiceCallsHandler) UpdateService(rw http.ResponseWriter, r *http.Request) {
	logger := internal.LoggerFromContext(r.Context())
	updateServiceRequest := &repositories.Service{}
	err := json.NewDecoder(r.Body).Decode(updateServiceRequest)
	if err != nil {
		logger.Error("Error decoding request body:",
			slog.String("error", err.Error()))
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	if id, ok := internal.GetGuidFromRequestPath("id", r); !ok || updateServiceRequest.Id != id {
		http.Error(rw, "Service Id is not valid", http.StatusBadRequest)
		return
	}
	err = updateServiceRequest.Validate()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	err = u.Repository.UpdateService(r.Context(), *updateServiceRequest)

	if err != nil {
		logger.Debug("Error updating service:",
			slog.String("error", err.Error()))
		customerrors.HandleError(rw, err)
		return
	}
	updateServiceRequest.Updated = time.Now().UTC()
	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
	rw.WriteHeader(http.StatusOK)
	err = json.NewEncoder(rw).Encode(updateServiceRequest)
	if err != nil {
		logger.Error("Error encoding response:",
			slog.String("error", err.Error()))
	}
}
