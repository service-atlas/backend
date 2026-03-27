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

func TestGetServiceTypes_Success(t *testing.T) {
	// Arrange
	expected := []repositories.ServiceType{
		{Type: "Worker", Count: 3},
		{Type: "Api", Count: 2},
		{Type: "Database", Count: 1},
	}

	h := CallsHandler{repository: mockReportRepository{ServiceTypes: expected}}

	req, err := http.NewRequest("GET", "/services/types", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()

	// Act
	h.GetServiceTypes(rw, req)

	// Assert
	if rw.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rw.Code)
	}
	if got := rw.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected content-type application/json, got %q", got)
	}

	var resp []repositories.ServiceType
	if err := json.NewDecoder(rw.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp) != len(expected) {
		t.Fatalf("expected %d items, got %d", len(expected), len(resp))
	}

	for i, v := range expected {
		if resp[i].Type != v.Type || resp[i].Count != v.Count {
			t.Errorf("at index %d: got %+v, want %+v", i, resp[i], v)
		}
	}
}

func TestGetServiceTypes_RepoError(t *testing.T) {
	// Arrange
	h := CallsHandler{repository: mockReportRepository{Err: errors.New("database failure")}}

	req, err := http.NewRequest("GET", "/services/types", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()

	// Act
	h.GetServiceTypes(rw, req)

	// Assert
	if rw.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rw.Code)
	}
}

func TestGetServiceTypes_HTTPError(t *testing.T) {
	// Arrange
	h := CallsHandler{repository: mockReportRepository{Err: &customerrors.HTTPError{Status: http.StatusTeapot, Msg: "short and stout"}}}

	req, err := http.NewRequest("GET", "/services/types", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()

	// Act
	h.GetServiceTypes(rw, req)

	// Assert
	if rw.Code != http.StatusTeapot {
		t.Fatalf("expected status %d, got %d", http.StatusTeapot, rw.Code)
	}
}
