package releases

import (
	"encoding/json"
	"net/http"
	"service-atlas/internal/customerrors"
	"service-atlas/repositories"

	"github.com/service-atlas/go-common/httphelpers"
)

func (s *ServiceCallsHandler) CreateRelease(rw http.ResponseWriter, req *http.Request) {
	serviceId, ok := httphelpers.GetGuidFromRequestPath("id", req)
	if !ok {
		http.Error(rw, "Invalid service ID", http.StatusBadRequest)
		return
	}

	r := &repositories.Release{}
	const maxBodySize = 1 << 20 // 1 MB
	req.Body = http.MaxBytesReader(rw, req.Body, maxBodySize)
	err := json.NewDecoder(req.Body).Decode(r)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}

	// Set the service ID from the path parameter
	r.ServiceId = serviceId

	if err = r.Validate(); err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	err = s.Repository.CreateRelease(req.Context(), *r)
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	rw.WriteHeader(http.StatusCreated)
}
