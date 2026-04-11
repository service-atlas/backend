package dependencies

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"service-atlas/internal/customerrors"
	"service-atlas/repositories"
	"testing"
)

func TestGetByIdSuccess(t *testing.T) {
	// Create mock dependencies
	mockDeps := []map[string]any{
		{
			"id":      "dependency-id-1",
			"name":    "Dependency 1",
			"version": "1.0.0",
		},
		{
			"id":      "dependency-id-2",
			"name":    "Dependency 2",
			"version": "2.0.0",
		},
	}

	// Create a handler with mocked dependencies
	handler := ServiceCallsHandler{
		Repository: mockDependencyRepository{
			Data: func() []map[string]any {
				return mockDeps
			},
			Err: nil, // No error
		},
	}

	// Create a request
	req, err := http.NewRequest("GET", "/services/be00abbc-42c6-47aa-a45a-e4e02cb6363f/dependencies", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", "be00abbc-42c6-47aa-a45a-e4e02cb6363f")
	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetDependencies(rw, req)

	// Check the response
	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}

	// Check the content type
	contentType := rw.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type %s, got %s", "application/json", contentType)
	}

	// Decode the response
	var dependencies []*repositories.Dependency
	err = json.NewDecoder(rw.Body).Decode(&dependencies)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check the number of dependencies
	if len(dependencies) != len(mockDeps) {
		t.Errorf("Expected %d dependencies, got %d", len(mockDeps), len(dependencies))
	}

	// Check the dependencies
	for i, dep := range dependencies {
		if dep.Id != mockDeps[i]["id"] {
			t.Errorf("Expected dependency ID %s, got %s", mockDeps[i]["id"], dep.Id)
		}
		if dep.Name != mockDeps[i]["name"] {
			t.Errorf("Expected dependency name %s, got %s", mockDeps[i]["name"], dep.Name)
		}
		if dep.Version != mockDeps[i]["version"] {
			t.Errorf("Expected dependency version %s, got %s", mockDeps[i]["version"], dep.Version)
		}
	}
}

func TestGetDependenciesByInteractionTypeSuccess(t *testing.T) {
	// Create mock dependencies with different interaction types
	mockDeps := []map[string]any{
		{
			"id":               "dependency-id-1",
			"name":             "Dependency 1",
			"version":          "1.0.0",
			"interaction_type": "security",
		},
		{
			"id":               "dependency-id-2",
			"name":             "Dependency 2",
			"version":          "2.0.0",
			"interaction_type": "data",
		},
		{
			"id":               "dependency-id-3",
			"name":             "Dependency 3",
			"version":          "1.1.0",
			"interaction_type": "security",
		},
	}

	// Create a handler with mocked dependencies
	handler := ServiceCallsHandler{
		Repository: mockDependencyRepository{
			Data: func() []map[string]any {
				return mockDeps
			},
			Err: nil, // No error
		},
	}

	// Create a request with interaction_type filter
	req, err := http.NewRequest("GET", "/services/be00abbc-42c6-47aa-a45a-e4e02cb6363f/dependencies?interaction_type=security", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", "be00abbc-42c6-47aa-a45a-e4e02cb6363f")

	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetDependencies(rw, req)

	// Check the response
	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}

	// Check the content type
	contentType := rw.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type %s, got %s", "application/json", contentType)
	}

	// Decode the response
	var dependencies []*repositories.Dependency
	err = json.NewDecoder(rw.Body).Decode(&dependencies)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check that only dependencies with interaction_type 'security' are returned
	expectedCount := 2 // There are 2 dependencies with interaction_type 'security'
	if len(dependencies) != expectedCount {
		t.Errorf("Expected %d dependencies, got %d", expectedCount, len(dependencies))
	}

	// Check that all returned dependencies have the correct interaction_type
	for _, dep := range dependencies {
		if dep.InteractionType != "security" {
			t.Errorf("Expected interaction_type %s, got %s", "security", dep.InteractionType)
		}
	}
}

func TestGetDependenciesByInteractionTypeInvalid(t *testing.T) {
	// Create a handler
	handler := ServiceCallsHandler{
		Repository: mockDependencyRepository{
			Data: func() []map[string]any {
				return []map[string]any{}
			},
			Err: nil,
		},
	}

	// Create a request with an invalid interaction_type
	req, err := http.NewRequest("GET", "/services/be00abbc-42c6-47aa-a45a-e4e02cb6363f/dependencies?interaction_type=invalid_type", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", "be00abbc-42c6-47aa-a45a-e4e02cb6363f")

	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetDependencies(rw, req)

	// Check the response
	if rw.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rw.Code)
	}
}

func TestGetDependenciesWithEmptyInteractionType(t *testing.T) {
	// Create mock dependencies
	mockDeps := []map[string]any{
		{
			"id":               "dependency-id-1",
			"name":             "Dependency 1",
			"interaction_type": "security",
		},
	}

	// Create a handler
	handler := ServiceCallsHandler{
		Repository: mockDependencyRepository{
			Data: func() []map[string]any {
				return mockDeps
			},
			Err: nil,
		},
	}

	// Create a request with an empty interaction_type
	req, err := http.NewRequest("GET", "/services/be00abbc-42c6-47aa-a45a-e4e02cb6363f/dependencies?interaction_type=", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", "be00abbc-42c6-47aa-a45a-e4e02cb6363f")

	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetDependencies(rw, req)

	// Check the response
	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}

	// Decode the response
	var dependencies []*repositories.Dependency
	err = json.NewDecoder(rw.Body).Decode(&dependencies)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(dependencies) != 1 {
		t.Errorf("Expected 1 dependency, got %d", len(dependencies))
	}
}

func TestGetDependenciesNoDataReturnsEmptyArray(t *testing.T) {
	// Create a handler with no dependencies
	handler := ServiceCallsHandler{
		Repository: mockDependencyRepository{
			Data: func() []map[string]any {
				return []map[string]any{}
			},
			Err: nil, // No error
		},
	}

	// Create a request
	req, err := http.NewRequest("GET", "/services/be00abbc-42c6-47aa-a45a-e4e02cb6363f/dependencies", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", "be00abbc-42c6-47aa-a45a-e4e02cb6363f")
	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetDependencies(rw, req)

	// Check the response
	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}

	// Check the body for an empty array, not null
	body := rw.Body.String()
	if body != "[]\n" {
		t.Errorf("Expected empty array [], got %q", body)
	}
}

func TestGetDependenciesByInteractionTypeNoDataReturnsEmptyArray(t *testing.T) {
	// Create a handler with dependencies of other types, but not 'security'
	mockDeps := []map[string]any{
		{
			"id":               "dependency-id-1",
			"name":             "Dependency 1",
			"interaction_type": "data",
		},
	}

	handler := ServiceCallsHandler{
		Repository: mockDependencyRepository{
			Data: func() []map[string]any {
				return mockDeps
			},
			Err: nil, // No error
		},
	}

	// Create a request with interaction_type=security (which won't have results)
	req, err := http.NewRequest("GET", "/services/be00abbc-42c6-47aa-a45a-e4e02cb6363f/dependencies?interaction_type=security", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", "be00abbc-42c6-47aa-a45a-e4e02cb6363f")
	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetDependencies(rw, req)

	// Check the response
	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}

	// Check the body for an empty array, not null
	body := rw.Body.String()
	if body != "[]\n" {
		t.Errorf("Expected empty array [], got %q", body)
	}
}

func TestGetDependentsNoDataReturnsEmptyArray(t *testing.T) {
	// Create a handler with no dependents
	handler := ServiceCallsHandler{
		Repository: mockDependencyRepository{
			Data: func() []map[string]any {
				return []map[string]any{}
			},
			Err: nil, // No error
		},
	}

	// Create a request
	req, err := http.NewRequest("GET", "/services/be00abbc-42c6-47aa-a45a-e4e02cb6363f/dependents", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", "be00abbc-42c6-47aa-a45a-e4e02cb6363f")
	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetDependents(rw, req)

	// Check the response
	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}

	// Check the body for an empty array, not null
	body := rw.Body.String()
	if body != "[]\n" {
		t.Errorf("Expected empty array [], got %q", body)
	}
}

func TestGetByIdInvalidPath(t *testing.T) {
	// Create a handler with mocked dependencies
	handler := ServiceCallsHandler{
		Repository: mockDependencyRepository{
			Data: func() []map[string]any {
				return []map[string]any{} // Empty data, not used in this test
			},
			Err: nil, // No error
		},
	}

	// Create a request
	req, err := http.NewRequest("GET", "/services/invalid-id/dependencies", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetDependencies(rw, req)

	// Check the response
	if rw.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rw.Code)
	}
}

func TestGetByIdRepositoryError(t *testing.T) {
	// Create a handler with mocked dependencies
	handler := ServiceCallsHandler{
		Repository: mockDependencyRepository{
			Data: func() []map[string]any {
				return []map[string]any{} // Empty data, not used in this test
			},
			Err: errors.New("repository error"), // Simulate a repository error
		},
	}

	// Create a request
	req, err := http.NewRequest("GET", "/services/be00abbc-42c6-47aa-a45a-e4e02cb6363f/dependencies", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", "be00abbc-42c6-47aa-a45a-e4e02cb6363f")
	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetDependencies(rw, req)

	// Check the response
	if rw.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rw.Code)
	}
}

func TestGetByIdHTTPError(t *testing.T) {
	// Create a handler with mocked dependencies
	handler := ServiceCallsHandler{
		Repository: mockDependencyRepository{
			Data: func() []map[string]any {
				return []map[string]any{} // Empty data, not used in this test
			},
			Err: &customerrors.HTTPError{
				Status: http.StatusNotFound,
				Msg:    "Service not found",
			}, // Simulate an HTTP error
		},
	}

	// Create a request
	req, err := http.NewRequest("GET", "/services/be00abbc-42c6-47aa-a45a-e4e02cb6363f/dependencies", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", "be00abbc-42c6-47aa-a45a-e4e02cb6363f")
	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetDependencies(rw, req)

	// Check the response
	if rw.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rw.Code)
	}
}

func TestGetDependentsByIdSuccess(t *testing.T) {
	// Create mock dependents
	mockDeps := []map[string]any{
		{
			"id":      "dependent-id-1",
			"name":    "Dependent 1",
			"version": "1.0.0",
		},
		{
			"id":      "dependent-id-2",
			"name":    "Dependent 2",
			"version": "2.0.0",
		},
	}

	// Create a handler with mocked dependents
	handler := ServiceCallsHandler{

		Repository: mockDependencyRepository{
			Data: func() []map[string]any {
				return mockDeps
			},
			Err: nil, // No error
		},
	}

	// Create a request
	req, err := http.NewRequest("GET", "/services/be00abbc-42c6-47aa-a45a-e4e02cb6363f/dependents", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", "be00abbc-42c6-47aa-a45a-e4e02cb6363f")
	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetDependents(rw, req)

	// Check the response
	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}

	// Check the content type
	contentType := rw.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type %s, got %s", "application/json", contentType)
	}

	// Decode the response
	var dependents []*repositories.Dependency
	err = json.NewDecoder(rw.Body).Decode(&dependents)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check the number of dependents
	if len(dependents) != len(mockDeps) {
		t.Errorf("Expected %d dependents, got %d", len(mockDeps), len(dependents))
	}

	// Check the dependents
	for i, dep := range dependents {
		if dep.Id != mockDeps[i]["id"] {
			t.Errorf("Expected dependent ID %s, got %s", mockDeps[i]["id"], dep.Id)
		}
		if dep.Name != mockDeps[i]["name"] {
			t.Errorf("Expected dependent name %s, got %s", mockDeps[i]["name"], dep.Name)
		}
		if dep.Version != mockDeps[i]["version"] {
			t.Errorf("Expected dependent version %s, got %s", mockDeps[i]["version"], dep.Version)
		}
	}
}

func TestGetDependentsByIdInvalidPath(t *testing.T) {
	// Create a handler with mocked dependents
	handler := ServiceCallsHandler{

		Repository: mockDependencyRepository{
			Data: func() []map[string]any {
				return []map[string]any{} // Empty data, not used in this test
			},
			Err: nil, // No error
		},
	}

	// Create a request
	req, err := http.NewRequest("GET", "/services/invalid-id/dependents", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetDependents(rw, req)

	// Check the response
	if rw.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, rw.Code)
	}
}

func TestGetDependentsByIdRepositoryError(t *testing.T) {
	// Create a handler with mocked dependents
	handler := ServiceCallsHandler{

		Repository: mockDependencyRepository{
			Data: func() []map[string]any {
				return []map[string]any{} // Empty data, not used in this test
			},
			Err: errors.New("repository error"), // Simulate a repository error
		},
	}

	// Create a request
	req, err := http.NewRequest("GET", "/services/be00abbc-42c6-47aa-a45a-e4e02cb6363f/dependents", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", "be00abbc-42c6-47aa-a45a-e4e02cb6363f")
	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetDependents(rw, req)

	// Check the response
	if rw.Code != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, rw.Code)
	}
}

func TestGetDependentsByIdHTTPError(t *testing.T) {
	// Create a handler with mocked dependents
	handler := ServiceCallsHandler{
		Repository: mockDependencyRepository{
			Data: func() []map[string]any {
				return []map[string]any{} // Empty data, not used in this test
			},
			Err: &customerrors.HTTPError{
				Status: http.StatusNotFound,
				Msg:    "Service not found",
			}, // Simulate an HTTP error
		},
	}

	// Create a request
	req, err := http.NewRequest("GET", "/services/be00abbc-42c6-47aa-a45a-e4e02cb6363f/dependents", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", "be00abbc-42c6-47aa-a45a-e4e02cb6363f")
	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetDependents(rw, req)

	// Check the response
	if rw.Code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, rw.Code)
	}
}

func TestGetDependentsByIdWithVersionFilter(t *testing.T) {
	// Create mock dependents with different versions
	mockDeps := []map[string]any{
		{
			"id":      "dependent-id-1",
			"name":    "Dependent 1",
			"version": "1.0.0",
		},
		{
			"id":      "dependent-id-2",
			"name":    "Dependent 2",
			"version": "2.0.0",
		},
		{
			"id":      "dependent-id-3",
			"name":    "Dependent 3",
			"version": "1.0.0", // Same version as dependent-id-1
		},
	}

	// Create a handler with mocked dependents
	handler := ServiceCallsHandler{
		Repository: mockDependencyRepository{
			Data: func() []map[string]any {
				return mockDeps
			},
			Err: nil, // No error
		},
	}

	// Create a request with version filter
	req, err := http.NewRequest("GET", "/services/be00abbc-42c6-47aa-a45a-e4e02cb6363f/dependents?version=1.0.0", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", "be00abbc-42c6-47aa-a45a-e4e02cb6363f")
	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetDependents(rw, req)

	// Check the response
	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}

	// Check the content type
	contentType := rw.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type %s, got %s", "application/json", contentType)
	}

	// Decode the response
	var dependents []*repositories.Dependency
	err = json.NewDecoder(rw.Body).Decode(&dependents)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check that only dependencies with version 1.0.0 are returned
	expectedCount := 2 // There are 2 dependencies with version 1.0.0
	if len(dependents) != expectedCount {
		t.Errorf("Expected %d dependents, got %d", expectedCount, len(dependents))
	}

	// Check that all returned dependencies have the correct version
	for _, dep := range dependents {
		if dep.Version != "1.0.0" {
			t.Errorf("Expected dependent version %s, got %s", "1.0.0", dep.Version)
		}
	}
}

func TestGetDependentsByIdWithNonMatchingVersionFilter(t *testing.T) {
	// Create mock dependents with different versions
	mockDeps := []map[string]any{
		{
			"id":      "dependent-id-1",
			"name":    "Dependent 1",
			"version": "1.0.0",
		},
		{
			"id":      "dependent-id-2",
			"name":    "Dependent 2",
			"version": "2.0.0",
		},
	}

	// Create a handler with mocked dependents
	handler := ServiceCallsHandler{
		Repository: mockDependencyRepository{
			Data: func() []map[string]any {
				return mockDeps
			},
			Err: nil, // No error
		},
	}

	// Create a request with a version filter that doesn't match any dependency
	req, err := http.NewRequest("GET", "/services/be00abbc-42c6-47aa-a45a-e4e02cb6363f/dependents?version=3.0.0", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", "be00abbc-42c6-47aa-a45a-e4e02cb6363f")
	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetDependents(rw, req)

	// Check the response
	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}

	// Check the content type
	contentType := rw.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type %s, got %s", "application/json", contentType)
	}

	// Decode the response
	var dependents []*repositories.Dependency
	err = json.NewDecoder(rw.Body).Decode(&dependents)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check that no dependencies are returned
	if len(dependents) != 0 {
		t.Errorf("Expected 0 dependents, got %d", len(dependents))
	}
}

func TestGetDependentsByIdWithNoVersionFilter(t *testing.T) {
	// Create mock dependents with different versions
	mockDeps := []map[string]any{
		{
			"id":      "dependent-id-1",
			"name":    "Dependent 1",
			"version": "1.0.0",
		},
		{
			"id":      "dependent-id-2",
			"name":    "Dependent 2",
			"version": "2.0.0",
		},
	}

	// Create a handler with mocked dependents
	handler := ServiceCallsHandler{
		Repository: mockDependencyRepository{
			Data: func() []map[string]any {
				return mockDeps
			},
			Err: nil, // No error
		},
	}

	// Create a request without a version filter
	req, err := http.NewRequest("GET", "/services/be00abbc-42c6-47aa-a45a-e4e02cb6363f/dependents", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.SetPathValue("id", "be00abbc-42c6-47aa-a45a-e4e02cb6363f")
	// Create a response recorder
	rw := httptest.NewRecorder()

	// Call the handler
	handler.GetDependents(rw, req)

	// Check the response
	if rw.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rw.Code)
	}

	// Check the content type
	contentType := rw.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type %s, got %s", "application/json", contentType)
	}

	// Decode the response
	var dependents []*repositories.Dependency
	err = json.NewDecoder(rw.Body).Decode(&dependents)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check that all dependencies are returned
	if len(dependents) != len(mockDeps) {
		t.Errorf("Expected %d dependents, got %d", len(mockDeps), len(dependents))
	}

	// Check the dependents
	for i, dep := range dependents {
		if dep.Id != mockDeps[i]["id"] {
			t.Errorf("Expected dependent ID %s, got %s", mockDeps[i]["id"], dep.Id)
		}
		if dep.Name != mockDeps[i]["name"] {
			t.Errorf("Expected dependent name %s, got %s", mockDeps[i]["name"], dep.Name)
		}
		if dep.Version != mockDeps[i]["version"] {
			t.Errorf("Expected dependent version %s, got %s", mockDeps[i]["version"], dep.Version)
		}
	}
}
