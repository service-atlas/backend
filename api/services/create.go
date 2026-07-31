package services

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"service-atlas/internal"
	"service-atlas/repositories"
	"time"
)

func (u *ServiceCallsHandler) CreateService(rw http.ResponseWriter, r *http.Request) {
	logger := internal.LoggerFromContext(r.Context())
	createServiceRequest := &repositories.Service{}
	err := json.NewDecoder(r.Body).Decode(createServiceRequest)
	if err != nil {
		// return HTTP 400 bad request
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	err = createServiceRequest.Validate()
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	id, err := u.Repository.CreateService(ctxWithTimeout, *createServiceRequest)

	if err != nil {
		logger.Error("Error creating service:",
			slog.String("error", err.Error()))
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json; charset=UTF-8")
	rw.WriteHeader(http.StatusCreated)
	createServiceRequest.Id = id
	createServiceRequest.Created = time.Now().UTC()
	createServiceRequest.Updated = createServiceRequest.Created
	err = json.NewEncoder(rw).Encode(createServiceRequest)
	if err != nil {
		logger.Error("Error encoding response:",
			slog.String("error", err.Error()))
	}
}
