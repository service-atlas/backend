package reports

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"service-atlas/internal/customerrors"

	"github.com/service-atlas/go-common/httphelpers"
	"github.com/service-atlas/go-common/httplog"
)

func (c *CallsHandler) GetServicesByTeam(rw http.ResponseWriter, r *http.Request) {
	teamId, ok := httphelpers.GetGuidFromRequestPath("teamId", r)
	logger := httplog.LoggerFromContext(r.Context())
	if !ok {
		http.Error(rw, "Invalid team ID", http.StatusBadRequest)
		return
	}
	services, err := c.repository.GetServicesByTeam(r.Context(), teamId)
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(services)
	if err != nil {
		logger.Debug("Error encoding services json",
			slog.String("error", err.Error()),
		)
	}

}
