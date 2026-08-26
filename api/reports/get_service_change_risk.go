package reports

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"service-atlas/internal/customerrors"

	"github.com/service-atlas/go-common/httphelpers"
	"github.com/service-atlas/go-common/httplog"
)

func (c *CallsHandler) GetServiceChangeRisk(rw http.ResponseWriter, r *http.Request) {
	serviceId, ok := httphelpers.GetGuidFromRequestPath("id", r)
	if !ok {
		http.Error(rw, "Invalid service ID", http.StatusBadRequest)
		return
	}
	report, err := c.repository.GetServiceChangeRisk(r.Context(), serviceId)
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(report)
	if err != nil {
		logger := httplog.LoggerFromContext(r.Context())
		logger.Debug("Error encoding change risk report json",
			slog.String("error", err.Error()),
		)
	}
}
