package reports

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"service-atlas/internal/customerrors"
	"service-atlas/repositories"
	"testing"
)

func TestGetServiceRiskReportSuccess(t *testing.T) {
	// Create a handler with mocked dependencies
	validServiceId := "123e4567-e89b-12d3-a456-426614174000" // Valid GUID
	mockReport := &repositories.ServiceRiskReport{
		DebtCount: map[string]int64{
			"HIGH":   2,
			"MEDIUM": 3,
			"LOW":    5,
		},
		DependentCount: 10,
	}

	handler := CallsHandler{
		repository: mockReportRepository{
			Err:    nil, // No error
			Report: mockReport,
		},
	}

	// Create a request
	req, err := http.NewRequest("GET", "/reports/services/"+validServiceId+"/risk", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.SetPathValue("id", validServiceId)
	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetServiceRiskReport(rw, req)

	// Check the response
	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}

	// Decode the response body
	var report repositories.ServiceRiskReport
	err = json.NewDecoder(rw.Body).Decode(&report)
	if err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if contentType := rw.Header().Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type %s, got %s", "application/json", contentType)
	}

	// Check that the correct report was returned
	if report.DependentCount != mockReport.DependentCount {
		t.Errorf("Expected DependentCount %d, got %d", mockReport.DependentCount, report.DependentCount)
	}

	// Check the DebtCount map
	for key, value := range mockReport.DebtCount {
		if report.DebtCount[key] != value {
			t.Errorf("Expected DebtCount[%s] = %d, got %d", key, value, report.DebtCount[key])
		}
	}
}

func TestGetServiceRiskReportInvalidPathParameter(t *testing.T) {
	// Create a handler with mocked dependencies
	handler := CallsHandler{
		repository: mockReportRepository{
			Err: nil, // No error
		},
	}

	// Create a request with invalid service ID
	req, err := http.NewRequest("GET", "/reports/services/invalid-id/risk", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", "invalid-id")
	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetServiceRiskReport(rw, req)

	// Check the response
	if rw.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rw.Code)
	}
}

func TestGetServiceRiskReportRepositoryError(t *testing.T) {
	// Create a handler with mocked dependencies
	validServiceId := "123e4567-e89b-12d3-a456-426614174000" // Valid GUID
	handler := CallsHandler{
		repository: mockReportRepository{
			Err: errors.New("repository error"), // Simulate a repository error
		},
	}

	// Create a request
	req, err := http.NewRequest("GET", "/reports/services/"+validServiceId+"/risk", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", validServiceId)
	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetServiceRiskReport(rw, req)

	// Check the response
	if rw.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rw.Code)
	}
}

func TestGetServiceRiskReportHTTPError(t *testing.T) {
	// Create a handler with mocked dependencies
	validServiceId := "123e4567-e89b-12d3-a456-426614174000" // Valid GUID
	handler := CallsHandler{
		repository: mockReportRepository{
			Err: &customerrors.HTTPError{
				Status: http.StatusNotFound,
				Msg:    "Service not found",
			}, // Simulate an HTTP error
		},
	}

	// Create a request
	req, err := http.NewRequest("GET", "/reports/services/"+validServiceId+"/risk", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", validServiceId)
	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetServiceRiskReport(rw, req)

	// Check the response
	if rw.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rw.Code)
	}
}
