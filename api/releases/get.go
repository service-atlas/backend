package releases

import (
	"encoding/json"
	"net/http"
	"service-atlas/internal/customerrors"
	"strconv"

	"github.com/service-atlas/go-common/httphelpers"
)

func (s *ServiceCallsHandler) GetReleasesByServiceId(rw http.ResponseWriter, req *http.Request) {
	serviceId, ok := httphelpers.GetGuidFromRequestPath("id", req)
	if !ok {
		http.Error(rw, "Invalid service ID", http.StatusBadRequest)
		return
	}

	page, pageSize, err := validatePageParams(req)
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}

	releases, err := s.Repository.GetReleasesByServiceId(req.Context(), serviceId, page, pageSize)
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}

	// Set content type header
	rw.Header().Set("Content-Type", "application/json")

	// Encode the releases as JSON and write to the response
	err = json.NewEncoder(rw).Encode(releases)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *ServiceCallsHandler) GetReleasesInDateRange(rw http.ResponseWriter, req *http.Request) {
	startDate, ok := httphelpers.GetDateFromRequestPath("startDate", req)
	if !ok {
		http.Error(rw, "Invalid start date", http.StatusBadRequest)
		return
	}
	endDate, ok := httphelpers.GetDateFromRequestPath("endDate", req)
	if !ok {
		http.Error(rw, "Invalid end date", http.StatusBadRequest)
		return
	}
	if endDate.Before(startDate) {
		http.Error(rw, "End date must be after start date", http.StatusBadRequest)
		return
	}

	page, pageSize, err := validatePageParams(req)
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}

	releases, err := s.Repository.GetReleasesInDateRange(req.Context(), startDate, endDate, page, pageSize)
	if err != nil {
		customerrors.HandleError(rw, err)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(rw).Encode(releases)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
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
