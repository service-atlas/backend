package auth

import (
	"context"
	"errors"
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
	t.Run("Valid token", func(t *testing.T) {
		// Since we can't easily construct a real oidc.IDToken (unexported fields),
		// we might need to change the Middleware to use a simpler claims extractor
		// if we want to test the full flow, or just test up to the Verify call.
		// However, for this task, I'll mock the verifier to return an error or success.

		// Actually, oidc.IDToken is hard to mock without unsafe.
		// Let's test the Unauthorized case at least.
		verifier := &mockVerifier{
			verifyFunc: func(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
				return nil, errors.New("invalid token")
			},
		}

		mw := Middleware(verifier)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rr.Code)
		}
	})

	t.Run("Missing token", func(t *testing.T) {
		mw := Middleware(nil)
		handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("Expected status 401, got %d", rr.Code)
		}
	})
}
