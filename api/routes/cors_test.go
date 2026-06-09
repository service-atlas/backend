package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSHeaders(t *testing.T) {
	// Setup driver mock or nil if not used by the parts we test
	// SetupRouter calls setupCORS which uses internal/config
	// We can just test the handler returned by SetupRouter

	handler := SetupRouter(nil)
	server := httptest.NewServer(handler)
	defer server.Close()

	req, _ := http.NewRequest("OPTIONS", server.URL+"/helloworld", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Authorization")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected OK or No Content for OPTIONS, got %d", resp.StatusCode)
	}

	if h := resp.Header.Get("Access-Control-Allow-Origin"); h != "http://example.com" && h != "*" {
		t.Errorf("Expected Allow-Origin, got %s", h)
	}

	if h := resp.Header.Get("Access-Control-Allow-Headers"); h == "" {
		t.Errorf("Expected Allow-Headers to be set")
	}
}
