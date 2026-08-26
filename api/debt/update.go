package debt

import (
	"context"
	"encoding/json"
	"net/http"
	"service-atlas/internal"
	"service-atlas/internal/customerrors"
	"time"

	"github.com/service-atlas/go-common/httphelpers"
)

func (c CallsHandler) UpdateDebtStatus(rw http.ResponseWriter, r *http.Request) {
	id, ok := httphelpers.GetGuidFromRequestPath("id", r)
	if !ok {
		http.Error(rw, "debt id not valid", http.StatusBadRequest)
		return
	}
	body := map[string]string{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	if _, ok := body["status"]; !ok || !internal.DebtStatus.IsMember(body["status"]) {
		http.Error(rw, "status not valid", http.StatusBadRequest)
		return
	}
	timeout, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	err = c.Repository.UpdateStatus(timeout, id, body["status"])
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	rw.WriteHeader(http.StatusNoContent)
}
