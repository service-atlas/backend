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

func TestGetServiceChangeRisk_Success(t *testing.T) {
	// Arrange
	expected := &repositories.ServiceChangeRisk{Risk: "medium", Score: 42}
	h := CallsHandler{repository: mockReportRepository{Change: expected}}

	req := httptest.NewRequest(http.MethodGet, "/reports/service/{id}/change-risk", nil)
	req.SetPathValue("id", "3fa85f64-5717-4562-b3fc-2c963f66afa6") // valid UUID

	rw := httptest.NewRecorder()

	// Act
	h.GetServiceChangeRisk(rw, req)

	// Assert
	if rw.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rw.Code)
	}
	var got repositories.ServiceChangeRisk
	if err := json.NewDecoder(rw.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got.Risk != expected.Risk || got.Score != expected.Score {
		t.Fatalf("unexpected body: %+v; want %+v", got, expected)
	}
}

func TestGetServiceChangeRisk_InvalidGuid(t *testing.T) {
	h := CallsHandler{repository: mockReportRepository{}}

	req := httptest.NewRequest(http.MethodGet, "/reports/service/{id}/change-risk", nil)
	req.SetPathValue("id", "not-a-guid")

	rw := httptest.NewRecorder()
	h.GetServiceChangeRisk(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rw.Code)
	}
}

func TestGetServiceChangeRisk_HTTPError(t *testing.T) {
	h := CallsHandler{repository: mockReportRepository{Err: &customerrors.HTTPError{Status: http.StatusNotFound, Msg: "not found"}}}

	req := httptest.NewRequest(http.MethodGet, "/reports/service/{id}/change-risk", nil)
	req.SetPathValue("id", "3fa85f64-5717-4562-b3fc-2c963f66afa6")

	rw := httptest.NewRecorder()
	h.GetServiceChangeRisk(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rw.Code)
	}
}

func TestGetServiceChangeRisk_RepoError(t *testing.T) {
	h := CallsHandler{repository: mockReportRepository{Err: errors.New("repo failure")}}

	req := httptest.NewRequest(http.MethodGet, "/reports/service/{id}/change-risk", nil)
	req.SetPathValue("id", "3fa85f64-5717-4562-b3fc-2c963f66afa6")

	rw := httptest.NewRecorder()
	h.GetServiceChangeRisk(rw, req)

	if rw.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rw.Code)
	}
}
