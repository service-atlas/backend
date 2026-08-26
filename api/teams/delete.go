package teams

import (
	"context"
	"net/http"
	"service-atlas/internal/customerrors"
	"time"

	"github.com/service-atlas/go-common/httphelpers"
)

func (c CallsHandler) DeleteTeam(rw http.ResponseWriter, r *http.Request) {
	id, ok := httphelpers.GetGuidFromRequestPath("id", r)
	if !ok {
		http.Error(rw, "Invalid team ID", http.StatusBadRequest)
		return
	}
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	err := c.Repository.DeleteTeam(ctxWithTimeout, id)
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	rw.WriteHeader(http.StatusNoContent)
}
