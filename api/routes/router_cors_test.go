package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// helper to create a router with CORS middleware and a simple endpoint
func newTestRouter() *chi.Mux {
	r := chi.NewRouter()
	setupCORS(r)
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	return r
}

func TestSetupCORS_DefaultConfig_AllowsAnyOrigin(t *testing.T) {
	t.Setenv("CORS_CONFIG", "")

	router := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected Access-Control-Allow-Origin '*' with default config, got %q", got)
	}
}

func TestSetupCORS_CustomConfig_AllowsAndBlocksOrigins(t *testing.T) {
	t.Setenv("CORS_CONFIG", `{"AllowedOrigins":["https://allowed.com"],"AllowedMethods":["GET","OPTIONS"]}`)

	router := newTestRouter()

	// Allowed origin
	reqAllowed := httptest.NewRequest(http.MethodGet, "/", nil)
	reqAllowed.Header.Set("Origin", "https://allowed.com")
	rrAllowed := httptest.NewRecorder()
	router.ServeHTTP(rrAllowed, reqAllowed)
	if got := rrAllowed.Header().Get("Access-Control-Allow-Origin"); got != "https://allowed.com" {
		t.Fatalf("expected echo Access-Control-Allow-Origin for allowed origin, got %q", got)
	}

	// Blocked origin
	reqBlocked := httptest.NewRequest(http.MethodGet, "/", nil)
	reqBlocked.Header.Set("Origin", "https://blocked.com")
	rrBlocked := httptest.NewRecorder()
	router.ServeHTTP(rrBlocked, reqBlocked)
	if got := rrBlocked.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no Access-Control-Allow-Origin for blocked origin, got %q", got)
	}
}

func TestSetupCORS_PreflightSetsAllowMethods(t *testing.T) {
	t.Setenv("CORS_CONFIG", `{"AllowedOrigins":["https://allowed.com"],"AllowedMethods":["GET","POST","OPTIONS"]}`)

	router := newTestRouter()

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://allowed.com")
	req.Header.Set("Access-Control-Request-Method", "GET")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Access-Control-Allow-Origin should echo the allowed origin
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://allowed.com" {
		t.Fatalf("expected Access-Control-Allow-Origin to be echoed for preflight, got %q", got)
	}

	// Access-Control-Allow-Methods should include the requested method and be derived from config
	acam := rr.Header().Get("Access-Control-Allow-Methods")
	if acam == "" || !strings.Contains(acam, "GET") {
		t.Fatalf("expected Access-Control-Allow-Methods to include GET, got %q", acam)
	}

	// Status may be 200 or 204 depending on middleware; ensure it's a successful code
	if rr.Code < 200 || rr.Code >= 300 {
		t.Fatalf("expected successful status code for preflight, got %d", rr.Code)
	}
}
