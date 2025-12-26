package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

// helper to create a router with system calls registered
func newSystemRouter() *chi.Mux {
	r := chi.NewRouter()
	setupSystemCalls(r)
	return r
}

func TestSetupSystemCalls_HelloWorld(t *testing.T) {
	router := newSystemRouter()

	req := httptest.NewRequest(http.MethodGet, "/helloworld", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("/helloworld expected status 200, got %d", rr.Code)
	}
	if got := rr.Body.String(); got != "hello world" {
		t.Fatalf("/helloworld expected body 'hello world', got %q", got)
	}
}

func TestSetupSystemCalls_Time(t *testing.T) {
	router := newSystemRouter()

	req := httptest.NewRequest(http.MethodGet, "/time", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("/time expected status 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if _, err := time.Parse("2006-01-02 15:04:05", body); err != nil {
		t.Fatalf("/time expected time formatted as YYYY-MM-DD HH:MM:SS, got %q (err=%v)", body, err)
	}
}

func TestSetupSystemCalls_Database(t *testing.T) {
	t.Setenv("NEO4J_URL", "bolt://db:7687")

	router := newSystemRouter()

	req := httptest.NewRequest(http.MethodGet, "/database", nil)
	rr := httptest.NewRecorder()

	router.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("/database expected status 200, got %d", rr.Code)
	}

	if got := rr.Body.String(); got != "bolt://db:7687" {
		t.Fatalf("/database expected body 'bolt://db:7687', got %q", got)
	}
}
