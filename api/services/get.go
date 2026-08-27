package services

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"service-atlas/internal/customerrors"
	"strconv"
	"time"

	"github.com/service-atlas/go-common/httphelpers"
	"github.com/service-atlas/go-common/httplog"
)

func (u *ServiceCallsHandler) GetAllServices(rw http.ResponseWriter, r *http.Request) {
	logger := httplog.LoggerFromContext(r.Context())
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate that page is positive
	if page < 1 {
		http.Error(rw, "page must be positive", http.StatusBadRequest)
		return
	}
	pageSize, err := strconv.Atoi(r.URL.Query().Get("pageSize"))
	if err != nil {
		pageSize = 10
	}

	// Validate that pageSize is between 1 and 100
	if pageSize < 1 || pageSize > 100 {
		http.Error(rw, "pageSize must be between 1 and 100", http.StatusBadRequest)
		return
	}
	services, err := u.Repository.GetAllServices(r.Context(), page, pageSize)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(services)
	if err != nil {
		logger.Debug("Error encoding services json",
			slog.String("error", err.Error()))
	}
}

func (u *ServiceCallsHandler) GetById(rw http.ResponseWriter, r *http.Request) {
	logger := httplog.LoggerFromContext(r.Context())
	id, ok := httphelpers.GetGuidFromRequestPath("id", r)

	if !ok {
		http.Error(rw, "Service id is required", http.StatusBadRequest)
		return
	}

	service, err := u.Repository.GetServiceById(r.Context(), id)
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}

	// Check if service was found (Id will be empty if not found)
	if service.Id == "" {
		http.Error(rw, "Service not found", http.StatusNotFound)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(service)
	if err != nil {
		logger.Debug("Error encoding service json",
			slog.String("error", err.Error()))
	}
}

func (u *ServiceCallsHandler) GetTeamsByServiceId(rw http.ResponseWriter, req *http.Request) {
	serviceId, ok := httphelpers.GetGuidFromRequestPath("id", req)
	if !ok {
		http.Error(rw, "Invalid service ID", http.StatusBadRequest)
		return
	}
	ctxWithTimeout, cancel := context.WithTimeout(req.Context(), 10*time.Second)
	defer cancel()
	teams, err := u.Repository.GetTeamsByServiceId(ctxWithTimeout, serviceId)
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(teams)
	if err != nil {
		httplog.LoggerFromContext(req.Context()).Debug("Error encoding teams json",
			slog.String("error", err.Error()))
	}
}
