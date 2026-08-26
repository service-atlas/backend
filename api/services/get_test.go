package services

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"service-atlas/repositories"
	"strconv"
	"strings"
	"testing"
	"time"
	"uuid"
)

// Tests for GetAllServices

func TestGetAllSuccess(t *testing.T) {
	handler := ServiceCallsHandler{
		Repository: mockServiceRepository{
			Data: func() []map[string]any {
				var m []map[string]any
				for i := 0; i < 10; i++ {
					m = append(m, map[string]any{
						"id":          strconv.Itoa(i),
						"name":        "service" + strconv.Itoa(i),
						"description": "test desc",
						"type":        "service",
						"createdDate": time.Now(),
					})
				}
				return m
			},
			Err: nil,
		},
	}
	req, err := http.NewRequest("GET", "/services?page=1&pageSize=5", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()
	handler.GetAllServices(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}

	// Decode the response body
	var returnedServices []repositories.Service
	if err := json.NewDecoder(rw.Body).Decode(&returnedServices); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Verify that only 5 items were returned
	expectedCount := 5
	if len(returnedServices) != expectedCount {
		t.Errorf("Expected %d returnedServices, got %d", expectedCount, len(returnedServices))
	}
}

func TestGetAllBadRequest(t *testing.T) {
	handler := ServiceCallsHandler{
		Repository: mockServiceRepository{
			Data: func() []map[string]any {
				return []map[string]any{}
			},
			Err: nil,
		},
	}

	// Missing page parameter should result in 400 Bad Request
	req, err := http.NewRequest("GET", "/services", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()
	handler.GetAllServices(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rw.Code)
	}
}

func TestGetAllInternalServerError(t *testing.T) {
	handler := ServiceCallsHandler{
		Repository: mockServiceRepository{
			Data: func() []map[string]any {
				return []map[string]any{}
			},
			Err: errors.New("database error"),
		},
	}

	req, err := http.NewRequest("GET", "/services?page=1&pageSize=5", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()
	handler.GetAllServices(rw, req)

	if rw.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rw.Code)
	}
}

func TestGetAllWithZeroPageSize(t *testing.T) {
	handler := ServiceCallsHandler{

		Repository: mockServiceRepository{
			Data: func() []map[string]any {
				return []map[string]any{}
			},
			Err: nil,
		},
	}

	// pageSize=0 should result in 400 Bad Request
	req, err := http.NewRequest("GET", "/services?page=1&pageSize=0", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()
	handler.GetAllServices(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rw.Code)
	}
}

func TestGetAllWithLargePageSize(t *testing.T) {
	handler := ServiceCallsHandler{

		Repository: mockServiceRepository{
			Data: func() []map[string]any {
				return []map[string]any{}
			},
			Err: nil,
		},
	}

	// pageSize=101 should result in 400 Bad Request
	req, err := http.NewRequest("GET", "/services?page=1&pageSize=101", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()
	handler.GetAllServices(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rw.Code)
	}
}

func TestGetAllWithNegativePage(t *testing.T) {
	handler := ServiceCallsHandler{

		Repository: mockServiceRepository{
			Data: func() []map[string]any {
				return []map[string]any{}
			},
			Err: nil,
		},
	}

	// Negative page number should result in 400 Bad Request
	req, err := http.NewRequest("GET", "/services?page=-1&pageSize=5", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()
	handler.GetAllServices(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rw.Code)
	}
}

func TestGetAllWithNegativePageSize(t *testing.T) {
	handler := ServiceCallsHandler{

		Repository: mockServiceRepository{
			Data: func() []map[string]any {
				return []map[string]any{}
			},
			Err: nil,
		},
	}

	// Negative pageSize should result in 400 Bad Request
	req, err := http.NewRequest("GET", "/services?page=1&pageSize=-5", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()
	handler.GetAllServices(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rw.Code)
	}
}

func TestGetAllWithNonNumericValues(t *testing.T) {
	handler := ServiceCallsHandler{

		Repository: mockServiceRepository{
			Data: func() []map[string]any {
				return []map[string]any{}
			},
			Err: nil,
		},
	}

	// Non-numeric page value should result in 400 Bad Request
	req, err := http.NewRequest("GET", "/services?page=abc&pageSize=5", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()
	handler.GetAllServices(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rw.Code)
	}

	// Non-numeric pageSize value should result in default pageSize (10)
	req, err = http.NewRequest("GET", "/services?page=1&pageSize=abc", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rw = httptest.NewRecorder()
	handler.GetAllServices(rw, req)

	// This should succeed with default pageSize
	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}
}

func TestGetAllWithEmptyResultSet(t *testing.T) {
	handler := ServiceCallsHandler{

		Repository: mockServiceRepository{
			Data: func() []map[string]any {
				return []map[string]any{} // Empty data set
			},
			Err: nil,
		},
	}

	req, err := http.NewRequest("GET", "/services?page=1&pageSize=5", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()
	handler.GetAllServices(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}

	// Decode the response body
	var returnedServices []repositories.Service
	if err := json.NewDecoder(rw.Body).Decode(&returnedServices); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Verify that an empty array (not null) is returned
	if returnedServices == nil {
		t.Errorf("Expected empty array, got nil")
	}
	if len(returnedServices) != 0 {
		t.Errorf("Expected 0 services, got %d", len(returnedServices))
	}
}

func TestGetAllWithPageBeyondAvailableData(t *testing.T) {
	handler := ServiceCallsHandler{

		Repository: mockServiceRepository{
			Data: func() []map[string]any {
				var m []map[string]any
				for i := 0; i < 10; i++ {
					m = append(m, map[string]any{
						"id":          strconv.Itoa(i),
						"name":        "service" + strconv.Itoa(i),
						"description": "test desc",
						"type":        "service",
						"createdDate": time.Now(),
					})
				}
				return m
			},
			Err: nil,
		},
	}

	// Request page 100 when only 10 items exist
	req, err := http.NewRequest("GET", "/services?page=100&pageSize=5", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()
	handler.GetAllServices(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}

	// Decode the response body
	var returnedServices []repositories.Service
	if err := json.NewDecoder(rw.Body).Decode(&returnedServices); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Verify that an empty array is returned
	if len(returnedServices) != 0 {
		t.Errorf("Expected 0 services, got %d", len(returnedServices))
	}
}

func TestGetAllWithDefaultPageSize(t *testing.T) {
	handler := ServiceCallsHandler{

		Repository: mockServiceRepository{
			Data: func() []map[string]any {
				var m []map[string]any
				for i := 0; i < 20; i++ {
					m = append(m, map[string]any{
						"id":          strconv.Itoa(i),
						"name":        "service" + strconv.Itoa(i),
						"description": "test desc",
						"type":        "service",
						"createdDate": time.Now(),
					})
				}
				return m
			},
			Err: nil,
		},
	}

	// Omit the pageSize parameter
	req, err := http.NewRequest("GET", "/services?page=1", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()
	handler.GetAllServices(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}

	// Decode the response body
	var returnedServices []repositories.Service
	if err := json.NewDecoder(rw.Body).Decode(&returnedServices); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	// Verify that default pageSize (10) items were returned
	expectedCount := 10
	if len(returnedServices) != expectedCount {
		t.Errorf("Expected %d services (default pageSize), got %d", expectedCount, len(returnedServices))
	}
}

// Tests for GetById

func TestGetByIdSuccess(t *testing.T) {
	id := uuid.New().String()
	handler := ServiceCallsHandler{
		Repository: mockServiceRepository{
			Data: func() []map[string]any {
				var m []map[string]any

				m = append(m, map[string]any{
					"id":          id,
					"name":        "service",
					"description": "test desc",
					"type":        "service",
					"createdDate": time.Now(),
				})

				return m
			},
			Err: nil,
		},
	}

	req, err := http.NewRequest("GET", "/services/"+id, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", id)
	rw := httptest.NewRecorder()
	handler.GetById(rw, req)

	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}

	// Check that response contains expected data
	responseBody := rw.Body.String()
	if !strings.Contains(responseBody, id) {
		t.Errorf("Response does not contain service ID: %s", responseBody)
	}
	if !strings.Contains(responseBody, "service") {
		t.Errorf("Response does not contain service name: %s", responseBody)
	}
	if !strings.Contains(responseBody, "test desc") {
		t.Errorf("Response does not contain service description: %s", responseBody)
	}
}

func TestGetByIdInvalidId(t *testing.T) {
	handler := ServiceCallsHandler{
		Repository: mockServiceRepository{
			Data: func() []map[string]any {
				return []map[string]any{}
			},
			Err: nil,
		},
	}

	req, err := http.NewRequest("GET", "/services/invalid-id", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rw := httptest.NewRecorder()
	handler.GetById(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rw.Code)
	}

	expectedError := "Service id is required"
	if !strings.Contains(rw.Body.String(), expectedError) {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, rw.Body.String())
	}
}

func TestGetByIdRepositoryError(t *testing.T) {
	id := uuid.New().String()
	expectedError := "database connection error"

	handler := ServiceCallsHandler{
		Repository: mockServiceRepository{
			Data: func() []map[string]any {
				return []map[string]any{}
			},
			Err: errors.New(expectedError),
		},
	}

	req, err := http.NewRequest("GET", "/services/"+id, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", id)
	rw := httptest.NewRecorder()
	handler.GetById(rw, req)

	if rw.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rw.Code)
	}

	if !strings.Contains(rw.Body.String(), expectedError) {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, rw.Body.String())
	}
}

func TestGetByIdServiceNotFound(t *testing.T) {
	id := uuid.New().String()
	nonExistentId := uuid.New().String()

	handler := ServiceCallsHandler{
		Repository: mockServiceRepository{
			Data: func() []map[string]any {
				var m []map[string]any

				m = append(m, map[string]any{
					"id":          id, // Different ID than the one we'll request
					"name":        "service",
					"description": "test desc",
					"type":        "service",
					"createdDate": time.Now(),
				})

				return m
			},
			Err: nil,
		},
	}

	req, err := http.NewRequest("GET", "/services/"+nonExistentId, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", nonExistentId)
	rw := httptest.NewRecorder()
	handler.GetById(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rw.Code)
	}

	expectedError := "Service not found"
	if !strings.Contains(rw.Body.String(), expectedError) {
		t.Errorf("Expected error message '%s', got '%s'", expectedError, rw.Body.String())
	}
}
