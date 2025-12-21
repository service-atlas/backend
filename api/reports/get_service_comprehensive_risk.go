package reports

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"service-atlas/internal"
	"service-atlas/internal/customerrors"
	"service-atlas/repositories"
	"time"

	"golang.org/x/sync/errgroup"
)

func (c *CallsHandler) GetComprehensiveRiskReport(rw http.ResponseWriter, r *http.Request) {
	serviceId, ok := internal.GetGuidFromRequestPath("id", r)
	ctxWithTimeout, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	if !ok {
		http.Error(rw, "Invalid service ID", http.StatusBadRequest)
		return
	}
	var wg errgroup.Group
	report := repositories.ComprehensiveServiceRisk{}
	wg.Go(func() error {
		r, err := c.repository.GetServiceRiskReport(ctxWithTimeout, serviceId)
		if err != nil {
			return err
		}
		report.HealthRisk = r
		return nil
	})

	wg.Go(func() error {
		r, err := c.repository.GetServiceChangeRisk(ctxWithTimeout, serviceId)
		if err != nil {
			return err
		}
		report.ChangeRisk = r
		return nil
	})

	err := wg.Wait()
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(report)
	if err != nil {
		logger := internal.LoggerFromContext(r.Context())
		logger.Debug("Error encoding change risk report json",
			slog.String("error", err.Error()),
		)
	}

}
