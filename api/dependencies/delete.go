package dependencies

import (
	"net/http"
	"service-atlas/internal/customerrors"

	"github.com/service-atlas/go-common/httphelpers"
)

func (s *ServiceCallsHandler) DeleteDependency(rw http.ResponseWriter, req *http.Request) {
	id, ok := httphelpers.GetGuidFromRequestPath("id", req)
	if !ok {
		http.Error(rw, "path id not valid", http.StatusBadRequest)
		return
	}
	dependsOnID, ok := httphelpers.GetGuidFromRequestPath("id2", req)
	if !ok {
		http.Error(rw, "path id2 not valid", http.StatusBadRequest)
		return
	}
	err := s.Repository.DeleteDependency(req.Context(), id, dependsOnID)

	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	rw.WriteHeader(http.StatusNoContent)
}
