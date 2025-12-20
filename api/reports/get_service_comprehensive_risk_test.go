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

func TestGetComprehensiveRiskReport_Success(t *testing.T) {
	// Arrange
	validServiceID := "123e4567-e89b-12d3-a456-426614174000"
	health := &repositories.ServiceRiskReport{
		DebtCount:      map[string]int64{"open": 3, "resolved": 1},
		DependentCount: 7,
	}
	change := &repositories.ServiceChangeRisk{Risk: "medium", Score: 42}

	h := CallsHandler{repository: mockReportRepository{Report: health, Change: change}}

	req, err := http.NewRequest("GET", "/reports/services/"+validServiceID+"/risk", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.SetPathValue("id", validServiceID)

	rw := httptest.NewRecorder()

	// Act
	h.GetComprehensiveRiskReport(rw, req)

	// Assert
	if rw.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rw.Code)
	}
	if got := rw.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected content-type application/json, got %q", got)
	}

	var resp repositories.ComprehensiveServiceRisk
	if err := json.NewDecoder(rw.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.HealthRisk == nil || resp.ChangeRisk == nil {
		t.Fatalf("expected both HealthRisk and ChangeRisk to be present, got: %#v", resp)
	}
	if resp.HealthRisk.DependentCount != health.DependentCount {
		t.Errorf("DependentCount = %d, want %d", resp.HealthRisk.DependentCount, health.DependentCount)
	}
	if resp.ChangeRisk.Risk != change.Risk || resp.ChangeRisk.Score != change.Score {
		t.Errorf("ChangeRisk = %#v, want %#v", resp.ChangeRisk, change)
	}
}

func TestGetComprehensiveRiskReport_InvalidServiceID(t *testing.T) {
	h := CallsHandler{repository: mockReportRepository{}}

	req, err := http.NewRequest("GET", "/reports/services/invalid-id/risk", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	// Intentionally do not set a valid GUID path value

	rw := httptest.NewRecorder()
	h.GetComprehensiveRiskReport(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rw.Code)
	}
}

func TestGetComprehensiveRiskReport_HTTPError(t *testing.T) {
	validServiceID := "123e4567-e89b-12d3-a456-426614174000"
	h := CallsHandler{repository: mockReportRepository{Err: &customerrors.HTTPError{Status: http.StatusNotFound, Msg: "not found"}}}

	req, err := http.NewRequest("GET", "/reports/services/"+validServiceID+"/risk", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.SetPathValue("id", validServiceID)

	rw := httptest.NewRecorder()
	h.GetComprehensiveRiskReport(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rw.Code)
	}
}

func TestGetComprehensiveRiskReport_GenericError(t *testing.T) {
	validServiceID := "123e4567-e89b-12d3-a456-426614174000"
	h := CallsHandler{repository: mockReportRepository{Err: errors.New("boom")}}

	req, err := http.NewRequest("GET", "/reports/services/"+validServiceID+"/risk", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.SetPathValue("id", validServiceID)

	rw := httptest.NewRecorder()
	h.GetComprehensiveRiskReport(rw, req)

	if rw.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rw.Code)
	}
}
