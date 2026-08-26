package dependencies

import (
	"encoding/json"
	"net/http"
	"service-atlas/internal/customerrors"
	"service-atlas/repositories"

	"github.com/service-atlas/go-common/httphelpers"
)

func (s *ServiceCallsHandler) CreateDependency(rw http.ResponseWriter, req *http.Request) {
	id, ok := httphelpers.GetGuidFromRequestPath("id", req)
	if !ok {
		http.Error(rw, "path id not valid", http.StatusBadRequest)
		return
	}
	dep := &repositories.Dependency{}
	err := json.NewDecoder(req.Body).Decode(dep)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	if err := dep.Validate(); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	err = s.Repository.AddDependency(req.Context(), id, *dep)

	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}
