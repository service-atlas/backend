package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
)

func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		want       string
	}{
		{"Valid Bearer", "Bearer mytoken", "mytoken"},
		{"Short string", "Bear", ""},
		{"Empty string", "", ""},
		{"No Bearer prefix", "Token mytoken", ""},
		{"Case sensitive prefix", "bearer mytoken", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractBearerToken(tt.authHeader)
			if got != tt.want {
				t.Errorf("extractBearerToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNameFromContext(t *testing.T) {
	t.Run("Name exists in context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), nameContextKey{}, "John Doe")
		got := NameFromContext(ctx)
		if got != "John Doe" {
			t.Errorf("NameFromContext() = %v, want %v", got, "John Doe")
		}
	})

	t.Run("Name does not exist in context", func(t *testing.T) {
		got := NameFromContext(context.Background())
		if got != "" {
			t.Errorf("NameFromContext() = %v, want empty string", got)
		}
	})
}

type mockVerifier struct {
	verifyFunc func(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
}

func (m *mockVerifier) Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return m.verifyFunc(ctx, rawIDToken)
}

func TestMiddleware(t *testing.T) {
	t.Run("Disabled auth", func(t *testing.T) {
		cfg := &AuthConfig{enabled: false}
		mw := Middleware(cfg)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", rr.Code)
		}
	})

	t.Run("Missing token (Enabled)", func(t *testing.T) {
		// We can't easily test Enabled because it tries to reach the Issuer URL
		// But we can test that it fails to initialize if the Issuer is invalid
		cfg := &AuthConfig{
			enabled: true,
			issuer:  "http://localhost:1", // Use a port that is likely closed to fail fast
		}
		mw := Middleware(cfg)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		// It should fail initialization and return 500
		if rr.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500, got %d", rr.Code)
		}
	})
}
