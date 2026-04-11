package dependencies

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"service-atlas/internal"
	"service-atlas/internal/customerrors"
	"service-atlas/repositories"
)

func (s *ServiceCallsHandler) GetDependencies(rw http.ResponseWriter, req *http.Request) {
	id, ok := internal.GetGuidFromRequestPath("id", req)

	if !ok {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	interaction_type := req.URL.Query().Get("interaction_type")
	if !internal.InteractionType.IsMember(interaction_type) && interaction_type != "" {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	logger := internal.LoggerFromContext(req.Context())
	logger.Debug("Getting dependencies",
		slog.String("service_id", id),
		slog.String("interaction_type", interaction_type),
	)
	var dep []*repositories.Dependency
	var err error
	if interaction_type != "" {
		dep, err = s.Repository.GetDependenciesByInteractionType(req.Context(), id, interaction_type)
	} else {
		dep, err = s.Repository.GetDependencies(req.Context(), id)
	}
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(dep)
	if err != nil {
		logger := internal.LoggerFromContext(req.Context())
		logger.Debug("Error encoding dependencies json",
			slog.String("error", err.Error()),
		)
	}
}

func (s *ServiceCallsHandler) GetDependents(rw http.ResponseWriter, req *http.Request) {
	id, ok := internal.GetGuidFromRequestPath("id", req)
	if !ok {
		http.Error(rw, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	deps, err := s.Repository.GetDependents(req.Context(), id)
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	ver := req.URL.Query().Get("version")
	returnObj := make([]*repositories.Dependency, 0)
	if ver != "" {
		for _, dep := range deps {
			if ver == dep.Version {
				returnObj = append(returnObj, dep)
			}
		}
	} else {
		returnObj = append(returnObj, deps...)
	}

	err = json.NewEncoder(rw).Encode(returnObj)
	if err != nil {
		logger := internal.LoggerFromContext(req.Context())
		logger.Debug("Error encoding dependencies json",
			slog.String("error", err.Error()),
		)
	}
}
