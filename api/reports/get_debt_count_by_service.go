package reports

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"service-atlas/internal/customerrors"
	"time"

	"github.com/service-atlas/go-common/httplog"
)

func (c *CallsHandler) GetServiceDebtReport(rw http.ResponseWriter, req *http.Request) {
	ctxWithTimeout, cancel := context.WithTimeout(req.Context(), 10*time.Second)
	defer cancel()
	report, err := c.repository.GetDebtCountByService(ctxWithTimeout)
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(rw)
	err = encoder.Encode(report)
	if err != nil {
		logger := httplog.LoggerFromContext(req.Context())
		logger.Debug("Error encoding debt report json",
			slog.String("error", err.Error()),
		)
	}
}
