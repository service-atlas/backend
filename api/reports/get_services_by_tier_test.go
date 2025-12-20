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

func TestGetServicesByTierSuccess(t *testing.T) {
	services := []repositories.Service{
		{Id: "1", Name: "svc-a", ServiceType: "service", Description: "desc-a", Url: "https://a.example.com", Tier: 2},
		{Id: "2", Name: "svc-b", ServiceType: "service", Description: "desc-b", Url: "https://b.example.com", Tier: 2},
	}

	h := CallsHandler{repository: mockReportRepository{Services: services}}

	req, err := http.NewRequest("GET", "/reports/services/by-tier?tier=2", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()
	h.GetServicesByTier(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rw.Code)
	}

	var got []repositories.Service
	if err := json.NewDecoder(rw.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if len(got) != len(services) {
		t.Fatalf("expected %d services, got %d", len(services), len(got))
	}
	for i := range services {
		if got[i].Name != services[i].Name {
			t.Errorf("service[%d].Name = %q, want %q", i, got[i].Name, services[i].Name)
		}
		if got[i].Tier != 2 {
			t.Errorf("service[%d].Tier = %d, want %d", i, got[i].Tier, 2)
		}
	}
}

func TestGetServicesByTierInvalidTierValue(t *testing.T) {
	h := CallsHandler{repository: mockReportRepository{}}

	req, err := http.NewRequest("GET", "/reports/services/by-tier?tier=NaN", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()
	h.GetServicesByTier(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rw.Code)
	}
}

func TestGetServicesByTierOutOfRange(t *testing.T) {
	h := CallsHandler{repository: mockReportRepository{}}

	// tier 0 is out of allowed range [1..4]
	req, err := http.NewRequest("GET", "/reports/services/by-tier?tier=0", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()
	h.GetServicesByTier(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rw.Code)
	}
}

func TestGetServicesByTierRepositoryError(t *testing.T) {
	h := CallsHandler{repository: mockReportRepository{Err: errors.New("repo error")}}

	req, err := http.NewRequest("GET", "/reports/services/by-tier?tier=2", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()
	h.GetServicesByTier(rw, req)

	if rw.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rw.Code)
	}
}

func TestGetServicesByTierHTTPError(t *testing.T) {
	h := CallsHandler{repository: mockReportRepository{Err: &customerrors.HTTPError{Status: http.StatusNotFound, Msg: "not found"}}}

	req, err := http.NewRequest("GET", "/reports/services/by-tier?tier=3", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()
	h.GetServicesByTier(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rw.Code)
	}
}
