package debt

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"service-atlas/internal/customerrors"
	"strconv"
	"time"

	"github.com/service-atlas/go-common/httphelpers"
	"github.com/service-atlas/go-common/httplog"
)

func (c CallsHandler) GetDebtByServiceId(rw http.ResponseWriter, r *http.Request) {
	logger := httplog.LoggerFromContext(r.Context())
	id, ok := httphelpers.GetGuidFromRequestPath("id", r)
	if !ok {
		http.Error(rw, "service id not valid", http.StatusBadRequest)
		return
	}
	page, pageSize, err := validatePageParams(r)
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	onlyResolved := r.URL.Query().Get("onlyResolved")
	ctxWithTimeout, cancel := context.WithTimeoutCause(r.Context(), 10*time.Second, errors.New("database timeout"))
	defer cancel()
	debt, err := c.Repository.GetDebtByServiceId(ctxWithTimeout, id, page, pageSize, onlyResolved == "true")
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(debt)
	if err != nil {
		logger.Debug("Error encoding response:",
			slog.String("error", err.Error()))
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func validatePageParams(req *http.Request) (int, int, error) {

	// Parse page and page size from query parameters
	pageStr := req.URL.Query().Get("page")
	pageSizeStr := req.URL.Query().Get("pageSize")

	// Default values
	page := 1
	pageSize := 25

	// Parse page if provided
	if pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err != nil || parsedPage <= 0 {
			return 0, 0, &customerrors.HTTPError{
				Status: http.StatusBadRequest,
				Msg:    "Invalid page parameter",
			}
		}
		page = parsedPage
	}

	// Parse page size if provided
	if pageSizeStr != "" {
		parsedPageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || parsedPageSize <= 0 {
			return 0, 0, &customerrors.HTTPError{
				Status: http.StatusBadRequest,
				Msg:    "Invalid pageSize parameter",
			}
		}
		pageSize = parsedPageSize
	}

	return page, pageSize, nil
}
